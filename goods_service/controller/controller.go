package controller

import (
	"alivePlatform/goods_service/biz"
	"alivePlatform/goods_service/proto"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GoodsSrv struct {
	proto.UnimplementedGoodsServer
}

// GetGoodsByRoom 获取直播间的商品列表
func (g *GoodsSrv) GetGoodsByRoom(ctx context.Context, req *proto.GetGoodsByRoomReq) (*proto.GoodsListResp, error) {
	// 校验传入参数
	if req.GetRoomId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "直播间id小于0")
	}

	// 查询数据
	data, err := biz.GetGoodsByRoomId(ctx, req.RoomId)
	if err != nil {
		return nil, status.Error(codes.Internal, "内部逻辑错误")
	}

	return data, nil
}

// GetGoodsDetail 获取商品详情
func (g *GoodsSrv) GetGoodsDetail(context.Context, *proto.GetGoodsDetailReq) (*proto.GoodsDetail, error) {
	return &proto.GoodsDetail{}, nil
}
