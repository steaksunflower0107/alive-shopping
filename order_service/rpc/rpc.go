package rpc

import (
	"alivePlatform/order_service/config"
	"alivePlatform/order_service/proto"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	GoodsCli     proto.GoodsClient
	RepertoryCli proto.RepertoryClient
)

func InitSrcClient() error {
	if len(config.Conf.GoodsService.Name) == 0 {
		return errors.New("invalid GoodsService.Name")
	}

	if len(config.Conf.RepertoryService.Name) == 0 {
		return errors.New("invalid RepertoryService.Name")
	}

	// 程序启动的时候请求consul获取一个可以用的商品服务地址
	goodsConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s/%s?wait=14s", config.Conf.ConsulConfig.Addr, config.Conf.GoodsService.Name),
		// 指定round_robin策略
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		fmt.Printf("dial goods_srv failed, err:%v\n", err)
		return err
	}

	GoodsCli = proto.NewGoodsClient(goodsConn)

	repertoryConn, err := grpc.Dial(
		// consul服务
		fmt.Sprintf("consul://%s/%s?wait=14s", config.Conf.ConsulConfig.Addr, config.Conf.RepertoryService.Name),
		// 指定round_robin策略
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		fmt.Printf("dial stock_srv failed, err:%v\n", err)
		return err
	}

	RepertoryCli = proto.NewRepertoryClient(repertoryConn)
	return nil
}
