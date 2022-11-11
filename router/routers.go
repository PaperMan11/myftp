package router

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"myftp/model"
	"myftp/utils"
	"myftp/zlog"
	"myftp/znet"
	"myftp/znet/zserver"
	"os"
	"path"
	"strconv"
)

/*
	0 Ls
	1 Upload
	2 UploadingsStat
	3 CreateUploadDir
	4 UploadBySlice
	5 MergeSliceFiles
	6 Cd
	7 Download
*/

// ls
type LsCommand struct {
	zserver.BaseRouter
}

func (lc *LsCommand) Handle(req znet.IRequset) {
	resp := model.LsResp{
		Code: 200,
		Msg:  "",
		Data: make([]string, 0),
	}
	cmd := string(req.GetData())
	// 读取客户端数据
	fmt.Printf("recv from client: [msgID=%d] [data=%s]\n", req.GetMsgID(), cmd)
	ms := req.GetConnection().GetMyStack()
	strSlice, err := ms.ParseLsCmd(cmd)
	if err != nil {
		zlog.Errorf("ParseLsCmd failed: %s", err)
		resp.Code = 500
		resp.Msg = err.Error()
		goto SEND
	}
	// byteSlice, err := json.Marshal(strSlice)
	resp.Data = strSlice
SEND:
	byteSlice, _ := json.Marshal(resp)
	req.GetConnection().SendBuffMsg(0, byteSlice)
}

type CdCommand struct {
	zserver.BaseRouter
}

func (cd *CdCommand) Handle(req znet.IRequset) {
	cmd := string(req.GetData())
	// 读取客户端数据
	fmt.Printf("recv from client: [msgID=%d] [data=%s]\n", req.GetMsgID(), cmd)
	ms := req.GetConnection().GetMyStack()
	err := ms.ParseCdCmd(cmd)
	if err == nil {
		req.GetConnection().SendBuffMsg(6, []byte(""))
		return
	}
	req.GetConnection().SendBuffMsg(6, []byte(err.Error()))
}

// upload small file
type Upload struct {
	zserver.BaseRouter
}

func (up *Upload) Handle(req znet.IRequset) {
	var putReq model.SingleFileInfo
	if err := json.Unmarshal(req.GetData(), &putReq); err != nil {
		zlog.Errorf("json.Unmarshal failed: %s", err)
		req.GetConnection().SendBuffMsg(1, []byte(err.Error()))
		return
	}
	fmt.Println("put file:", putReq.Filename)
	stack := req.GetConnection().GetMyStack()
	err := stack.PutOne(putReq.Filename, putReq.Data)
	if err != nil {
		zlog.Errorf("PutOne failed: %s", err)
		req.GetConnection().SendBuffMsg(1, []byte(err.Error()))
		return
	}
	req.GetConnection().SendBuffMsg(1, []byte("put file success"))
}

// 获取上传文件在服务端的状态（断点续传、重载）
type UploadingStat struct {
	zserver.BaseRouter
}

func (upst *UploadingStat) Handle(req znet.IRequset) {
	var upReq model.UploadingFileReq // req
	retrySeq := model.SliceSeq{      // resp
		Slices: make([]int, 0),
	}

	json.Unmarshal(req.GetData(), &upReq)

	stack := req.GetConnection().GetMyStack()
	smetaDataPath := stack.GetMetaDataFilePath(upReq.Filename)
	uploadingDir := stack.GetUploadingFilePath(upReq.Fid)

	smetaData, err := stack.LoadMetaData(smetaDataPath)
	if err != nil || smetaData.Fid != upReq.Fid {
		retrySeq.Slices = append(retrySeq.Slices, -2) // 文件uuid不正确 通知停止上传
		goto SEND
	}

	if exists := stack.CheckFileExists(upReq.Fid); exists {
		// 查询需要重新上传的切片
		retrySeq.Slices = stack.FindRetryReq(uploadingDir, smetaData)
	}

SEND:
	b, _ := json.Marshal(retrySeq)
	req.GetConnection().SendBuffMsg(2, b)
}

