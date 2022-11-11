package model

// Server response
/*****************upload*****************/
type LsResp struct {
	Code int
	Msg  string
	Data []string
}

type SliceSeq struct {
	Slices []int // 需要重传的分片号
}

type DownloadFileResp struct {
	Code int    // 0 正在传 1 结束 2 错误
	Msg  string // 错误信息
	Data []byte // 真实数据
}
