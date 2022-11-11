package srvservice

import (
	"errors"
	"fmt"
	"myftp/utils"
	"myftp/zlog"
	"strings"
)

func (stack *MyStack) ParseCdCmd(cmdStr string) error {
	str := strings.Replace(cmdStr, "\"", "", -1)
	s := strings.Fields(str) // 按空格（可多个）分割
	fmt.Println(s)
	if len(s) == 0 {
		return errors.New("null cmd")
	}

	if s[0] != "cd" || len(s) > 2 {
		return errors.New("invalid cmd")
	}

	if len(s) == 1 {
		// 输出当前目录中的文件
		return nil
	}

	// 添加目录的绝对路径（后续改成前缀树的方式存储）
	absPath := stack.curPath[stack.size-1]
	if strings.HasPrefix(s[1], "/") {
		absPath += s[1]
		stack.push(absPath)
	} else {
		pathSlice := strings.Split(s[1], "/")
		if pathSlice[0] == "." { // 暂时只有一层 eg cd .
			absPath = strings.Replace(s[1], ".", absPath, 1)
		} else if pathSlice[0] == ".." {
			if stack.size > 1 { // 暂时只有一层 eg cd ..
				absPath = strings.Replace(s[1], "..", stack.curPath[stack.size-2], 1)
				stack.pop()
			} else {
				return errors.New("invalid path")
			}
		} else if strings.HasPrefix(strings.Split(s[1], "/")[0], ".") { // 3个点以上的情况
			return errors.New("invalid path")
		} else {
			absPath += "/" + s[1]
			stack.push(absPath)
		}
	}
	zlog.Debugf("absPath:%s", absPath)

	// 正则匹配
	if ok := utils.IsValidDir(absPath); !ok {
		zlog.Error("invalid path")
		stack.pop()
		return errors.New("invalid path")
	}

	if !utils.IsDir(absPath) {
		stack.pop()
		return errors.New("is not directory")
	}
	return nil
}
