package biz

import (
	"alivePlatform/goods_service/dao/mysql"
	"alivePlatform/goods_service/proto"
	"context"
	"encoding/json"
	"fmt"
)

func GetGoodsByRoomId(ctx context.Context, roomId int64) (*proto.GoodsListResp, error) {
	goodsList, err := mysql.GetGoodsByRoomId(ctx, roomId)
	if err != nil {
		return nil, err
	}

	var (
		currGoodsId int64
		idList      = make([]int64, 0, len(goodsList))
	)

	for _, good := range goodsList {
		idList = append(idList, good.GoodsId)
		if good.IsCurrent == 1 {
			currGoodsId = good.GoodsId
		}
	}

	goodsDetailList, err := mysql.GetGoodsById(ctx, idList)
	if err != nil {
		return nil, err
	}

	data := make([]*proto.GoodsInfo, 0, len(goodsDetailList))
	for _, good := range goodsDetailList {
		var headImgs []string
		json.Unmarshal([]byte(good.HeadImgs), &headImgs)
		data = append(data, &proto.GoodsInfo{
			GoodsId:     good.GoodsId,
			CategoryId:  good.CategoryId,
			Status:      int32(good.Status),
			Title:       good.Title,
			MarketPrice: int64(int32(good.MarketPrice / 100)),
			Price:       fmt.Sprintf("%.2f", float64(good.MarketPrice/100)),
			Brief:       good.Brief,
			HeadImgs:    headImgs,
		})
	}

	resp := &proto.GoodsListResp{
		CurrentGoodId: currGoodsId,
		Data:          data,
	}

	return resp, nil
}
