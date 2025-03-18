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

	user := NewUser(conn, server)
	user.Online()
	// 处理连接
	fmt.Println(user.Name + "连接成功")

	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接收客户端发送的消息  客户端消息需要自定义协议来进行分类解析
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn read err:", err)
				return
			}

			// msg := TrimNewLineWithoutSpace(string(buf[:n]))
			msg := string(buf[:n-1])
			// 将消息进行广播
			user.DoMessage(msg)

			// 用户的任意消息，代表当前用户是一个活跃的
			isLive <- true
		}

	}()

	for {
		// 当前handler阻塞 todo 超时机制
		select {
		case <-isLive:
			// 当前用户是活跃的，有消息上行，重置定时器，什么都不做，更新下面的定时器

		case <-time.After(time.Second * 300):
			println(user.Name + "超时被踢!")
			user.SendMsg("超时未发送消息，你被踢了")
			// 销毁用户资源
			close(user.C)
			// 关闭连接
			conn.Close()
			return
		}
	}
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
