package cliservice

import (
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"myftp/model"
	"myftp/utils"
	"myftp/utils/uuid"
	"myftp/zlog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Uploader 上传器
type Uploader struct {
	Conn          net.Conn
	FileMetaData  model.ClientFileMetadata // 文件元数据
	SliceSeq      model.SliceSeq           // 需要重传的序号
	waitGoroutine sync.WaitGroup           // 同步
	NewLoader     bool                     // 是否为新创建的上传器
	FilePath      string                   // 上传文件路径
	SliceBytes    int                      // 文件片大小
	RetryChannel  chan *model.FilePart     // 重传channel
	MaxGoChannel  chan struct{}            // 限制上传goroutine数量
	StartTime     int64                    // 上传开始时间
	StopUp        bool                     // 停止上传
}

// NewUploader 新建上传器
func NewUploader(conn net.Conn, filePath string, sliceBytes int) *Uploader {
	uuid := uuid.UUID()

	fileStat, err := os.Stat(filePath)
	if err != nil {
		zlog.Errorf("读取文件 %s 失败 err: %s", filePath, err)
		return nil
	}
	fileSize := fileStat.Size()
	if fileSize <= 0 {
		zlog.Errorf("%s文件是空的", filePath)
		return nil
	}

	// 计算文件文件切片数量
	sliceNum := int(math.Ceil(float64(fileSize) / float64(sliceBytes)))

	metadata := model.ClientFileMetadata{
		Fid:        uuid,
		Filesize:   fileSize,
		Filename:   filepath.Base(filePath),
		SliceNum:   sliceNum,
		Md5sum:     "",
		ModifyTime: fileStat.ModTime(),
	}

	uploader := &Uploader{
		Conn:         conn,
		FileMetaData: metadata,
		SliceSeq:     model.SliceSeq{Slices: []int{-1}}, // 表示全部上传
		NewLoader:    true,
		FilePath:     filePath,
		SliceBytes:   sliceBytes,
		RetryChannel: make(chan *model.FilePart, model.DownloadRetryChannelNum),
		MaxGoChannel: make(chan struct{}, model.DpGoroutineMaxNumPerFile),
		StartTime:    time.Now().Unix(),
		StopUp:       false,
	}

	err = StoreMetaData(getUploaderMetaPath(filePath), &metadata)
	if err != nil {
		zlog.Errorf("新建上传器失败: %s", err)
		return nil
	}

	return uploader
}

// GetUploader 获取一个上传器，用以初始化之前未上传完的
func GetUploader(conn net.Conn, filePath string, sliceBytes int) *Uploader {
	metaPath := getUploaderMetaPath(filePath) // 上传时临时的隐藏文件
	if exists, _ := utils.FileExists(metaPath); exists {
		file, err := os.Open(metaPath)
		if err != nil {
			zlog.Errorf("获取元数据文件状态失败:%s", err)
			os.Remove(metaPath)
			return nil
		}

		var metadata model.ClientFileMetadata
		dec := gob.NewDecoder(file)
		err = dec.Decode(&metadata)
		if err != nil {
			zlog.Errorf("解析uploding文件失败: %s", err)
			os.Remove(metaPath)
			return nil
		}

		curFileStat, err := os.Stat(filePath)
		if err != nil {
			zlog.Errorf("获取上传文件属性失败: %s", err)
			os.Remove(metaPath)
			return nil
		}

		// 比较文件数据
		if metadata.Filesize != curFileStat.Size() || metadata.ModifyTime != curFileStat.ModTime() {
			zlog.Infof("%s 文件已修改，重新上传", filePath)
			return nil
		}

		uploader := &Uploader{
			Conn:         conn,
			FileMetaData: metadata,
			FilePath:     filePath,
			SliceBytes:   sliceBytes,
			NewLoader:    false,
			RetryChannel: make(chan *model.FilePart, model.DownloadRetryChannelNum),
			MaxGoChannel: make(chan struct{}, model.DpGoroutineMaxNumPerFile),
			StartTime:    time.Now().Unix(),
			StopUp:       false,
		}

		// 向服务器获取需要重传的切片
		sliceSeq, err := uploader.getRetrySlice(metadata.Fid, metadata.Filename)
		if err != nil {
			sliceSeq = &model.SliceSeq{
				Slices: []int{-1}, // 表示全部上传
			}
		}
		if sliceSeq.Slices[0] == -2 {
			uploader.StopUp = true
			return uploader
		}
		uploader.SliceSeq = *sliceSeq
		return uploader
	}
	// 返回nil创建新的上传器
	return nil
}

func UploaderSmallFile(conn net.Conn, fileName string) error {
	size, _ := utils.FileSize(fileName)
	fileReq := model.SingleFileInfo{
		Filename: filepath.Base(fileName),
		Filesize: size,
	}
	f, _ := os.Open(fileName)
	defer f.Close()
	fileReq.Data = make([]byte, size)
	f.ReadAt(fileReq.Data, 0)
	data, err := Pack(1, fileReq)
	if err != nil {
		zlog.Errorf("Pack failed: %s", err)
		return err
	}
	conn.Write(data)
	msg, err := Unpack(conn)
	if err != nil {
		zlog.Errorf("Unpack failed: %s", err)
		return err
	}
	zlog.Info(string(msg.GetData()))
	zlog.Infof("%s 小文件上传成功", filepath.Base(fileName))
	return nil
}

// getRetrySlice 获取需要重传的切片
func (u *Uploader) getRetrySlice(fid string, filename string) (*model.SliceSeq, error) {
	data, err := Pack(2, u.FileMetaData)
	if err != nil {
		zlog.Errorf("Pack failed: %s", err)
		return nil, err
	}
	u.Conn.Write(data)

	msg, err := Unpack(u.Conn)
	if err != nil {
		zlog.Errorf("Unpack failed: %s", err)
		return nil, err
	}

	sliceSeq := model.SliceSeq{
		Slices: make([]int, 0),
	}
	err = json.Unmarshal(msg.GetData(), &sliceSeq)
	if err != nil {
		zlog.Errorf("json.Unmarshal failed: %s", err)
		return nil, err
	}
	zlog.Info("需要上传的切片:", sliceSeq.Slices)
	return &sliceSeq, nil
}

func (u *Uploader) UploadFileBySlice() error {
	if u.NewLoader {
		// 新上传文件需要通知服务器初始化
		if err := u.sendInitCmd(); err != nil {
			zlog.Errorf("sendInitCmd: %s", err)
			os.Remove(getUploaderMetaPath(u.FilePath)) // 删除uploading文件
			return err
		}
	}

	md5sum := u.FileMetaData.Md5sum
	if len(u.SliceSeq.Slices) == 0 && md5sum != "" {
		// 分片都已保存在服务端了，提出合并请求即可
		if err := u.sendMergeCmd(); err != nil {
			zlog.Errorf("sendMergeCmd: %s", err)
			return err
		}
		os.Remove(getUploaderMetaPath(u.FilePath)) // 删除uploading文件
		return nil
	}

	f, err := os.Open(u.FilePath)
	if err != nil {
		zlog.Errorf("open file failed: %s", err)
		return err
	}
	defer f.Close()

	// 启动重传 goroutine
	go u.retryUploadSlice()

	hash := md5.New()

	startIndex := 0
	if len(u.SliceSeq.Slices) > 0 && u.SliceSeq.Slices[0] >= 0 && md5sum != "" {
		startIndex = u.SliceSeq.Slices[0] // 重传时确定文件的偏移量
	}
	f.Seek(int64(startIndex)*int64(u.SliceBytes), 0) // 跳过不需要的部分

	for i := startIndex; i < u.FileMetaData.SliceNum; i++ {
		tmpData := make([]byte, u.SliceBytes)
		nr, err := f.Read(tmpData[:])
		if err != nil {
			zlog.Error("read file slice failed")
			return err
		}
		if md5sum == "" {
			hash.Write(tmpData[:nr])
		}
		tmpData = tmpData[:nr]

		if len(u.SliceSeq.Slices) <= 0 {
			if md5sum == "" {
				continue // 没有重传的还需计算md5
			}
			break
		} else if u.SliceSeq.Slices[0] != -1 && i != u.SliceSeq.Slices[0] {
			continue // 不需要重传的直接跳过
		}

		if u.SliceSeq.Slices[0] != -1 {
			// 去掉重传的片（每传完一个就去掉）
			u.SliceSeq.Slices = u.SliceSeq.Slices[1:]
		}

		// 构造切片并上传
		part := &model.FilePart{
			Fid:   u.FileMetaData.Fid,
			Index: i,
			Data:  tmpData,
		}
		u.waitGoroutine.Add(1)
		go u.uploadSlice(part)
	}

	if md5sum == "" {
		// 计算MD5
		md5sum = hex.EncodeToString(hash.Sum(nil))
		// 保存md5到元数据文件
		u.FileMetaData.Md5sum = md5sum
		err := StoreMetaData(getUploaderMetaPath(u.FilePath), &u.FileMetaData)
		if err != nil {
			return err
		}
	}
	zlog.Info("等待分片上传完成")
	u.waitGoroutine.Wait()
	if time.Now().Unix()-u.StartTime > model.UploadTimeout {
		zlog.Errorf("%s 上传超时", u.FileMetaData.Filename)
		return errors.New("上传超时")
	}
	// 删除元数据
	defer os.Remove(getUploaderMetaPath(u.FilePath))

	// 发起合并
	if err = u.sendMergeCmd(); err != nil {
		zlog.Errorf("%s 文件合并失败", u.FileMetaData.Filename)
		return err
	}
	zlog.Infof("%s 文件上传成功", u.FileMetaData.Filename)
	return nil
}

func (u *Uploader) retryUploadSlice() {
	for part := range u.RetryChannel {
		// 检查上传是否超时了，如果超时了则开始快速退出
		if time.Now().Unix()-u.StartTime > model.UploadTimeout {
			zlog.Error("上传超时，请重试")
			u.waitGoroutine.Done()
			continue
		}
		zlog.Infof("重传文件分片，文件名:%s, 分片序号:%d", u.FileMetaData.Filename, part.Index)
		go u.uploadSlice(part)
	}
}

// sendInitCmd 通知服务器初始化
func (u *Uploader) sendInitCmd() error {
	data, err := Pack(3, u.FileMetaData)
	if err != nil {
		zlog.Errorf("Pack failed: %s", err)
		return err
	}
	u.Conn.Write(data)

	msg, err := Unpack(u.Conn)
	if err != nil {
		zlog.Errorf("Unpack failed: %s", err)
		return err
	}

	if string(msg.GetData()) == "ready" {
		return nil
	}
	return errors.New(string(msg.GetData()))
}

// sendInitCmd 通知服务器合并文件
func (u *Uploader) sendMergeCmd() error {
	data, err := Pack(5, u.FileMetaData)
	if err != nil {
		zlog.Errorf("Pack failed: %s", err)
		return err
	}
	u.Conn.Write(data)

	msg, err := Unpack(u.Conn)
	if err != nil {
		zlog.Errorf("Unpack failed: %s", err)
		return err
	}
	if string(msg.GetData()) == "success" {
		return nil
	}
	return errors.New(string(msg.GetData()))
}

// uploadSlice 上传分片
func (u *Uploader) uploadSlice(part *model.FilePart) {
	// 控制上传文件片goroutine数量
	u.MaxGoChannel <- struct{}{}
	defer func() {
		<-u.MaxGoChannel
	}()

	// upload
	data, err := Pack(4, part)
	if err != nil {
		zlog.Errorf("Pack [%d] failed: %s", part.Index, err)
		return
	}
	u.Conn.Write(data)
	msg, err := Unpack(u.Conn)
	if err != nil || string(msg.GetData()) != "ok" {
		zlog.Errorf("上传文件分片失败，文件ID: %s, 序号：%d", part.Fid, part.Index)
		// 进行切片重传
		u.RetryChannel <- part
		return
	}
	// 不用defer的原因: 有重传的存在，重传的在重传中 Done
	u.waitGoroutine.Done()
}
