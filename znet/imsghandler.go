package znet

import "context"

// 消息管理抽象层 (多路由)
type IMsgHandle interface {
	DoMsgHandler(req IRequset)              // 马上以非阻塞方式处理消息
	AddRouter(msgID uint32, router IRouter) // 为消息添加具体的处理逻辑
	StartWorkerPool(context.Context)        // 开启工作池 (协程池)
	SendMsgToTaskQueue(req IRequset)        // 添加任务到任务队列
}
