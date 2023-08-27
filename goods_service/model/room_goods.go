package model

// RoomGoods 直播间与商品对应
type RoomGoods struct {
	BaseModel
	RoomId    int64
	GoodsId   int64
	Weight    int64
	IsCurrent int8 `gorm:"is_current"`
}

func (RoomGoods) TableName() string {
	return "platform_room_goods"
}
