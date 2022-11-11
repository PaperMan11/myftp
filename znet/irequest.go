package znet

// 把客户端 请求的连接信息 和 请求的数据 封装到request中
type IRequset interface {
	GetConnection() IConnection // 获取请求连接信息
	GetData() []byte            // 获取请求消息的数据
	GetMsgID() uint32           // 获取消息ID
}
