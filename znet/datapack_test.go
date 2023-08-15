package znet

import (
	"io"
	"net"
	"testing"
)

func TestDatePack(t *testing.T) {
	/*
		模拟服务器
	*/
	listenner, err := net.Listen("tcp", "127.0.0.1:7777")
	if err != nil {
		t.Error("server listen err:", err)
		return
	}
	// 创建一个go承载，负责从客户端处理业务
	go func() {
		conn, err := listenner.Accept()
		if err != nil {
			t.Error("server accept error:", err)
			return
		}
		// 处理客户端的请求
		go func(conn net.Conn) {
			// 拆包的过程
			// 定义一个拆包的对象dp
			dp := NewDatePack()
			for {
				// 1. 第一次从conn读，把包的head读出来
				headData := make([]byte, dp.GetHeadLen())
				_, err := io.ReadFull(conn, headData)
				if err != nil {
					t.Error("read head error")
					break
				}
				// 将headData字节流拆包到msg中
				msgHead, err := dp.Unpack(headData)
				if err != nil {
					t.Error("server unpack err:", err)
					return
				}
				if msgHead.GetMsgLen() > 0 {
					// msg是有data数据的，需要再次读取data数据
					// 2. 第二次从conn读，根据head中的dataLen再读取data内容
					msg := msgHead.(*Message)
					msg.Data = make([]byte, msg.GetMsgLen())
					// 根据dataLen的长度再次从io流中读取
					_, err := io.ReadFull(conn, msg.Data)
					if err != nil {
						t.Error("server unpack data err:", err)
						return
					}

					// 完整的一个消息已经读取完毕
					t.Logf("-----> Recv MsgID: %d, dataLen: %d, data: %s", msg.Id, msg.DataLen, string(msg.Data))
				}
			}
		}(conn)
	}()
	/*
		模拟客户端
	*/
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		t.Error("client dial error:", err)
		return
	}
	// 创建一个封包对象dp
	dp := NewDatePack()
	// 模拟粘包过程，封装两个msg一同发送
	// 封装第一个msg1包
	msg1 := &Message{
		Id:      1,
		DataLen: 5,
		Data:    []byte{'h', 'e', 'l', 'l', 'o'},
	}
	sendData1, err := dp.Pack(msg1)
	if err != nil {
		t.Error("client pack msg1 error:", err)
		return
	}
	// 封装第二个msg2包
	msg2 := &Message{
		Id:      2,
		DataLen: 7,
		Data:    []byte{'w', 'o', 'r', 'l', 'd', '!', '!'},
	}
	sendData2, err := dp.Pack(msg2)
	if err != nil {
		t.Error("client pack msg2 error:", err)
		return
	}
	// 将sendData1，和sendData2拼接一起，组成粘包
	sendData1 = append(sendData1, sendData2...)
	// 一次性发送给服务端
	conn.Write(sendData1)
	// 客户端阻塞
	select {}
}
