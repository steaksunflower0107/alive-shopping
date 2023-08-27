package controller

import (
	"alivePlatform/order_service/config"
	"alivePlatform/order_service/dao/mq"
	"alivePlatform/order_service/dao/mysql"
	"alivePlatform/order_service/model"
	"alivePlatform/order_service/proto"
	"alivePlatform/order_service/rpc"
	"alivePlatform/order_service/snowflake"

	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// OrderEntity 自定义结构体，实现了两个方法
// 发送事务消息的时候 RocketMQ 会自动根据情况调用那两个方法
type OrderEntity struct {
	OrderId int64
	Param   *proto.OrderReq
	err     error
}

// ExecuteLocalTransaction 当发送prepare(half) message 成功后, 本地的事务方就会被执行
func (o *OrderEntity) ExecuteLocalTransaction(*primitive.Message) primitive.LocalTransactionState {
	fmt.Println("in ExecuteLocalTransaction...")
	if o.Param == nil {
		zap.L().Error("ExecuteLocalTransaction param is nil")
		o.err = status.Error(codes.Internal, "invalid OrderEntity")
		return primitive.CommitMessageState
	}

	param := o.Param
	ctx := context.Background()
	// 查询商品金额（营销）--> RPC连接 goods_service
	goodsDatail, err := rpc.GoodsCli.GetGoodsDetail(ctx, &proto.GetGoodsDetailReq{
		GoodsId: param.GoodsId,
		UserId:  param.UserId,
	})

	if err != nil {
		zap.L().Error("GoodsCli.GetGoodsDetail failed", zap.Error(err))
		// 库存未扣减
		o.err = status.Error(codes.Internal, err.Error())
		return primitive.RollbackMessageState
	}

	payAmountStr := goodsDatail.Price
	payAmount, _ := strconv.ParseInt(payAmountStr, 10, 64)

	// 2. 库存校验及扣减  --> RPC连接 stock_service
	_, err = rpc.RepertoryCli.ReduceRepertory(ctx, &proto.GoodsRepertoryInfo{
		OrderId: o.OrderId,
		GoodsId: o.Param.GoodsId,
		Num:     o.Param.Num,
	})

	if err != nil {
		// 库存扣减失败
		zap.L().Error("StockCli.ReduceStock failed", zap.Error(err))
		o.err = status.Error(codes.Internal, "ReduceStock failed")
		return primitive.RollbackMessageState
	}

	// 扣减库存成功
	// 从这里开始如果本地事务执行失败就需要回滚库存

	// 生成订单表
	orderData := model.Order{
		OrderId:        o.OrderId,
		UserId:         param.UserId,
		PayAmount:      payAmount,
		ReceiveAddress: param.Address,
		ReceiveName:    param.Name,
		ReceivePhone:   param.Phone,
		Status:         100, // 待支付
	}

	orderDetail := model.OrderDetail{
		OrderId: o.OrderId,
		UserId:  param.UserId,
		GoodsId: param.GoodsId,
		Num:     param.Num,
	}

	// 在本地事务创建订单和订单详情记录
	err = mysql.CreateOrderWithTransaction(ctx, &orderData, &orderDetail)
	if err != nil {
		// 本地事务执行失败了，上一步已经库存扣减成功
		// 就需要将库存回滚的消息投递出去，小游根据消息进行库存回滚
		zap.L().Error("CreateOrderWithTransation failed", zap.Error(err))
		return primitive.CommitMessageState // 将之前发送的hal-message commit
	}

	// 发送延迟消息
	// 1s 5s 10s 30s 1m 2m 3m 4m 5m 6m 7m 8m 9m 10m 20m 30m 1h 2h
	data := model.OrderGoodsStockInfo{
		OrderId: o.OrderId,
		GoodsId: param.GoodsId,
		Num:     param.Num,
	}

	b, _ := json.Marshal(data)
	msg := primitive.NewMessage("platform_pay_timeout", b)
	msg.WithDelayTimeLevel(3)
	_, err = mq.Producer.SendSync(context.Background(), msg)
	if err != nil {
		// 发送延时消息失败
		zap.L().Error("send delay msg failed", zap.Error(err))
		return primitive.CommitMessageState
	}

	// 本地事务执行成功
	// 将之前的half-message rollback， 丢弃掉
	return primitive.RollbackMessageState
}

// CheckLocalTransaction 当 prepare(half) message 没有响应时
func (o *OrderEntity) CheckLocalTransaction(*primitive.MessageExt) primitive.LocalTransactionState {
	// 检查本地状态是否创建成功订单
	_, err := mysql.QueryOrder(context.Background(), o.OrderId)
	// 需要再查询订单详情表
	if err == gorm.ErrRecordNotFound {
		// 没查询到说明订单创建失败，需要回滚库存
		return primitive.CommitMessageState
	}
	return primitive.RollbackMessageState
}

func Create(ctx context.Context, param *proto.OrderReq) error {
	// 生成订单号
	orderId := snowflake.GenID()

	orderEntity := &OrderEntity{
		OrderId: orderId,
		Param:   param,
	}

	p, err := rocketmq.NewTransactionProducer(
		orderEntity,
		producer.WithNsResolver(primitive.NewPassthroughResolver([]string{"127.0.0.1:9876"})),
		//producer.WithNsResolver(primitive.NewPassthroughResolver(endPoint)),
		producer.WithRetry(2),
		producer.WithGroupName("order_srv_1"), // 生产者组
	)
	if err != nil {
		zap.L().Error("NewTransactionProducer failed", zap.Error(err))
		return status.Error(codes.Internal, "NewTransactionProducer failed")
	}
	p.Start()

	// 封装消息 orderId GoodsId num
	data := model.OrderGoodsStockInfo{
		OrderId: orderId,
		GoodsId: param.GoodsId,
		Num:     param.Num,
	}
	b, _ := json.Marshal(data)
	msg := &primitive.Message{
		Topic: config.Conf.RocketMqConfig.Topic.RepertoryRollback, // xx_stock_rollback
		Body:  b,
	}

	// 发送事务消息
	res, err := p.SendMessageInTransaction(context.Background(), msg)
	if err != nil {
		zap.L().Error("SendMessageInTransaction failed", zap.Error(err))
		return status.Error(codes.Internal, "create order failed")
	}
	zap.L().Info("p.SendMessageInTransaction success", zap.Any("res", res))

	// 如果回滚库存的消息被投递出去（commit）说明本地事务执行失败，也就是创建订单失败
	if res.State == primitive.CommitMessageState {
		return status.Error(codes.Internal, "create order failed")
	}

	// 其他内部错误
	if orderEntity.err != nil {
		return orderEntity.err
	}
	return nil
}
