package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 监听Message广播消息channel的goroutine,一旦有消息就广播给全部的在线User
func (server *Server) ListenMessage() {
	for {
		msg := <-server.Message

		// 将msg发送给全部的在线User
		server.mapLock.Lock()
		for _, cli := range server.OnlineMap {
			cli.C <- msg
		}
		server.mapLock.Unlock()
	}
}

// 广播消息的方法
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	server.Message <- sendMsg
}

func NewServer(ip string, port int) *Server {

	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

func (server *Server) Handler(conn net.Conn) {
	// 处理连接
	fmt.Println("连接成功")

	user := NewUser(conn)

	server.mapLock.Lock()
	// 用户上线，将用户添加到OnlineMap表中
	server.OnlineMap[user.Name] = user
	server.mapLock.Unlock()
	// 广播当前用户上线消息
	server.BroadCast(user, "已上线")
	// 当前handler阻塞
	select {}
}

func (server *Server) Start() {
	// 启动服务器

	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))

	if err != nil {
		fmt.Println("server listen error")
	}

	// close listen socket
	defer listener.Close()

	// 启动监听msg的goroutine
	go server.ListenMessage()

	for {

		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("server accept error")
			continue
		}

		// do handler
		go server.Handler(conn)
	}

}
