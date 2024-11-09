package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的 channel
	Message chan string
}

// NewServer 创建 server 的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// ListenMessager 监听 Message 广播消息 channel 的 goroutinue，一旦有消息就发送给全部的在线 User
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		// 将 msg 发送给全部的在线 User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// BroadCast 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	// 当前连接的业务
	//fmt.Println("连接创建成功")

	user := NewUser(conn, this)

	user.Online()

	// 监听用户是否活跃的 channel
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			// 提取用户的消息（去除'\n'）
			msg := string(buf[:n-1])

			// 用户针对 msg 进行消息处理
			user.DoMessage(msg)

			// 用户任意消息，代表当前用户活跃
			isLive <- true
		}
	}()

	// 当前 handler 阻塞
	for {
		select {
		case <-isLive:
		case <-time.After(time.Second * 10):
			// 已经超时
			// 将当前的 user 强制关闭
			user.SendMsg("长时间不活跃，被踢除")

			// 销毁用的资源
			close(user.C)

			// 关闭连接
			conn.Close()

			// 退出当前 handler
			return
		}
	}
}

// Start 启动服务器的接口
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// close socket
	defer listener.Close()

	// 启动监听 Message 的 goroutinue
	go this.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// do handler
		go this.Handler(conn)
	}
}
