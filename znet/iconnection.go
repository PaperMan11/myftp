package znet

import (
	"myftp/service/srvservice"
	"net"
)

// 会话处理
type IConnection interface {
	Start()                                      // 启动连接
	Stop()                                       // 关闭连接
	GetTCPConnection() *net.TCPConn              // 获取连接套接字
	GetConnID() int64                            // 获取连接ID
	RemoteAddr() net.Addr                        // 获取远程客户端地址
	SendMsg(msgID uint32, data []byte) error     // 发数据 (无缓冲)
	SendBuffMsg(msgID uint32, data []byte) error // 发数据 (有缓冲)
	SetProperty(key string, value interface{})   // 设置连接属性
	GetProperty(key string) (interface{}, error) // 获取连接属性
	RemoveProperty(key string)                   // 删除连接属性
	// new add
	GetMyStack() *srvservice.MyStack
}
