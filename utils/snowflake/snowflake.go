package snowflake

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

/* 分布式 ID 生成器 */
// 一台机器一个 node
var node *snowflake.Node

// 初始化起始时间（能用69年）
// StartTime format: 2006-01-02
// machineID: 机器ID
func Init(StartTime string, machineID int64) (err error) {
	var st time.Time
	st, err = time.Parse("2006-01-02", StartTime)
	if err != nil {
		return
	}
	snowflake.Epoch = st.UnixNano() / 1000000
	node, err = snowflake.NewNode(machineID)
	return
}

func GetID() int64 {
	return node.Generate().Int64()
}
