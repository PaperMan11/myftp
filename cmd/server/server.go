package main

import (
	"myftp/router"
	"myftp/utils/snowflake"
	"myftp/znet/zserver"
)

func main() {
	snowflake.Init("2022-11-04", 1)
	// 服务端测试
	s := zserver.NewServer()
	// 注册路由
	s.AddRouter(0, new(router.LsCommand))
	s.AddRouter(1, new(router.Upload))
	s.AddRouter(2, new(router.UploadingStat))
	s.AddRouter(3, new(router.CreateUploadDir))
	s.AddRouter(4, new(router.UploadBySlice))
	s.AddRouter(5, new(router.MergeSliceFiles))
	s.AddRouter(6, new(router.CdCommand))
	s.AddRouter(7, new(router.Download))

	// 设置 hook 函数

	// 开启服务
	s.Serve()
}
