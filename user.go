package main

import (
	"fmt"
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

// NewUser 创建一个用户的接口
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// 启动监听当前 user channel 消息的 goroutinue
	go user.ListenMessage()

	return user
}

// Online 用户的上线业务
func (this *User) Online() {
	// 用户上线，将用户加入到 OnlineMap 中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播当前用户上线消息
	this.server.BroadCast(this, "已上线")
}

// Offline 用户的下线业务
func (this *User) Offline() {
	// 用户下线，将用户从 OnlineMap 中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播当前用户下线消息
	this.server.BroadCast(this, "下线")
}

// SendMsg 给当前 user 对应的客户端发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// DoMessage 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户有哪些
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式：rename|张三
		newName := strings.Split(msg, "|")[1]
		// 判断 name 是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("当前用户名被使用\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("成功更新用户名为：" + this.Name + "\n")
		}
	} else {
		this.server.BroadCast(this, msg)
	}
}

// ListenMessage 监听当前 User channel 的方法，一旦有消息，直接发给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		_, err := this.conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Printf("write to %s error!", this.Name)
			return
		}
	}
}
