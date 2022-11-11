package zserver

type Message struct {
	ID      uint32 // 消息ID
	DataLen uint32 // 消息长度
	Data    []byte // 消息内容
}

// 创建一个消息对象
func NewMsgPackage(id uint32, data []byte) *Message {
	return &Message{
		ID:      id,
		DataLen: uint32(len(data)),
		Data:    data,
	}
}

// 获取数据长度
func (msg *Message) GetDataLen() uint32 {
	return msg.DataLen
}

// 获取消息ID
func (msg *Message) GetMsgID() uint32 {
	return msg.ID
}

// 获取数据
func (msg *Message) GetData() []byte {
	return msg.Data
}

// 设置消息ID
func (msg *Message) SetMsgID(msgID uint32) {
	msg.ID = msgID
}

// 设置数据
func (msg *Message) SetData(data []byte) {
	msg.Data = data
}

// 设置数据长度
func (msg *Message) SetDataLen(dataLen uint32) {
	msg.DataLen = dataLen
}
