package znet

// 将请求的消息封装到message中，定义抽象层接口
type IMessage interface {
	GetDataLen() uint32 // 获取数据长度
	GetMsgID() uint32   // 获取消息ID
	GetData() []byte    // 获取数据

	SetMsgID(uint32)   // 设置消息ID
	SetData([]byte)    // 设置数据
	SetDataLen(uint32) // 设置数据长度
}
