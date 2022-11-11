package zserver

import (
	"context"
	"myftp/global"
	"myftp/zlog"
	"myftp/znet"
)

// 工厂模式
type MsgHandle struct {
	Apis           map[uint32]znet.IRouter // 存放每个 msgID 对应的 handler
	WorkerPoolSize uint32                  // 工作池大小 (协程池)
	TaskQueue      []chan znet.IRequset    // 任务队列

}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]znet.IRouter),
		WorkerPoolSize: global.GlobalObject.WorkerPoolSize,
		TaskQueue:      make([]chan znet.IRequset, global.GlobalObject.WorkerPoolSize),
	}
}

// 马上以非阻塞方式处理消息
func (mh *MsgHandle) DoMsgHandler(req znet.IRequset) {
	handlerFunc, ok := mh.Apis[req.GetMsgID()]
	if !ok {
		zlog.Errorf("api msgID: [%d] is not found", req.GetMsgID())
		return
	}

	// do
	handlerFunc.PreHandle(req)
	handlerFunc.Handle(req)
	handlerFunc.PostHandle(req)
}

// 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgID uint32, router znet.IRouter) {
	if _, ok := mh.Apis[msgID]; ok {
		zlog.Infof("api msgID: [%d] exsits", msgID)
		return
	}
	mh.Apis[msgID] = router
	zlog.Infof("Add api msgID: [%d]", msgID)
}

// 开启工作池 (协程池)
func (mh *MsgHandle) StartWorkerPool(ctx context.Context) {
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		// 一个 worker 启动
		// 给任务队列开辟资源
		mh.TaskQueue[i] = make(chan znet.IRequset, global.GlobalObject.MaxWokerTaskLen)
		go mh.startOneWorker(ctx, i, mh.TaskQueue[i])
	}
}

// 一个 worker 工作流程
func (mh *MsgHandle) startOneWorker(ctx context.Context, workerID int, taskQueue chan znet.IRequset) {
	zlog.Infof("Worker ID:[%d] is started", workerID)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			req := <-taskQueue
			mh.DoMsgHandler(req)
		}
	}
}

// 添加任务到任务队列
func (mh *MsgHandle) SendMsgToTaskQueue(req znet.IRequset) {
	// 根据 ConnID 来分配当前连接请求由哪个 worker 负责处理 (eg. 0~9)
	workerID := req.GetConnection().GetConnID() % int64(mh.WorkerPoolSize)

	zlog.Infof("ConnID:[%d], Req MsgID:[%d] --> WorkerID:[%d]",
		req.GetConnection().GetConnID(), req.GetMsgID(), workerID)

	mh.TaskQueue[workerID] <- req // 放入任务队列
}
