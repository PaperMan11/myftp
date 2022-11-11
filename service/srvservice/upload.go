package srvservice

import (
	"encoding/gob"
	"io/ioutil"
	"myftp/model"
	"myftp/zlog"
	"os"
	"path"
	"strconv"
)

// PutOne 上传小文件
func (stack *MyStack) PutOne(fileName string, data []byte) error {
	// filePath := stack.curPath[stack.size-1] + "/" + fileName
	filePath := path.Join(stack.curPath[stack.size-1], fileName)
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		zlog.Errorf("create file failed")
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		zlog.Errorf("save file failed")
		return err
	}
	zlog.Info("upload file success")
	return nil
}

// GetUploadingFilePath 获取当前文件夹的上传文件夹路径（保存切片文件）
func (stack *MyStack) GetUploadingFilePath(uplodingDir string) string {
	return path.Join(stack.curPath[stack.size-1], uplodingDir)
}

// GetMetaDataFilePath 获取当前文件夹的 smetaDataFile（保存上传时大文件的基本信息）
func (stack *MyStack) GetMetaDataFilePath(fileName string) string {
	return path.Join(stack.curPath[stack.size-1], fileName) + ".slice"
}

// StoreMetaData 写元数据文件信息
func (stack *MyStack) StoreMetaData(filePath string, smetaData *model.ServerFileMetadata) error {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	return enc.Encode(smetaData)
}

// StoreMetaData 加载元数据文件信息
func (stack *MyStack) LoadMetaData(filePath string) (*model.ServerFileMetadata, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	var smetaData = new(model.ServerFileMetadata)
	dec := gob.NewDecoder(f)
	err = dec.Decode(smetaData)
	if err != nil {
		return nil, err
	}
	return smetaData, nil
}

// FindRetryReq 找到需要重传的文件片
func (stack *MyStack) FindRetryReq(upDir string, smetaData *model.ServerFileMetadata) []int {
	slices := make([]int, 0) // 记录缺失的文件片

	// 获取已保存的文件片需要
	storeSeq := make(map[string]bool)
	files, _ := ioutil.ReadDir(upDir)
	for _, file := range files {
		_, err := strconv.Atoi(file.Name())
		if err != nil {
			zlog.Errorf("[%s] 文件片有错误: %s", file.Name(), err)
			continue
		}
		storeSeq[file.Name()] = true
	}

	// 找到缺失的文件片
	i := 0
	for ; i < smetaData.SliceNum && len(storeSeq) > 0; i++ {
		seqStr := strconv.Itoa(i)
		if _, ok := storeSeq[seqStr]; ok {
			delete(storeSeq, seqStr)
		} else {
			slices = append(slices, i)
		}
	} // -1指代slices的最大数字序号到最后一片都没有收到
	if i < smetaData.SliceNum {
		slices = append(slices, i)
		i += 1
		if i < smetaData.SliceNum {
			slices = append(slices, -1)
		}
	}
	zlog.Infof("需要重传的文件片: %v", slices)
	return slices
}
