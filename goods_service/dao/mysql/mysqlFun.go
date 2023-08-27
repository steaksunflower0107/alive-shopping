package mysql

import (
	"alivePlatform/goods_service/errno"
	"alivePlatform/goods_service/model"
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetGoodsByRoomId(ctx context.Context, roomId int64) ([]*model.RoomGoods, error) {
	var data []*model.RoomGoods
	err := db.WithContext(ctx).Model(&model.RoomGoods{}).
		Where("room_id = ?", roomId).Order("weight").
		Find(&data).Error

	if err != nil && err != gorm.ErrEmptySlice {
		return nil, errno.ErrQueryFailed
	}

	return data, nil
}

func GetGoodsById(ctx context.Context, idList []int64) ([]*model.Goods, error) {
	var data []*model.Goods
	err := db.WithContext(ctx).Model(&model.Goods{}).
		Where("goods_id IN ?", idList).
		Clauses(clause.OrderBy{
			Expression: clause.Expr{SQL: "FIELD(goods_id,?)", Vars: []interface{}{idList}, WithoutParentheses: true},
		}).Find(&data).Error

	if err != nil && err != gorm.ErrEmptySlice {
		return nil, errno.ErrQueryFailed
	}

	return data, nil
}
