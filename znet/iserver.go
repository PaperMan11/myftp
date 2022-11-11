package znet

// Server Interface
type IServer interface {
	Start()                                 // 启动服务器方法
	Stop()                                  // 停止服务器方法
	Serve()                                 // 开启业务方法
	AddRouter(msgid uint32, router IRouter) // 路由功能: 给当前服务注册一个路由业务方法, 供客户端连接处理使用
	GetConnMgr() IConnManager               // 获取连接管理
	SetOnConnStart(func(IConnection))       // 设置连接启动时的 Hook 函数
	SetOnConnStop(func(IConnection))        // 设置连接结束时的 Hook 函数
	CallOnConnStart(conn IConnection)       // 调用 OnConnStart 函数
	CallOnConnStop(conn IConnection)        // 调用 OnConnStop 函数
}
