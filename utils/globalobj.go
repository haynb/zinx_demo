package utils

import (
	"encoding/json"
	"os"
	"zinx/ziface"
)

/*
存储全局变量
*/
type GlobalObj struct {
	/*
		Server
	*/
	TcpServer ziface.IServer
	Host      string
	TcpPort   int
	Name      string

	/*
		Zinx
	*/
	Version          string
	MaxConn          int
	MaxPackageSize   uint32
	WorkerPoolSize   uint32
	MaxWorkerTaskLen uint32
}

/*
定义一个全局的对外GlobalObj
*/
var GlobalObject *GlobalObj

func (g *GlobalObj) Reload() {
	data, err := os.ReadFile("conf/zinx.json")
	if err != nil {
		panic(err)
	}
	// 将json数据解析到struct中
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}

// 初始化方法,初始化当前的GlobalObject
func init() {
	// 如果配置文件没有加载,默认值
	GlobalObject = &GlobalObj{
		Name:             "ZinxServerApp",
		Version:          "V0.8",
		TcpPort:          8999,
		Host:             "0.0.0.0",
		MaxConn:          1000,
		MaxPackageSize:   4096,
		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024, //每个worker对应的消息队列的任务的数量最大值
	}
	// 从配置文件中加载一些用户配置的参数
	GlobalObject.Reload()
}
