package biz

import (
	"alivePlatform/repertory_service/dao/mysql"
	"alivePlatform/repertory_service/errno"
	"alivePlatform/repertory_service/proto"

	"context"
)

func GetRepertoryByGoodsId(ctx context.Context, goodsId int64) (*proto.GoodsRepertoryInfo, error) {
	record, err := mysql.GetRepertoryByGoodsId(ctx, goodsId)
	if err != nil {
		return nil, err
	}

	resp := &proto.GoodsRepertoryInfo{
		GoodsId: record.GoodsId,
		Num:     record.Num,
	}

	return resp, nil
}

func ReduceRepertoryByGoodsId(ctx context.Context, goodsId, num int64) (*proto.GoodsRepertoryInfo, error) {
	data, err := mysql.ReduceRepertoryByGoodsId(ctx, goodsId, num)
	if err != nil {
		return nil, errno.ErrUnderRepertory
	}

	return &proto.GoodsRepertoryInfo{
		GoodsId: data.GoodsId,
		Num:     data.Num,
	}, nil
}
