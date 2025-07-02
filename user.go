package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()

	return user
}

func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}

// 用户上线业务
func (this *User) Online() {
	this.server.mapLock.Lock()

	this.server.OnlineMap[this.Name] = this

	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "已上线")
}

// 下线
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	this.server.BroadCast(this, "已下线")
}

// 给指定用户发消息
func (this *User) sendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 发消息
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前用户有哪些
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			this.sendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else { //神经病，go的if的闭括号，即 if {} 的右括号，要和else的紧邻，一起写，不能分开
		this.server.BroadCast(this, msg)
	}

}
