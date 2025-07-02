package main

import (
	"net"
	"strings"
)

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
		//重命名
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式：rename|张三
		newName := strings.Split(msg, "|")[1]
		//查询用户名是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.sendMsg("当前用户名被使用\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()
			this.Name = newName
			this.sendMsg("您已更新用户名:" + newName + "\n")
		}
	} else if len(msg) >= 4 && msg[:3] == "to|" {
		//消息格式：to|张三|消息内容

		//1. 获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.sendMsg("消息格式不正确，请使用 \"to|张三|你好啊\" 的格式完成\n")
			return
		}
		//2.获取User对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.sendMsg("该用户不存在\n")
			return
		}
		//3.消息内容发送给对方
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.sendMsg("无消息内容，请重发\n")
			return
		}
		remoteUser.sendMsg(this.Name + "对你说：" + content + "\n")
	} else { //神经病，go的if的闭括号，即 if {} 的右括号，要和else的紧邻，一起写，不能分开
		this.server.BroadCast(this, msg)
	}

}
