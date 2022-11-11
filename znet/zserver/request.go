package zserver

import "myftp/znet"

// 客户端请求对象
type Request struct {
	conn znet.IConnection // 客户端连接
	msg  znet.IMessage    // 请求消息
}

// 获取请求连接信息
func (req *Request) GetConnection() znet.IConnection {
	return req.conn
}

// 获取请求消息的数据
func (req *Request) GetData() []byte {
	return req.msg.GetData()
}

// 获取消息ID
func (req *Request) GetMsgID() uint32 {
	return req.msg.GetMsgID()
}
