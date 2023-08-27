package controller

import (
	"alivePlatform/order_service/model"
	"alivePlatform/repertory_service/biz"
	"alivePlatform/repertory_service/dao/mysql"
	model2 "alivePlatform/repertory_service/model"
	"alivePlatform/repertory_service/proto"
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RepertorySrv struct {
	proto.UnimplementedRepertoryServer
}

// GetRepertory 获取商品详情
func (c *RepertorySrv) GetRepertory(ctx context.Context, req *proto.GoodsRepertoryInfo) (*proto.GoodsRepertoryInfo, error) {
	if req.GoodsId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "参数错误")
	}

	data, err := biz.GetRepertoryByGoodsId(ctx, req.GoodsId)
	if err != nil {
		zap.L().Error("GetRepertory failed", zap.Int64("goods_id", req.GoodsId), zap.Error(err))
		return nil, status.Error(codes.Internal, "内部错误")
	}

	return data, nil
}

//func (c *RepertorySrv) BatchGetRepertory(ctx context.Context, in *proto.GoodsListRepertory, opts ...grpc.CallOption) (*proto.GoodsListRepertory, error) {
//	out := new(proto.GoodsListRepertory)
//	err := c.cc.Invoke(ctx, proto.Repertory_BatchGetRepertory_FullMethodName, in, out, opts...)
//	if err != nil {
//		return nil, err
//	}
//	return out, nil
//}

// SetRepertory 库存
//func (c *RepertorySrv) SetRepertory(ctx context.Context, req *proto.GetGoodsByRoomReq) (*proto.GoodsListResp, error) {
//	// 校验传入参数
//	if req.GetRoomId() <= 0 {
//		return nil, status.Error(codes.InvalidArgument, "直播间id小于0")
//	}
//
//	// 查询数据
//	data, err := biz.GetGoodsByRoomId(ctx, req.RoomId)
//	if err != nil {
//		return nil, status.Error(codes.Internal, "内部逻辑错误")
//	}
//
//	return data, nil
//}

func (c *RepertorySrv) ReduceRepertory(ctx context.Context, req *proto.GoodsRepertoryInfo) (*proto.GoodsRepertoryInfo, error) {
	fmt.Printf("in ReduceStock... req:%#v\n", req)
	// 参数处理
	if req.GoodsId <= 0 || req.Num <= 0 {
		return nil, status.Error(codes.InvalidArgument, "请求参数有误")
	}

	// 扣减库存
	data, err := biz.ReduceRepertoryByGoodsId(ctx, req.GetGoodsId(), req.GetNum())
	if err != nil {
		return nil, status.Error(codes.Internal, "内部错误")
	}

	return data, nil
}

// RollbackMsghandle 监听rocketmq消息进行库存回滚的处理函数
// 需考虑重复归还的问题（幂等性）
// 添加库存扣减记录表
func RollbackMsghandle(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range msgs {
		var data model.OrderGoodsStockInfo
		err := json.Unmarshal(msgs[i].Body, &data)
		if err != nil {
			zap.L().Error("json.Unmarshal RollbackMsg failed", zap.Error(err))
			continue
		}
		// 将库存回滚
		err = mysql.RollbackRepertoryByMsg(ctx, model2.OrderGoodsRepertoryInfo(data))
		if err != nil {
			return consumer.ConsumeRetryLater, nil
		}
		return consumer.ConsumeSuccess, nil
	}
	return consumer.ConsumeSuccess, nil
}
