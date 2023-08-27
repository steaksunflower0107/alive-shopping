package mysql

import (
	"alivePlatform/repertory_service/dao/redis"
	"alivePlatform/repertory_service/errno"
	"alivePlatform/repertory_service/model"
	"context"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func GetRepertoryByGoodsId(ctx context.Context, goodsId int64) (*model.Repertory, error) {
	var data *model.Repertory
	err := db.WithContext(ctx).Model(&model.Repertory{}).
		Where("goods_id = ?", goodsId).
		First(&data).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		zap.L().Error("get repertory failed", zap.Int64("goods_id", goodsId))
		return nil, errno.ErrQueryFailed
	}

	return data, nil
}

//// ReduceRepertoryByGoodsId 扣减库存
//func ReduceRepertoryByGoodsId(ctx context.Context, goodsId, num int64) (*model.Repertory, error) {
//	var data *model.Repertory
//
//	// 悲观锁
//	//db.Transaction(func(tx *gorm.DB) error {
//	//	err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
//	//		Model(&model.Repertory{}).
//	//		Where("goods_id = ?", goodsId).
//	//		First(&data).Error
//	//
//	//	if err != nil && err != gorm.ErrRecordNotFound {
//	//		zap.L().Error("get repertory failed", zap.Int64("goods_id", goodsId))
//	//		return errno.ErrQueryFailed
//	//	}
//	//
//	//	if data.Num-num <= 0 {
//	//		return errno.ErrUnderRepertory
//	//	}
//	//
//	//	data.Num -= num
//	//
//	//	err = tx.Save(&data).Error
//	//	if err != nil {
//	//		zap.L().Error("reduceRepertory save failed", zap.Int64("goods_id", goodsId))
//	//		return err
//	//	}
//	//
//	//	return nil
//	//})
//
//	var (
//		retry     int
//		isSuccess bool
//	)
//	for !isSuccess && retry < 25 {
//		// 乐观锁
//		err := db.Model(&model.Repertory{}).
//			Where("goods_id = ?", goodsId).
//			First(&data).Error
//
//		if err != nil && err != gorm.ErrRecordNotFound {
//			zap.L().Error("get repertory failed", zap.Int64("goods_id", goodsId))
//			return nil, errno.ErrQueryFailed
//		}
//
//		if data.Num-num <= 0 {
//			return nil, errno.ErrUnderRepertory
//		}
//
//		data.Num -= num
//
//		rows := db.WithContext(ctx).Model(&model.Repertory{}).
//			Where("goods_id = ? and version = ? ", data.GoodsId, data.Version).
//			Updates(map[string]interface{}{
//				"goods_id": data.GoodsId,
//				"version":  data.Version + 1,
//				"num":      data.Num,
//			}).RowsAffected
//		if rows < 1 {
//			zap.L().Error("reduceRepertory save failed", zap.Int64("goods_id", goodsId))
//			retry++
//			continue
//		}
//
//		isSuccess = true
//		break
//	}
//
//	if !isSuccess {
//		return nil, errno.ErrReduceRepertoryFailed
//	}
//
//	return data, nil
//}

// ReduceRepertoryByGoodsId 扣减库存 基于redis的分布式锁
func ReduceRepertoryByGoodsId(ctx context.Context, goodsId, num int64) (*model.Repertory, error) {
	var data *model.Repertory

	// 创建key
	mutexName := fmt.Sprintf("repertory-%d", goodsId)
	// 创建锁
	mutex := redis.Rs.NewMutex(mutexName)
	// 获取锁
	if err := mutex.Lock(); err != nil {
		return nil, errno.ErrReduceRepertoryFailed
	}

	defer mutex.Unlock()

	db.Transaction(func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).
			Model(&model.Repertory{}).
			Where("goods_id = ?", goodsId).
			First(&data).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			zap.L().Error("get repertory failed", zap.Int64("goods_id", goodsId))
			return errno.ErrQueryFailed
		}

		if data.Num-num <= 0 {
			return errno.ErrUnderRepertory
		}

		data.Num -= num

		err = tx.WithContext(ctx).
			Save(&data).Error
		if err != nil {
			zap.L().Error("reduceRepertory save failed", zap.Int64("goods_id", goodsId))
			return err
		}

		return nil
	})

	return data, nil
}

// RollbackRepertoryByMsg 监听rocketmq消息进行库存回滚
func RollbackRepertoryByMsg(ctx context.Context, data model.OrderGoodsRepertoryInfo) error {
	// 先查询库存数据，需要放到事务操作中
	db.Transaction(func(tx *gorm.DB) error {
		var sr model.RepertoryRecord
		err := tx.WithContext(ctx).
			Model(&model.RepertoryRecord{}).
			Where("order_id = ? and goods_id = ? and status = 1", data.OrderId, data.GoodsId).
			First(&sr).Error
		// 没找到记录
		// 压根就没记录或者已经回滚过 不需要后续操作
		if err == gorm.ErrRecordNotFound {
			return nil
		}

		if err != nil {
			zap.L().Error("query stock_record by order_id failed", zap.Error(err), zap.Int64("order_id", data.OrderId), zap.Int64("goods_id", data.GoodsId))
			return err
		}

		// 开始归还库存
		var s model.Repertory
		err = tx.WithContext(ctx).
			Model(&model.Repertory{}).
			Where("goods_id = ?", data.GoodsId).
			First(&s).Error
		if err != nil {
			zap.L().Error("query stock by goods_id failed", zap.Error(err), zap.Int64("goods_id", data.GoodsId))
			return err
		}
		s.Num += data.Num  // 库存加上
		s.Lock -= data.Num // 锁定的库存减掉
		if s.Lock < 0 {    // 预扣库存不能为负
			return errno.ErrRollbackstockFailed
		}

		err = tx.WithContext(ctx).Save(&s).Error
		if err != nil {
			zap.L().Warn("RollbackStock stock save failed", zap.Int64("goods_id", s.GoodsId), zap.Error(err))
			return err
		}
		// 将库存扣减记录的状态变更为已回滚
		sr.Status = 3
		err = tx.WithContext(ctx).Save(&sr).Error
		if err != nil {
			zap.L().Warn("RollbackStock stock_record save failed", zap.Int64("goods_id", s.GoodsId), zap.Error(err))
			return err
		}
		return nil
	})
	return nil
}
