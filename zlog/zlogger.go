package zlog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// 日志全部方法

const (
	LOG_MAX_BUFF = 1024 * 1024
)

// 日志头部信息标记
const (
	BitDate         = 1 << iota                            // 日期标记位 2022/04/28
	BitTime                                                // 时间标记位 13:57:54
	BitMicroSeconds                                        // 微秒级标记位 13:57:54.132151
	BitLongFile                                            // 完整文件名 /root/zinx/zlog/server.log
	BitShortFile                                           // 文件名 server.log
	BitLevel                                               // 当前日志级别 0(Debug) 1(Info) 2(Warn) 3(Error) 4(Panic) 5(Fatal)
	BitStdFlag      = BitDate | BitTime                    // 标准头部日志格式
	BitDefault      = BitLevel | BitShortFile | BitStdFlag // 默认日志头部格式
)

// 日志级别
const (
	LogDebug = iota
	LogInfo
	LogWarn
	LogError
	LogPanic
	LogFatal
)

// 日志级别对应的信息
var levels = []string{
	"[DEBUG]",
	"[INFO]",
	"[WARN]",
	"[ERROR]",
	"[PANIC]",
	"[FATAL]",
}

type ZinxLoggor struct {
	mu         sync.Mutex   // 确保多线程读写
	prefix     string       // 每行日志信息前缀
	flag       int          // 日志标记
	out        io.Writer    // 日志输出的文件描述符
	buf        bytes.Buffer // 输出缓冲区
	file       *os.File     // 当前文件绑定的文件
	debugClose bool         // 是否打印调试信息
	callDepth  int          // 获取日志文件和代码上述的 runtime.call 函数调用信息
}

// 创建一个日志
func NewZinxLog(prefix string, flag int) *ZinxLoggor {
	// 默认打开 debug calldepth 深度为2层
	zlog := &ZinxLoggor{
		out:        os.Stderr, // 默认标准出错
		prefix:     prefix,
		flag:       flag,
		file:       nil,
		debugClose: false,
		callDepth:  2,
	}
	// 设置 log 对象回收资源 (不设置也行)
	runtime.SetFinalizer(zlog, CleanZinxLog)
	return zlog
}

// 日志回收
func CleanZinxLog(log *ZinxLoggor) {
	log.closeFile()
}

// 制作当前日志数据格式信息
// <prefix>2022/04/28 15:13:2.123456 [ERROR]fileName.go/12: err
func (log *ZinxLoggor) formatHeader(buf *bytes.Buffer, t time.Time, file string, line int, level int) {
	// 如果当前前缀不为空，写前缀
	if log.prefix != "" {
		buf.WriteByte('<')
		buf.WriteString(log.prefix)
		buf.WriteByte('>')
	}
	// 已经设置了时间相关的标识位，那么需要添加时间信息日志头部
	if log.flag&(BitDate|BitTime|BitMicroSeconds) != 0 {
		// 日期标记位
		if log.flag&BitDate != 0 {
			year, month, day := t.Date()
			intToWidth(buf, year)       // 转成相应的字符串
			buf.WriteByte('/')          // 2022/
			intToWidth(buf, int(month)) // 转成相应的字符串
			buf.WriteByte('/')          // 2022/04
			intToWidth(buf, day)        // 转成相应的字符串
			buf.WriteByte(' ')          // 2022/04/28
		}
		// 时间位标记
		if log.flag&(BitTime|BitMicroSeconds) != 0 {
			hour, min, sec := t.Clock()
			intToWidth(buf, hour)
			buf.WriteByte(':') // 14:
			intToWidth(buf, min)
			buf.WriteByte(':')   // 14:59:
			intToWidth(buf, sec) // 14:59:11
			if log.flag&BitMicroSeconds != 0 {
				buf.WriteByte('.') // 14:59:11.123456
				intToWidth(buf, t.Nanosecond()/1e3)
			}
			buf.WriteByte(' ')
		}
		// 日志级别
		if log.flag&BitLevel != 0 {
			buf.WriteString(levels[level])
		}
		// 日志当前代码调用文件名名称标记
		if log.flag&(BitLongFile|BitShortFile) != 0 {
			// 打印短文件名
			if log.flag&BitShortFile != 0 {
				short := file
				for i := len(file) - 1; i > 0; i-- {
					if file[i] == '/' {
						// 找到最后一个 / 之后的字符串
						short = file[i+1:]
						break
					}
				}
				file = short
			}
			buf.WriteString(file)
			buf.WriteByte(':')
			intToWidth(buf, line) // 行数
			buf.WriteString(": ")
		}
	}
}

