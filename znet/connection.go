package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"zinx/utils"
	"zinx/ziface"
)

/*
链接模块
*/
type connection struct {
	//当前Conn属于哪个Server
	TcpServer ziface.IServer
	// 当前连接的 socket TCP 套接字
	Conn *net.TCPConn
	// 当前连接的 ID
	ConnID uint32
	// 当前连接的状态
	isClosed bool
	// 告知当前连接已经退出/停止的 channel
	ExitChan chan bool
	//消息管理MsgID和对应处理方法的消息管理模块
	MsgHandle ziface.IMsgHandler
	// 无缓冲管道，用于读、写两个 goroutine 之间的消息通信
	msgChan chan []byte
	//链接属性集合
	property map[string]interface{}
	//保护链接属性的锁
	propertyLock sync.RWMutex
}

// NewConnection 初始化连接模块的方法
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandle ziface.IMsgHandler) *connection {
	c := &connection{
		TcpServer: server,
		Conn:      conn,
		ConnID:    connID,
		MsgHandle: msgHandle,
		isClosed:  false,
		ExitChan:  make(chan bool, 1),
		msgChan:   make(chan []byte),
		property:  make(map[string]interface{}),
	}
	// 将 conn 加入到 ConnManager 中
	c.TcpServer.GetConnMgr().Add(c)
	return c
}

func (c *connection) startReader() {
	fmt.Println("Reader Goroutine is running...")
	defer fmt.Println("connID = ", c.ConnID, "Reader is exit, remote addr is ", c.RemoteAddr().String())
	defer c.Stop()
	for {
		//创建一个拆包解包对象
		dp := NewDatePack()
		//读取客户端的Msg Head 二进制流 8 字节
		headData := make([]byte, dp.GetHeadLen())
		_, err := io.ReadFull(c.GetTCPConnection(), headData) //ReadFull 会把msg填充满为止
		if err != nil {
			fmt.Println("read msg head error ", err)
			break
		}
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error ", err)
			break
		}
		//根据dataLen 再次读取Data，放在msg.Data中
		var data []byte
		if msg.GetMsgLen() > 0 {
			data = make([]byte, msg.GetMsgLen())
			_, err := io.ReadFull(c.GetTCPConnection(), data)
			if err != nil {
				fmt.Println("read msg data error ", err)
				break
			}
		}
		//将data封装到msg中
		msg.SetData(data)
		//得到当前 conn 数据的 Request 请求数据
		req := Request{
			conn: c,
			msg:  msg,
		}
		if utils.GlobalObject.WorkerPoolSize > 0 {
			//已经开启了工作池机制，将消息发送给Worker工作池处理即可
			c.MsgHandle.SendMsgToTaskQueue(&req)
		} else {
			//从路由中，找到注册绑定的Conn对应的router调用
			//根据绑定好的MsgID找到对应处理api业务执行
			go c.MsgHandle.DoMsgHandler(&req)
		}
	}
}

// 提供一个SendMsg方法，将我们要发送给客户端的数据，先进行封包，再发送
func (c *connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("connection closed when send msg")
	}
	//将data进行封包
	dp := NewDatePack()
	binaryMsg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}
	//将数据发送给客户端
	c.msgChan <- binaryMsg
	return nil
}

// StartWriter 写消息的 goroutine， 用户将数据发送给客户端
// 秒啊，太牛逼了，这个写数据的 goroutine 居然是一个死循环，不停的从 channel 中取数据，然后写给客户端
func (c *connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")
	// 不断的阻塞的等待 channel 的消息，进行写给客户端
	for {
		select {
		case data := <-c.msgChan:
			// 有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send data error:, ", err, " Conn Writer exit")
				return
			}
		case <-c.ExitChan:
			// 代表 reader 已经退出，此时 writer 也要退出
			return
		}
	}
}

// 启动连接，让当前连接开始工作
func (c *connection) Start() {
	fmt.Println("Conn Start()...ConnID = ", c.ConnID)
	// 启动从当前连接的读数据业务
	go c.startReader()
	// 启动从当前连接写数据业务
	go c.StartWriter()
	// 按照开发者传递进来的 创建连接之后需要调用的处理业务，执行对应的 Hook 函数
	c.TcpServer.CallOnConnStart(c)
}

// 停止连接，结束当前连接状态 M
func (c *connection) Stop() {
	fmt.Println("Conn Stop()...ConnID = ", c.ConnID)
	// 如果当前连接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true
	// 调用开发者注册的 销毁连接之前需要执行的业务 Hook 函数
	c.TcpServer.CallOnConnStop(c)
	// 关闭 socket 连接
	c.Conn.Close()
	// 关闭 Writer goroutine
	c.ExitChan <- true
	// 将当前连接从 ConnMgr 中摘除掉
	c.TcpServer.GetConnMgr().Remove(c)
	close(c.ExitChan)
	close(c.msgChan)
}

// 获取当前连接绑定的 socket conn
func (c *connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// 获取当前连接模块的连接 ID
func (c *connection) GetConnID() uint32 {
	return c.ConnID
}

// 获取远程客户端的 TCP 状态 IP port
func (c *connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 设置链接属性
func (c *connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	//添加一个链接属性
	c.property[key] = value
}

// 获取链接属性
func (c *connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	//读取属性
	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

// 移除链接属性
func (c *connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	//删除属性
	delete(c.property, key)
}
