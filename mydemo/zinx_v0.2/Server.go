package main

import "zinx/znet"

func main() {
	s := znet.NewServer("[zinx_v0.2]")
	s.Serve()
}
