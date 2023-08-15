package main

import (
	"fmt"
	"io"
	"net"
	"time"
	"zinx/znet"
)

/*
 模拟客户端
*/

func main() {
	fmt.Println("client start...")
	time.Sleep(1 * time.Second)
	//客户端发起请求
	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("client start err, exit!")
		return
	}
	for {
		//发封包message消息
		dp := znet.NewDatePack()
		binaryMsg, err := dp.Pack(znet.NewMsgPackage(0, []byte("num0_ZinxV0.6 client Test Message")))
		if err != nil {
			fmt.Println("Pack error:", err)
			return
		}
		if _, err := conn.Write(binaryMsg); err != nil {
			fmt.Println("write error:", err)
			return
		}
		//服务器应该回复message数据，msgID:1 ping...ping...ping
		//先读取流中的head部分，得到ID和dataLen
		binaryHead := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(conn, binaryHead); err != nil {
			fmt.Println("read head error:", err)
			break
		}
		//将二进制的head拆包到msg结构体中
		msgHead, err := dp.Unpack(binaryHead)
		if err != nil {
			fmt.Println("client unpack msgHead error:", err)
			break
		}
		if msgHead.GetMsgLen() > 0 {
			//msg是有data数据的，需要再次读取data数据
			msg := msgHead.(*znet.Message)
			msg.Data = make([]byte, msg.GetMsgLen())
			//根据dataLen从io中读取字节流
			if _, err := io.ReadFull(conn, msg.Data); err != nil {
				fmt.Println("read msg data error:", err)
				return
			}
			fmt.Println("---->Recv Server Msg:ID=", msg.Id, ", len=", msg.DataLen, ", data=", string(msg.Data))
		}
		//cpu阻塞
		time.Sleep(1 * time.Second)
	}
}
