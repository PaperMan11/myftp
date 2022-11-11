package srvservice

import (
	"errors"
	"fmt"
	"myftp/utils"
	"myftp/zlog"
	"strings"
)

// eg. ls file -l
func (stack *MyStack) ParseLsCmd(cmdStr string) ([]string, error) {
	s := strings.Fields(cmdStr) // 按空格（可多个）分割

	if len(s) == 0 {
		return nil, errors.New("null cmd")
	}

	if !strings.Contains(s[0], "ls") || len(s) > 3 {
		return nil, errors.New("invalid cmd")
	}

	if len(s) == 1 {
		// 输出当前目录中的文件
		return utils.ListDir(stack.curPath[stack.size-1], "")
	}

	// 添加目录的绝对路径（后续改成前缀树的方式存储路径）
	absPath := stack.curPath[stack.size-1]
	if strings.HasPrefix(s[1], "/") {
		absPath += s[1]
	} else {
		pathSlice := strings.Split(s[1], "/")
		if pathSlice[0] == "." {
			absPath = strings.Replace(s[1], ".", absPath, 1)
		} else if pathSlice[0] == ".." {
			if stack.size > 1 {
				absPath = strings.Replace(s[1], "..", stack.curPath[stack.size-2], 1)
			} else {
				return nil, errors.New("invalid path")
			}
		} else if strings.HasPrefix(strings.Split(s[1], "/")[0], ".") { // 3个点以上的情况
			return nil, errors.New("invalid path")
		} else {
			absPath += "/" + s[1]
		}
	}
	zlog.Debugf("absPath:%s", absPath)

	if utils.IsDir(absPath) { // dir
		if len(s) == 3 && !strings.Contains(s[0], "-l") {
			fmt.Println(2)
			return utils.GetMutiFileStat(absPath)
		}
		fmt.Println(3)
		return utils.ListDir(absPath, "")
	} else {
		if len(s) == 3 && !strings.Contains(s[0], "-l") {
			fmt.Println(1)
			fileStat, err := utils.GetFileStat(absPath)
			return []string{fileStat}, err // file
		}
		fmt.Println(4)
		return []string{s[1]}, nil
	}
}
