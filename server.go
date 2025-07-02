package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port, //每行后面好像都要有一个逗号，不能最后一行缺一个
	}
	return server
}

func (this *Server) Handler(conn net.Conn) {
	fmt.Println("链接建立成功")
}

func (this *Server) Start() {
	//listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("listen err:", err)
		return
	}
	defer listener.Close()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept err:", err)
			//说明go的Accept不是阻塞的？
			continue
		}

		//回调
		go this.Handler(conn)
	}

	//close
}
