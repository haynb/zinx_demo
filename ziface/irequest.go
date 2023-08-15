package ziface

/*
IRequest 接口：
实际上是把客户端请求的连接信息和请求的数据包装到了 Request 里
*/
type IRequest interface {
	// GetConnection 获取当前连接
	GetConnection() IConnection
	// GetData 获取请求的消息数据
	GetData() []byte
	// GetMsgID 获取请求的消息 ID
	GetMsgID() uint32
}
