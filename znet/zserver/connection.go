package zserver

import (
	"errors"
	"io"
	"myftp/global"
	"myftp/service/srvservice"
	"myftp/zlog"
	"myftp/znet"
	"net"
	"sync"
)

// 连接对象
type Connection struct {
	TcpServer    znet.IServer           // 当前连接属于哪个 Server
	Conn         *net.TCPConn           // 连接套接字
	ConnID       int64                  // 连接ID
	isClosed     bool                   // 是否关闭连接
	MsgHandler   znet.IMsgHandle        // 消息管理MsgId和对应处理方法的消息管理模块
	ExitBuffChan chan struct{}          // 告知连接关闭的channel
	msgChan      chan []byte            // 用于读、写两个 groutine 之间的消息通信
	msgBuffChan  chan []byte            // 有缓冲
	property     map[string]interface{} // 连接属性
	propertyLock sync.RWMutex           // 锁
	MyStack      *srvservice.MyStack    // ftp命令操作对象，保存当前用户目录位置
}

func NewConnection(server znet.IServer, conn *net.TCPConn, connId int64, msgHandler znet.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:    server,
		Conn:         conn,
		ConnID:       connId,
		isClosed:     false,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan struct{}, 1),
		msgChan:      make(chan []byte),
		msgBuffChan:  make(chan []byte, global.GlobalObject.MaxMsgChanLen),
		property:     make(map[string]interface{}),
		MyStack:      srvservice.NewStack("/root/tan", 0, 20),
	}
	// 将连接放到 connMgr 统一管理
	c.TcpServer.GetConnMgr().Add(c)
	return c
}

// 启动连接
func (conn *Connection) Start() {
	conn.TcpServer.CallOnConnStart(conn) // hook

	// 读写分离
	go conn.startReader()
	go conn.startWriter()

	<-conn.ExitBuffChan
}

func (conn *Connection) startReader() {
	zlog.Debugf("ConnID: %d Reader running...", conn.ConnID)
	defer zlog.Debugf("ConnID: %d Reader exit...", conn.ConnID)
	defer conn.Stop()

	for {
		// read client package
		pkg := NewDataPack()

		// read head
		headData := make([]byte, pkg.GetHeadLen())
		if _, err := io.ReadFull(conn.GetTCPConnection(), headData); err != nil {
			zlog.Infof("read msg head err: %s", err)
			conn.ExitBuffChan <- struct{}{}
			break
		}

		// unpack: get msgID msgLen
		msg, err := pkg.Unpack(headData)
		if err != nil {
			zlog.Errorf("Unpak err: %s", err)
			conn.ExitBuffChan <- struct{}{}
			break
		}

		// get msg data
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(conn.GetTCPConnection(), data); err != nil {
				zlog.Errorf("read msg data err: %s", err)
				conn.ExitBuffChan <- struct{}{}
				break
			}
		}
		msg.SetData(data) // 将数据放入msg中

		// set client request
		req := Request{
			conn: conn,
			msg:  msg,
		}

		if global.GlobalObject.WorkerPoolSize > 0 {
			conn.MsgHandler.SendMsgToTaskQueue(&req) // 将请求加入任务队列，对应的协程进行处理
		} else { // 没有开启协程池
			go conn.MsgHandler.DoMsgHandler(&req)
		}
	}
}

func (conn *Connection) startWriter() {
	zlog.Debugf("ConnID: %d Writer running...", conn.ConnID)
	defer zlog.Debugf("ConnID: %d Writer exit...", conn.ConnID)

	for {
		select {
		case data, ok := <-conn.msgChan:
			if ok {
				if _, err := conn.Conn.Write(data); err != nil {
					zlog.Errorf("send data err: %s", err)
					return
				}
			} else {
				zlog.Error("msgChan closed")
			}
		case data, ok := <-conn.msgBuffChan:
			if ok {
				if _, err := conn.Conn.Write(data); err != nil {
					zlog.Errorf("send data err: %s", err)
					return
				}
			} else {
				zlog.Error("msgBuffChan closed")
			}
		case <-conn.ExitBuffChan:
			// 放到这里close 防止管道关闭了还在写操作
			close(conn.msgBuffChan)
			close(conn.msgChan)
			return
		}
	}
}

// 关闭连接
func (conn *Connection) Stop() {
	zlog.Infof("Conn: %d Stop", conn.ConnID)
	if conn.isClosed {
		return
	}
	conn.isClosed = true

	conn.TcpServer.CallOnConnStop(conn) // hook

	// 关闭当前连接
	conn.ExitBuffChan <- struct{}{}
	conn.Conn.Close()

	// 将连接从连接管理对象中删除
	conn.TcpServer.GetConnMgr().Remove(conn)

	// 关闭所有管道
	close(conn.ExitBuffChan)
}

// 获取连接套接字
func (conn *Connection) GetTCPConnection() *net.TCPConn {
	return conn.Conn
}

// 获取连接ID
func (conn *Connection) GetConnID() int64 {
	return conn.ConnID
}

// 获取远程客户端地址
func (conn *Connection) RemoteAddr() net.Addr {
	return conn.Conn.RemoteAddr()
}

// new add
func (conn *Connection) GetMyStack() *srvservice.MyStack {
	return conn.MyStack
}

// 发数据 (无缓冲)
func (conn *Connection) SendMsg(msgID uint32, data []byte) error {
	if conn.isClosed {
		return errors.New("Connection closed")
	}
	pkg := NewDataPack()
	msg, err := pkg.Pack(NewMsgPackage(msgID, data)) // 封包
	if err != nil {
		zlog.Errorf("msg %d pack err: %s", msgID, err)
		return errors.New("pack msg error")
	}

	// recive client
	conn.msgChan <- msg
	return nil
}

// 发数据 (有缓冲)
func (conn *Connection) SendBuffMsg(msgID uint32, data []byte) error {
	if conn.isClosed {
		return errors.New("Connection closed")
	}
	pkg := NewDataPack()
	msg, err := pkg.Pack(NewMsgPackage(msgID, data)) // 封包
	if err != nil {
		zlog.Errorf("msg %d pack err: %s", msgID, err)
		return errors.New("pack msg error")
	}

	// recive client
	conn.msgBuffChan <- msg
	return nil
}

// 设置连接属性
func (conn *Connection) SetProperty(key string, value interface{}) {
	conn.propertyLock.Lock()
	defer conn.propertyLock.Unlock()
	conn.property[key] = value
}

// 获取连接属性
func (conn *Connection) GetProperty(key string) (interface{}, error) {
	conn.propertyLock.RLock()
	defer conn.propertyLock.RUnlock()
	if value, ok := conn.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

// 删除连接属性
func (conn *Connection) RemoveProperty(key string) {
	conn.propertyLock.Lock()
	defer conn.propertyLock.Unlock()
	delete(conn.property, key)
}
