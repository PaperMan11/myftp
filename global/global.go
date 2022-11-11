package global

import (
	"encoding/json"
	"io/ioutil"
	"myftp/znet"
)

type GlobalObj struct {
	// Server
	TcpServer znet.IServer // 全局服务器对象
	Host      string       // 当前服务器IP
	TcpPort   int          // 当前服务器监听端口
	Name      string       // 当前服务器名称

	// Zinx
	Version         string // 当前 Zinx 版本号
	MaxPacketSize   uint32 // 数据包的最大值
	MaxConn         int    // 最大连接数
	WorkerPoolSize  uint32 // 业务工作 Worker 池的数量
	MaxWokerTaskLen uint32 // 业务工作 Worker 对应负责的任务队列最大任务存储数量
	MaxMsgChanLen   uint32 // 发数据的缓冲大小

	// Config file path
	ConfFilePath string
}

// 定义一个全局对象 (其他模块能够访问到)
var GlobalObject *GlobalObj

// 读取配置文件
func (g *GlobalObj) Reload() {
	data, err := ioutil.ReadFile(g.ConfFilePath)
	if err != nil {
		panic(err)
	}
	// 将 json 数据解析到结构体中
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}

func init() {
	// 初始化 GlobalObject 对象
	GlobalObject = &GlobalObj{
		Name:            "Ftp",
		Version:         "V0.5",
		TcpPort:         7777,
		Host:            "0.0.0.0",
		MaxConn:         12000,
		MaxPacketSize:   1024 * 1024 * 6,
		WorkerPoolSize:  10,
		MaxWokerTaskLen: 1024,
		ConfFilePath:    "../../config/zinx.json",
		MaxMsgChanLen:   1024,
	}
	// 从配置文件中加载一些用户配置参数
	GlobalObject.Reload()
}
