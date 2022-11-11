package znet

// 连接管理抽象层
type IConnManager interface {
	Add(conn IConnection)                  // 添加连接
	Remove(conn IConnection)               // 删除连接
	Get(connID int64) (IConnection, error) // 根据 connID 获取连接
	Len() int                              // 获取当前连接数
	ClearConn()                            // 删除并停止所有链接
}
