package model

import "time"

// 定义常量
const (
	SmallFileSize            = 1024 * 1024     // 小文件大小
	SliceBytes               = 1024 * 1024 * 3 // 分片大小
	UploadRetryChannelNum    = 100             // 上传的重试通道队列大小
	DownloadRetryChannelNum  = 100             // 下载的重试通道队列大小
	UploadTimeout            = 300             // 上传超时时间，单位秒
	DownloadTimeout          = 300             // 上传超时时间，单位秒
	UpGoroutineMaxNumPerFile = 10              // 每个上传文件开启的goroutine最大数量
	DpGoroutineMaxNumPerFile = 10              // 每个下载文件开启的goroutine最大数量
	StoreDir                 = "/root/tan/download"
)

/*****************upload*****************/

// SingleFileInfo 上传单个文件结构体
type SingleFileInfo struct {
	Filesize int64  // 文件大小（字节单位）
	Filename string // 文件名称
	Md5sum   string // 文件md5值
	Data     []byte
}

// FilePart 文件分片结构
type FilePart struct {
	Fid   string // 操作文件ID，随机生成的UUID
	Index int    // 文件切片序号
	Data  []byte // 分片数据
}

// ClientFileMetadata 客户端传来的文件元数据结构
type ClientFileMetadata struct {
	Fid        string    // 操作文件ID，随机生成的UUID
	Filesize   int64     // 文件大小（字节单位）
	Filename   string    // 文件名称
	SliceNum   int       // 切片数量
	Md5sum     string    // 文件md5值
	ModifyTime time.Time // 文件修改时间
}

// ServerFileMetadata 服务端保存的文件元数据结构
type ServerFileMetadata struct {
	ClientFileMetadata        // 隐式嵌套
	State              string // 文件状态，目前有uploading、downloading和active（没用到）
}

// UploadingFileReq 文件上传前查询
type UploadingFileReq struct {
	Fid      string
	Filename string
}

/*****************download*****************/

// // FileInfo 文件列表单元结构
// type FileInfo struct {
// 	Filename string // 文件名
// 	Filesize int64  // 文件大小
// 	Filetype string // 文件类型（目前有普通文件和切片文件两种）
// }

// // ListFileInfos 文件列表结构
// type ListFileInfos struct {
// 	Files []FileInfo
// }

// DownloadFileReq 下载文件请求
type DownloadFileReq struct {
	FileName string
	Index    int // 断点续传
}
