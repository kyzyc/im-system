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
}

// NewUser 创建一个用户的接口
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}

	// 启动监听当前 user channel 消息的 goroutinue
	go user.ListenMessage()

	return user
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
