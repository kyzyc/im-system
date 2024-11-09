package main

import (
	"fmt"
	"net"
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

// DoMessage 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	this.server.BroadCast(this, msg)
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