// 开始上传文件、创建文件信息隐藏文件
type CreateUploadDir struct {
	zserver.BaseRouter
}

func (crdir *CreateUploadDir) Handle(req znet.IRequset) {
	var cMetaData model.ClientFileMetadata
	if err := json.Unmarshal(req.GetData(), &cMetaData); err != nil {
		zlog.Errorf("json.Unmarshal failed: %s", err)
		req.GetConnection().SendBuffMsg(3, []byte(err.Error()))
		return
	}
	stack := req.GetConnection().GetMyStack()
	metaDataPath := stack.GetMetaDataFilePath(cMetaData.Filename)
	if exists, _ := utils.FileExists(metaDataPath); exists {
		zlog.Info("upload file exists")
		req.GetConnection().SendBuffMsg(3, []byte("exists"))
		return
	}
	if err := stack.MkdirUploading(cMetaData.Fid); err != nil {
		zlog.Errorf("create uploading file failed: %s", err)
		req.GetConnection().SendBuffMsg(3, []byte(err.Error()))
		return
	}

	// 上传文件基本信息
	sMetaData := model.ServerFileMetadata{
		ClientFileMetadata: cMetaData,
		State:              "uploading",
	}

	// 写入元数据文件
	if err := stack.StoreMetaData(metaDataPath, &sMetaData); err != nil {
		zlog.Errorf("StoreMetaData failed: %s", err)
		req.GetConnection().SendBuffMsg(3, []byte(err.Error()))
	}
	req.GetConnection().SendBuffMsg(3, []byte("ready"))
}

// upload big file
type UploadBySlice struct {
	zserver.BaseRouter
}

func (ups *UploadBySlice) Handle(req znet.IRequset) {
	var putReq model.FilePart
	if err := json.Unmarshal(req.GetData(), &putReq); err != nil {
		zlog.Errorf("json.Unmarshal failed: %s", err)
		req.GetConnection().SendBuffMsg(4, []byte(err.Error()))
		return
	}
	// fmt.Println("Upload Slice:", putReq)
	// 检查分片是否存在
	path := path.Join(putReq.Fid, strconv.Itoa(putReq.Index))
	stack := req.GetConnection().GetMyStack()
	exists := stack.CheckFileExists(path)
	if exists {
		zlog.Infof("%s分片文件已存在，直接丢弃, part.Fid: %s, index: %s\n", path, putReq.Fid, putReq.Index)
		return
	}
	stack.PutOne(path, putReq.Data)
	req.GetConnection().SendBuffMsg(4, []byte("ok"))
}

// merge file
type MergeSliceFiles struct {
	zserver.BaseRouter
}

