package srvservice

import (
	"myftp/utils"
	"os"
	"path"
)

// GetNewFilePath 获取文件的绝对路径
func (stack *MyStack) GetNewFilePath(fileName string) string {
	return path.Join(stack.curPath[stack.size-1], fileName)
}

// MkdirUploading 在当前路径创建目录
func (stack *MyStack) MkdirUploading(dirName string) error {
	uploadDir := path.Join(stack.curPath[stack.size-1], dirName)
	return os.Mkdir(uploadDir, 0766)
}

// CheckFileExists 检查文件在当前文件夹是否存在
func (stack *MyStack) CheckFileExists(fileName string) bool {
	absPath := path.Join(stack.curPath[stack.size-1], fileName)
	exists, _ := utils.FileExists(absPath)
	return exists
}
