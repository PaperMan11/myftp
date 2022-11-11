package zserver

import (
	"context"
	"fmt"
	"myftp/global"
	"myftp/utils/snowflake"
	"myftp/zlog"
	"myftp/znet"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	Name      string // 服务器名称
	IpVersion string // tcp4 or other
	Ip        string
	Port      int
	MsgHandle znet.IMsgHandle   // 消息处理
	ConnMgr   znet.IConnManager // 连接管理

	// 关闭 wookpool
	ctx        context.Context
	cancelFunc context.CancelFunc

	// hook
	OnConnStart func(conn znet.IConnection)
	OnConnStop  func(conn znet.IConnection)
}

func NewServer() *Server {
	return &Server{
		Name:      global.GlobalObject.Name,
		IpVersion: "tcp4",
		Ip:        global.GlobalObject.Host,
		Port:      global.GlobalObject.TcpPort,
		MsgHandle: NewMsgHandle(),
		ConnMgr:   NewConnManager(),
	}
}

// 启动服务器方法
func (s *Server) Start() {
	zlog.Infof("%s Server Start\n", global.GlobalObject.Name)
	zlog.Infof("Listener <Ip:%s Port:%d>\n", global.GlobalObject.Host, global.GlobalObject.TcpPort)
	go func() {
		// 启动工作池
		s.ctx, s.cancelFunc = context.WithCancel(context.Background())
		s.MsgHandle.StartWorkerPool(s.ctx)

		addr, err := net.ResolveTCPAddr(s.IpVersion, fmt.Sprintf("%s:%d", s.Ip, s.Port))
		if err != nil {
			zlog.Errorf("net.ResolveTCPAddr err: %s", err.Error())
			return
		}

		lis, err := net.ListenTCP(s.IpVersion, addr)
		if err != nil {
			zlog.Errorf("net.ListenTCP err: %s", err.Error())
			return
		}

		var connID int64
		for {
			conn, err := lis.AcceptTCP()
			if err != nil {
				zlog.Errorf("lis.AcceptTCP err: %s", err.Error())
				continue
			}
			// 设置最大连接控制
			if s.ConnMgr.Len() >= global.GlobalObject.MaxConn {
				zlog.Error("conn overflow !!!")
				conn.Close()
				continue
			}
			// handle
			connID = snowflake.GetID()
			dealConn := NewConnection(s, conn, connID, s.MsgHandle)
			go dealConn.Start()
		}
	}()
}

// 停止服务器方法
func (s *Server) Stop() {
	// TODO: 清理所有连接和系统资源
	s.ConnMgr.ClearConn()
	s.cancelFunc()
	zlog.Infof("%s Server Stop\n", global.GlobalObject.Name)
}

// 开启业务方法
func (s *Server) Serve() {
	//TODO Server.Serve() 是否在启动服务的时候 还要处理其他的事情呢 可以在这里添加
	s.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-quit
	s.Stop()
}

// 路由功能: 给当前服务注册一个路由业务方法, 供客户端连接处理使用
func (s *Server) AddRouter(msgid uint32, router znet.IRouter) {
	s.MsgHandle.AddRouter(msgid, router)
}

// 获取连接管理
func (s *Server) GetConnMgr() znet.IConnManager {
	return s.ConnMgr
}

// 设置连接启动时的 Hook 函数
func (s *Server) SetOnConnStart(hook func(znet.IConnection)) {
	s.OnConnStart = hook
}

// 设置连接结束时的 Hook 函数
func (s *Server) SetOnConnStop(hook func(znet.IConnection)) {
	s.OnConnStop = hook
}

// 调用 OnConnStart 函数
func (s *Server) CallOnConnStart(conn znet.IConnection) {
	if s.OnConnStart != nil {
		zlog.Infof("call on connection start")
		s.OnConnStart(conn)
	}
}

// 调用 OnConnStop 函数
func (s *Server) CallOnConnStop(conn znet.IConnection) {
	if s.OnConnStop != nil {
		zlog.Infof("call on connection stop")
		s.OnConnStop(conn)
	}
}
