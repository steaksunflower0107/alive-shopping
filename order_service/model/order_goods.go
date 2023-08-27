package model

type OrderDetail struct {
	BaseModel // 嵌入默认的7个字段

	OrderId int64
	GoodsId int64
	UserId  int64
	Num     int64
}

func (OrderDetail) TableName() string {
	return "platform_order_detail"
}
