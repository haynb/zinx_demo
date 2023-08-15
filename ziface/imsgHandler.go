package ziface

/*
消息管理抽象层
*/
type IMsgHandler interface {
	//调度、执行对应的Router消息处理方法
	DoMsgHandler(request IRequest)
	//添加路由
	AddRouter(msgID uint32, router IRouter)
	//启动Worker工作池
	StartWorkerPool()
	//将消息交给TaskQueue，由worker进行处理
	SendMsgToTaskQueue(request IRequest)
}
