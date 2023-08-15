package znet

import (
	"fmt"
	"sync"
	"zinx/ziface"
)

/*
连接管理模块
*/
type ConnManager struct {
	// 连接集合
	connections map[uint32]ziface.IConnection
	// 读写锁
	connLock sync.RWMutex
}

// 创建当前连接管理模块的方法
func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]ziface.IConnection),
	}
}

// 添加连接
func (cm *ConnManager) Add(conn ziface.IConnection) {
	// 保护共享资源map，加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()
	// 将conn连接添加到ConnManager中
	cm.connections[conn.GetConnID()] = conn
	fmt.Println("connection add to ConnManager successfully: conn num = ", cm.Len())
}

// 删除连接
func (cm *ConnManager) Remove(conn ziface.IConnection) {
	// 保护共享资源map，加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()
	// 删除连接信息
	delete(cm.connections, conn.GetConnID())
	fmt.Println("connection Remove ConnID=", conn.GetConnID(), " successfully: conn num = ", cm.Len())
}

// 根据connID获取连接
func (cm *ConnManager) Get(connID uint32) (ziface.IConnection, error) {
	// 保护共享资源map，加读锁
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()
	// 根据connID获取连接信息
	if conn, ok := cm.connections[connID]; ok {
		return conn, nil
	} else {
		return nil, fmt.Errorf("connection not FOUND!")
	}
}

// 得到当前连接总数
func (cm *ConnManager) Len() int {
	return len(cm.connections)
}

// 清除并终止所有连接
func (cm *ConnManager) ClearConn() {
	// 保护共享资源map，加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()
	// 删除conn并停止conn的工作
	for connID, conn := range cm.connections {
		// 停止
		conn.Stop()
		// 删除
		delete(cm.connections, connID)
	}
}
