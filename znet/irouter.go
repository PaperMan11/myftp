package znet

// 路由接口
type IRouter interface {
	PreHandle(req IRequset)  // 处理 conn 业务之前的handle
	Handle(req IRequset)     // 处理 conn 业务的handle
	PostHandle(req IRequset) // 处理 conn 业务之后的handle
}
