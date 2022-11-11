package znet

/*
包处理
	封包和拆包
	解决粘包问题
*/

// request -> message -> package
type IDataPack interface {
	GetHeadLen() uint32                // 获取包头长度
	Pack(msg IMessage) ([]byte, error) // 封包
	Unpack([]byte) (IMessage, error)   // 拆包
}