func (mfs *MergeSliceFiles) Handle(req znet.IRequset) {
	var cMetaData model.ClientFileMetadata
	if err := json.Unmarshal(req.GetData(), &cMetaData); err != nil {
		zlog.Errorf("json.Unmarshal failed: %s", err)
		req.GetConnection().SendBuffMsg(5, []byte(err.Error()))
		return
	}

	stack := req.GetConnection().GetMyStack()
	uploadingDir := stack.GetUploadingFilePath(cMetaData.Fid)
	hash := md5.New()

	filePath := stack.GetNewFilePath(cMetaData.Filename) // 最终合并的文件
	if exists, _ := utils.FileExists(filePath); exists {
		os.Remove(filePath)
	}
	newFile, _ := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer newFile.Close()
	// 计算MD5
	for i := 0; i < cMetaData.SliceNum; i++ {
		slicePath := path.Join(uploadingDir, strconv.Itoa(i))
		sliceFile, err := os.Open(slicePath)
		if err != nil {
			zlog.Errorf("read slice file failed: %s", err)
			req.GetConnection().SendBuffMsg(5, []byte("failed"))
			return
		}
		io.Copy(hash, sliceFile)
		sliceFile.Seek(0, 0)
		io.Copy(newFile, sliceFile)
		sliceFile.Close()
	}
	md5Sum := hex.EncodeToString(hash.Sum(nil))
	if md5Sum != cMetaData.Md5sum {
		zlog.Errorf("原始md5 [%s], 计算的md5 [%s]", cMetaData.Md5sum, md5Sum)
		zlog.Error("文件MD5校验不通过，数据传输有误，需要重新上传文件")
		req.GetConnection().SendBuffMsg(5, []byte("failed"))
		// 删除临时文件夹
		os.RemoveAll(uploadingDir)
		os.RemoveAll(stack.GetMetaDataFilePath(cMetaData.Filename))
		return
	}

	// // 更新元数据信息
	// smetaDataPath := stack.GetMetaDataFilePath(cMetaData.Filename)
	// smetaData, _ := stack.LoadMetaData(smetaDataPath)
	// smetaData.Md5sum = md5Sum
	// smetaData.State = "active"
	// stack.StoreMetaData(smetaDataPath, smetaData)

	// 删除临时文件夹
	os.RemoveAll(uploadingDir)
	os.RemoveAll(stack.GetMetaDataFilePath(cMetaData.Filename))

	req.GetConnection().SendBuffMsg(5, []byte("success"))
}

// Download 下载
type Download struct {
	zserver.BaseRouter
}

func (d *Download) Handle(req znet.IRequset) {
	var (
		downloadReq model.DownloadFileReq
		size        int64
		err         error
		f           *os.File
		buf         []byte = make([]byte, 1024)
		filePath    string
	)
	downlaodResp := model.DownloadFileResp{
		Code: 0,
		Msg:  "",
		Data: make([]byte, 0),
	}
	stack := req.GetConnection().GetMyStack()

	if err := json.Unmarshal(req.GetData(), &downloadReq); err != nil {
		zlog.Errorf("json.Unmarshal failed: %s", err)
		downlaodResp.Code = 2
		downlaodResp.Msg = err.Error()
		goto SEND
	}
	filePath = stack.GetNewFilePath(downloadReq.FileName)

	// 检查文件是否存在且不为目录
	if exists := stack.CheckFileExists(downloadReq.FileName); !exists {
		zlog.Errorf("%s: not found", filePath)
		downlaodResp.Code = 2
		downlaodResp.Msg = "file not found"
		goto SEND
	}
	if utils.IsDir(filePath) {
		zlog.Errorf("%s: file is directory", filePath)
		downlaodResp.Code = 2
		downlaodResp.Msg = "file is directory"
		goto SEND
	}
	size, _ = utils.FileSize(filePath)
	if size <= int64(downloadReq.Index) {
		zlog.Infof("%s: file is up to date", filePath)
		downlaodResp.Code = 2
		downlaodResp.Msg = "file is up to date"
		goto SEND
	}

	// 根据index发送
	f, err = os.Open(filePath)
	if err != nil {
		zlog.Errorf("%s: open file failed", filePath)
		downlaodResp.Code = 2
		downlaodResp.Msg = "open file failed"
		goto SEND
	}
	defer f.Close()
	_, err = f.Seek(int64(downloadReq.Index), 0)
	if err != nil {
		zlog.Errorf("%s: file Seek failed: %s", filePath, err)
		downlaodResp.Code = 2
		downlaodResp.Msg = "file Seek failed"
		goto SEND
	}

	for {
		n, err := f.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				downlaodResp.Code = 1
				b, _ := json.Marshal(downlaodResp)
				req.GetConnection().SendBuffMsg(7, b)
				break
			}
			downlaodResp.Code = 2
			downlaodResp.Msg = err.Error()
			goto SEND
		}
		downlaodResp.Data = buf[:n]
		b, _ := json.Marshal(downlaodResp)
		req.GetConnection().SendBuffMsg(7, b)
	}
	return
SEND:
	b, _ := json.Marshal(downlaodResp)
	req.GetConnection().SendBuffMsg(7, b)
}
