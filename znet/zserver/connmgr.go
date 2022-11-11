package zserver

import (
	"errors"
	"myftp/zlog"
	"myftp/znet"
	"sync"
)

// 连接管理
type ConnManager struct {
	connections map[int64]znet.IConnection // 存储连接信息
	connLock    sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[int64]znet.IConnection),
	}
}

// 添加连接
func (connMgr *ConnManager) Add(conn znet.IConnection) {
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()
	connID := conn.GetConnID()
	connMgr.connections[connID] = conn
	zlog.Infof("conn [%d] add success", connID)
}

// 删除连接
func (connMgr *ConnManager) Remove(conn znet.IConnection) {
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()
	delete(connMgr.connections, conn.GetConnID())
	zlog.Infof("conn [%d] delete success", conn.GetConnID())
}

// 根据 connID 获取连接
func (connMgr *ConnManager) Get(connID int64) (znet.IConnection, error) {
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()
	if conn, ok := connMgr.connections[connID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("conn not found")
	}
}

// 获取当前连接数
func (connMgr *ConnManager) Len() int {
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()
	return len(connMgr.connections)
}

// 删除并停止所有链接
func (connMgr *ConnManager) ClearConn() {
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()
	for connID, conn := range connMgr.connections {
		conn.Stop()
		delete(connMgr.connections, connID)
	}
	zlog.Info("clean All connections success")
}
