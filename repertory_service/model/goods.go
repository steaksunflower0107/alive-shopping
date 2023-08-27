package model

type Goods struct {
	BaseModel
	GoodsId     int64
	CategoryId  int64
	BrandName   string
	Code        int64
	Status      int8
	Title       string
	MarketPrice int64
	Price       int64
	Brief       string
	HeadImgs    string
	Videos      string
	Detail      string
	ExtJson     string
}

// TableName 对应数据库表名
func (Goods) TableName() string {
	return "platform_goods"
}