// 输出日志文件
// @level: 日志级别
// @s: 日志内容
func (log *ZinxLoggor) OutPut(level int, s string) error {
	now := time.Now()
	var file string // 文件名
	var line int    // 当前代码行号
	log.mu.Lock()
	defer log.mu.Unlock()

	if log.flag&(BitShortFile|BitLongFile) != 0 {
		log.mu.Unlock()
		var ok bool
		// 得到当前调用者的文件名称和执行代码行数
		_, file, line, ok = runtime.Caller(log.callDepth)
		if !ok {
			file = "unknownfile"
			line = 0
		}
		log.mu.Lock()
	}

	// 清空 buf
	log.buf.Reset()
	// 写日志头
	log.formatHeader(&log.buf, now, file, line, level)
	// 写日志内容
	log.buf.WriteString(s)
	// 补充回车
	if len(s) > 0 && s[len(s)-1] != '\n' {
		log.buf.WriteByte('\n')
	}
	// 将 buf 输出到 io
	_, err := log.out.Write(log.buf.Bytes())
	return err
}

// debug
func (log *ZinxLoggor) Debugf(format string, v ...interface{}) {
	if log.debugClose {
		return
	}
	log.OutPut(LogDebug, fmt.Sprintf(format, v...))
}

func (log *ZinxLoggor) Debug(v ...interface{}) {
	if log.debugClose {
		return
	}
	log.OutPut(LogDebug, fmt.Sprint(v...))
}

// info
func (log *ZinxLoggor) Infof(format string, v ...interface{}) {
	log.OutPut(LogInfo, fmt.Sprintf(format, v...))
}

func (log *ZinxLoggor) Info(v ...interface{}) {
	log.OutPut(LogInfo, fmt.Sprint(v...))
}

// warn
func (log *ZinxLoggor) Warnf(format string, v ...interface{}) {
	log.OutPut(LogWarn, fmt.Sprintf(format, v...))
}

func (log *ZinxLoggor) Warn(v ...interface{}) {
	log.OutPut(LogWarn, fmt.Sprint(v...))
}

// error
func (log *ZinxLoggor) Errorf(format string, v ...interface{}) {
	log.OutPut(LogError, fmt.Sprintf(format, v...))
}

func (log *ZinxLoggor) Error(v ...interface{}) {
	log.OutPut(LogError, fmt.Sprint(v...))
}

// panic
func (log *ZinxLoggor) Panicf(format string, v ...interface{}) {
	log.OutPut(LogPanic, fmt.Sprintf(format, v...))
}

func (log *ZinxLoggor) Panic(v ...interface{}) {
	log.OutPut(LogPanic, fmt.Sprint(v...))
}

// Fatal
func (log *ZinxLoggor) Fatalf(format string, v ...interface{}) {
	log.OutPut(LogFatal, fmt.Sprintf(format, v...))
}

func (log *ZinxLoggor) Fatal(v ...interface{}) {
	log.OutPut(LogFatal, fmt.Sprint(v...))
}

// stack
func (log *ZinxLoggor) Stack(v ...interface{}) {
	s := fmt.Sprint(v...)
	s += "\n"
	buf := make([]byte, LOG_MAX_BUFF)
	n := runtime.Stack(buf, true) // 得到当前堆栈信息
	s += string(buf[:n])
	s += "\n"
	log.OutPut(LogError, s)
}

// 获取当前日志标记
func (log *ZinxLoggor) Flags() int {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.flag
}

// 重置当前日志标记
func (log *ZinxLoggor) ResetFlags(flag int) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.flag = flag
}

// 添加flag标记
func (log *ZinxLoggor) AddFlags(flag int) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.flag |= flag
}

// 设置日志前缀
func (log *ZinxLoggor) SetPrefix(prefix string) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.prefix = prefix
}

// 设置日志文件输出
func (log *ZinxLoggor) SetLogFile(fileDir string, fileName string) {
	var file *os.File
	// 创建日志文件夹
	mkdirLog(fileDir)

	fullPath := fileDir + "/" + fileName
	if log.checkFileExist(fullPath) {
		file, _ = os.OpenFile(fullPath, os.O_APPEND|os.O_RDWR, 0644)
	} else {
		// 文件不存在
		file, _ = os.OpenFile(fullPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	}
	log.mu.Lock()
	defer log.mu.Unlock()
	// 关闭之前绑定的文件
	log.closeFile()
	log.file = file
	log.out = file
}

// 关闭日志绑定文件
func (log *ZinxLoggor) closeFile() {
	if log.file != nil {
		log.file.Close()
		log.file = nil
		log.out = os.Stderr
	}
}

// debug open/close
func (log *ZinxLoggor) CloseDebug() {
	log.debugClose = true
}

func (log *ZinxLoggor) OpenDebug() {
	log.debugClose = false
}

// 一些工具方法
// 判断日志是否存在
func (log *ZinxLoggor) checkFileExist(fileName string) bool {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return false
	}
	return true
}

// 创建日志目录
func mkdirLog(dir string) error {
	_, err := os.Stat(dir)
	if !(err == nil || os.IsExist(err)) { // 目录不存在
		err = os.MkdirAll(dir, 0775)
	}
	return err
}

// 将一个整形转成对应的字符串
func intToWidth(buf *bytes.Buffer, i int) {
	buf.WriteString(strconv.Itoa(i))
}
