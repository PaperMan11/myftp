package cliservice

import (
	"encoding/json"
	"errors"
	"myftp/model"
	"myftp/utils"
	"myftp/zlog"
	"net"
	"os"
	"path"
)

type Downloader struct {
	Conn         net.Conn
	FileMetaData model.DownloadFileReq // 下载请求
}

func NewDownloader(conn net.Conn, fileName, storeDir string) *Downloader {
	metadata := model.DownloadFileReq{
		FileName: fileName,
		Index:    0,
	}
	// 是否有该文件
	filePath := path.Join(storeDir, fileName)
	if exists, _ := utils.FileExists(filePath); exists && !utils.IsDir(fileName) {
		size, err := utils.FileSize(filePath)
		if err != nil {
			zlog.Errorf("get %s size failed: %s", filePath, err)
			return nil
		}
		metadata.Index = int(size)
	}
	return &Downloader{
		Conn:         conn,
		FileMetaData: metadata,
	}
}

func (d *Downloader) DownloadFile(conn net.Conn, fileName, storeDir string) error {
	data, err := Pack(7, d.FileMetaData)
	if err != nil {
		zlog.Errorf("Pack failed: %s", err)
		return err
	}
	d.Conn.Write(data)

	filePath := path.Join(storeDir, fileName)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		zlog.Errorf("OpenFile failed: %s", err)
		return err
	}
	defer f.Close()
	for {
		msg, err := Unpack(d.Conn)
		if err != nil {
			zlog.Errorf("Unpack failed: %s", err)
			return err
		}
		var resp model.DownloadFileResp
		err = json.Unmarshal(msg.GetData(), &resp)
		if err != nil {
			zlog.Errorf("json.Unmarshal failed: %s", err)
			return err
		}
		if resp.Code == 2 {
			zlog.Errorf("download failed: %s", resp.Msg)
			return errors.New(resp.Msg)
		} else if resp.Code == 1 {
			zlog.Info("download complete")
			return nil
		} else {
			f.Write(resp.Data[:])
		}
	}
}
