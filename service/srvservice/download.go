package srvservice

// func (stack *MyStack) ParseGetCmd(cmdStr string) ([]string, error) {
// 	s := strings.Fields(cmdStr) // 按空格（可多个）分割
// 	if len(s) < 2 {             // 暂时支持一个文件
// 		return nil, errors.New("invalid cmd")
// 	}
// 	if s[0] != "get" {
// 		return nil, errors.New("invalid cmd")
// 	}

// 	// 添加目录的绝对路径（后续改成前缀树的方式存储路径）
// 	var absPaths = make([]string, 0)
// 	for i := 1; i < len(s); i++ {
// 		absPath := stack.curPath[stack.size-1]
// 		if strings.HasPrefix(s[i], "/") {
// 			absPath += s[i]
// 		} else {
// 			pathSlice := strings.Split(s[i], "/")
// 			if pathSlice[0] == "." {
// 				absPath = strings.Replace(s[i], ".", absPath, 1)
// 			} else if pathSlice[0] == ".." {
// 				if stack.size > 1 {
// 					absPath = strings.Replace(s[i], "..", stack.curPath[stack.size-2], 1)
// 				} else {
// 					return nil, errors.New("invalid file path")
// 				}
// 			} else if strings.HasPrefix(strings.Split(s[i], "/")[0], ".") { // 3个点以上的情况
// 				return nil, errors.New("invalid file path")
// 			} else {
// 				absPath += "/" + s[i]
// 			}
// 		}
// 		if utils.IsDir(absPath) { // dir 暂时不支持
// 			return nil, errors.New("invalid file path")
// 		}
// 		absPaths = append(absPaths, absPath)
// 	}

// 	zlog.Debugf("absPath:%s", absPaths)

// 	return absPaths, nil
// }
