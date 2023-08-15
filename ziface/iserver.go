package ziface

type IServer interface {
	Start()
	Stop()
	Serve()
	// 路由功能：给当前服务注册一个路由方法，供客户端的链接处理使用
	AddRouter(msgID uint32, router IRouter)
	// 获取当前 Server 的链接管理器
	GetConnMgr() IConnManager
	// 注册 OnConnStart 钩子函数的方法
	SetOnConnStart(func(connection IConnection))
	// 注册 OnConnStop 钩子函数的方法
	SetOnConnStop(func(connection IConnection))
	// 调用 OnConnStart 钩子函数的方法
	CallOnConnStart(connection IConnection)
	// 调用 OnConnStop 钩子函数的方法
	CallOnConnStop(connection IConnection)
}
