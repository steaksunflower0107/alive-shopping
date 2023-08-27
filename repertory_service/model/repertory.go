package model

type Repertory struct {
	BaseModel
	GoodsId int64
	Num     int64
	Lock    int64
}

// TableName 对应数据库表名
func (Repertory) TableName() string {
	return "platform_repertory"
}
