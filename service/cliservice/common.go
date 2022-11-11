package cliservice

import (
	"encoding/gob"
	"encoding/json"
	"io"
	"myftp/model"
	"myftp/znet"
	"myftp/znet/zserver"
	"net"
	"os"
	"path"
	"path/filepath"
)

// StoreMetaData 写uploding文件
func StoreMetaData(filePath string, metadata *model.ClientFileMetadata) error {
	// 写入文件
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	return enc.Encode(metadata)
}

// getUploaderMetaPath 获取uploding文件路径
func getUploaderMetaPath(filePath string) string {
	paths, fileName := filepath.Split(filePath)
	// 组合成隐藏文件
	return path.Join(paths, "."+fileName+".uploading")
}

// Pack 封包
func Pack(msgId uint32, metadata interface{}) ([]byte, error) {
	b, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	req := zserver.NewMsgPackage(msgId, b)
	pac := zserver.NewDataPack()
	data, err := pac.Pack(req)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Unpack 解析包
func Unpack(conn net.Conn) (znet.IMessage, error) {
	pac := zserver.NewDataPack()
	//先读出流中的head部分
	headData := make([]byte, pac.GetHeadLen())
	_, err := io.ReadFull(conn, headData) //ReadFull 会把msg填充满为止
	if err != nil {
		return nil, err
	}
	//将headData字节流 拆包到msg中
	msgHead, err := pac.Unpack(headData)
	if err != nil {
		return nil, err
	}
	if msgHead.GetDataLen() > 0 {
		//msg 是有data数据的，需要再次读取data数据
		msg := msgHead.(*zserver.Message)
		msg.Data = make([]byte, msg.GetDataLen())

		//根据dataLen从io中读取字节流
		_, err := io.ReadFull(conn, msg.Data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	}
	return msgHead, nil
}
