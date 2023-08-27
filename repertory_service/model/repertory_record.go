package model

type RepertoryRecord struct {
	BaseModel // 嵌入默认的7个字段

	OrderId int64
	GoodsId int64
	Num     int64
	Status  int32
}

// TableName 声明表名
func (RepertoryRecord) TableName() string {
	return "platform_repertory_record"
}
