package snowflake

import (
	"errors"
	"time"

	sf "github.com/bwmarrin/snowflake"
)

const (
	_defaultStartTime = "2023-7-21" // 默认开始时间
)

var node *sf.Node

func Init(startTime string, machineID int64) (err error) {
	if machineID < 0 {
		return errors.New("snowflake need machineID")
	}
	if len(startTime) == 0 {
		startTime = _defaultStartTime
	}
	var st time.Time
	st, err = time.Parse("2006-01-02", startTime)
	if err != nil {
		return
	}
	sf.Epoch = st.UnixNano() / 1000000 // 时间戳的开始时间，默认从1970年开始计算
	node, err = sf.NewNode(machineID)  // 机器编号，最多1024
	return
}

func GenID() int64 {
	return node.Generate().Int64()
}

func GenIDStr() string {
	return node.Generate().String()
}
