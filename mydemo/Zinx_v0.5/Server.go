package main

import (
	"fmt"
	"zinx/ziface"
	"zinx/znet"
)

// ping test 自定义路由
type PingRouter struct {
	znet.BaseRouter
}

// Test Handle
func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("call back ping...")
	// 先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client: msgID=", request.GetMsgID(), ", data=", string(request.GetData()))
	err := request.GetConnection().SendMsg(1, []byte("ping...ping...ping"))
	if err != nil {
		fmt.Println(err)
	}
}
func main() {
	s := znet.NewServer("[zinx_v0.3]")
	s.AddRouter(&PingRouter{})
	s.Serve()
}
