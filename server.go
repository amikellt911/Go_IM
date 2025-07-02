package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port, //每行后面好像都要有一个逗号，不能最后一行缺一个
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听Message chan，广播
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		//msg发送所有在线User
		this.mapLock.Lock()
		for _, user := range this.OnlineMap {
			user.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//fmt.Println("链接建立成功")
	//用户上线，将用户加入表中

	user := NewUser(conn, this)

	user.Online()

	//读取用户消息
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			//eof表示读到文件的末尾，不需要管
			if err != nil && err != io.EOF {
				fmt.Println("conn.Read err:", err)
				return
			}
			//提取用户消息（去除'\n'）
			msg := string(buf[:n-1])
			//消息广播
			user.DoMessage(msg)
		}
	}()

	//阻塞
	select {}
}

func (this *Server) Start() {
	//listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("listen err:", err)
		return
	}
	//close
	defer listener.Close()
	go this.ListenMessager()
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

}
