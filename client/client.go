package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	menuChoose int // 当前 client 的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		menuChoose: -1,
	}

	// 连接 server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("client connect server err:", err)
		return nil
	}

	client.conn = conn

	// 返回对象
	return client
}

// DealResponse 处理 server 回应的消息，直接显示到标准输出
func (client *Client) DealResponse() {
	// 一旦 client.conn 有数据，就直接拷贝到标准输出上，永久阻塞监听
	_, err := io.Copy(os.Stdout, client.conn)
	if err != nil {
		fmt.Println("io.Copy err:", err)
		return
	}
}

func (client *Client) UpdateName() bool {
	fmt.Printf(">>>>请输入用户名：")
	_, err := fmt.Scanln(&client.Name)
	if err != nil {
		fmt.Println("fmt.Scanln err:", err)
		return false
	}

	sendMsg := "rename|" + client.Name + "\n"
	_, err = client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

func (client *Client) menu() bool {
	var menuChoose int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	_, err := fmt.Scanln(&menuChoose)
	if err != nil {
		fmt.Println("fmt.Scanln err:", err)
		return false
	}

	if menuChoose >= 0 && menuChoose <= 3 {
		client.menuChoose = menuChoose
		return true
	} else {
		fmt.Println(">>>>请输入范围合法的数字<<<<")
		return false
	}
}

func (client *Client) Run() {
	for client.menuChoose != 0 {
		for client.menu() != true {
		}
		// 根据不同的模式处理不同的业务
		switch client.menuChoose {
		case 1:
			// 公聊模式
			fmt.Println("公聊模式选择...")
		case 2:
			// 私聊模式
			fmt.Println("私聊模式选择...")
		case 3:
			// 更新用户名
			fmt.Println("更新用户名选择...")
			client.UpdateName()
		}
	}
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址（默认是127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口号（默认是8888）")
}

func main() {
	flag.Parse()

	client := NewClient(serverIp, serverPort)

	if client == nil {
		fmt.Println("连接服务器失败...")
		return
	}
	fmt.Println("连接服务器成功...")

	// 单独开启一个 goroutinue 去处理服务器发来的消息
	go client.DealResponse()

	// 启动客户端的业务
	client.Run()
}
