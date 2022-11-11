package utils

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// IsValidDir 解析输入的目录路径是否合法
func IsValidDir(str string) bool {
	// 正则匹配
	reg := regexp.MustCompile(`^\/(\w+\/?)+$`)
	if reg == nil {
		return false
	}
	res := reg.FindIndex([]byte(str))
	if len(res) == 0 || len(res) != 2 {
		return false
	}

	return true
}

// IsDir 是否为目录
func IsDir(path string) bool {
	file, err := os.Stat(path)
	if err != nil {
		return false
	}
	return file.IsDir()
}

func FileSize(path string) (int64, error) {
	file, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	size := file.Size()
	return size, nil
}

// FileExists 判断文件或者文件夹是否存在
// 一般判断第一个参数即可，第二个参数可以忽略，或者严谨一些，把err日志记录起来
func FileExists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err == nil { //文件或者文件夹存在
		return true, nil
	}
	if os.IsNotExist(err) { //不存在
		return false, nil
	}
	return false, err //不存在，这里的err可以查到具体的错误信息
}

// ListDir 获取指定路径下的所有文件，只搜索当前路径，不进入下一级目录
// 可匹配后缀过滤（suffix为空则不过滤）
func ListDir(dir, suffix string) (files []string, err error) {
	files = make([]string, 0)

	_dir, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	suffix = strings.ToLower(suffix) // 匹配后缀

	for _, _file := range _dir {
		if _file.IsDir() {
			// files = append(files, path.Join(dir, _file.Name()))
			files = append(files, _file.Name())
			continue
		}
		if len(suffix) == 0 || strings.HasSuffix(strings.ToLower(_file.Name()), suffix) {
			// files = append(files, path.Join(dir, _file.Name()))
			files = append(files, _file.Name())
		}
	}
	return files, nil
}

// listDirAbs 获取指定路径下的所有文件（绝对路径），只搜索当前路径，不进入下一级目录
// 可匹配后缀过滤（suffix为空则不过滤）
func listDirAbs(dir, suffix string) (files []string, err error) {
	files = make([]string, 0)

	_dir, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	suffix = strings.ToLower(suffix) // 匹配后缀

	for _, _file := range _dir {
		if _file.IsDir() {
			files = append(files, path.Join(dir, _file.Name()))
			// files = append(files, _file.Name())
			continue
		}
		if len(suffix) == 0 || strings.HasSuffix(strings.ToLower(_file.Name()), suffix) {
			files = append(files, path.Join(dir, _file.Name()))
			// files = append(files, _file.Name())
		}
	}
	return files, nil
}

// GetMutiFileStat 获取目录下所有文件的属性
func GetMutiFileStat(dir string) ([]string, error) {
	var (
		listStat = make([]string, 0)
		err      error
	)

	files, err := listDirAbs(dir, "")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		f, err := GetFileStat(file)
		if err != nil {
			return nil, err
		}
		listStat = append(listStat, f)
	}

	return listStat, err
}

// GetFileStat 获取单个文件的属性
func GetFileStat(fileName string) (string, error) {
	info, err := os.Stat(fileName)
	if err != nil {
		return "", errors.New("not found")
	}

	var str strings.Builder
	str.WriteString(info.Mode().String())
	str.WriteString("\t")
	str.WriteString(strconv.Itoa(int(info.Size())))
	str.WriteString("\t")
	str.WriteString(info.ModTime().String())
	str.WriteString("\t")
	str.WriteString(info.Name())

	return str.String(), nil
}
