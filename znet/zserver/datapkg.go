package zserver

import (
	"bytes"
	"encoding/binary"
	"errors"
	"myftp/global"
	"myftp/znet"
)

// TODO: 可以和message合在一起

// 封包拆包对象，暂时不需要成员
type DataPack struct{}

// 实例化
func NewDataPack() *DataPack {
	return &DataPack{}
}

// 获取包头长度
func (dp *DataPack) GetHeadLen() uint32 {
	//Id uint32(4字节) +  DataLen uint32(4字节)
	return 8
}

// 封包
func (dp *DataPack) Pack(msg znet.IMessage) ([]byte, error) {
	dataBuffer := bytes.NewBuffer([]byte{})
	// (binary包)func Write(w io.Writer, order ByteOrder, data interface{}) error
	// 将data的binary编码格式写入w，data必须是定长值、定长值的切片、定长值的指针。
	// order指定写入数据的字节序，写入结构体时，名字中有'_'的字段会置为0。
	// 写 msgID
	if err := binary.Write(dataBuffer, binary.LittleEndian, msg.GetMsgID()); err != nil {
		return nil, err
	}

	// 写 dataLen
	if err := binary.Write(dataBuffer, binary.LittleEndian, msg.GetDataLen()); err != nil {
		return nil, err
	}

	// 写 data
	if err := binary.Write(dataBuffer, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}

	return dataBuffer.Bytes(), nil
}

// 拆包
func (dp *DataPack) Unpack(BinaryData []byte) (znet.IMessage, error) {
	dataBuffer := bytes.NewReader(BinaryData)

	// 只接收 head 的信息，得到 dataLen 和 msgID
	msg := &Message{}
	// 读 msgID
	if err := binary.Read(dataBuffer, binary.LittleEndian, &msg.ID); err != nil {
		return nil, err
	}

	// 读 msgLen
	if err := binary.Read(dataBuffer, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	// 判断 dataLen 的长度是否超出了自己设置的最大包长
	if global.GlobalObject.MaxPacketSize > 0 && global.GlobalObject.MaxPacketSize < msg.DataLen {
		return nil, errors.New("too large msg data")
	}

	// 通过解析出来的信息(id len)，通过conn再读取一次数据
	return msg, nil
}
