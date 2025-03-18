package main

import (
	"log"
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// 用户上线业务
func (user *User) Online() {
	user.server.mapLock.Lock()
	// 用户上线，将用户添加到OnlineMap表中
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()
	// 广播当前用户上线消息
	user.server.BroadCast(user, "已上线")
}

// 用户下线业务
func (user *User) Offline() {
	user.server.mapLock.Lock()
	// 用户上线，将用户添加到OnlineMap表中
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()
	// 广播当前用户离线消息
	user.server.BroadCast(user, "下线了")
}

// 解析用户消息
func (user *User) DoMessage(msg string) {
	log.Println("msg=", len(msg))
	if msg == "who" {
		user.server.mapLock.Lock()
		for _, cli := range user.server.OnlineMap {
			onlineMsg := "[" + cli.Addr + "]" + cli.Name + ":" + "在线...\n"
			user.SendMsg(onlineMsg)
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式 rename|张三
		newName := strings.Split(msg, "|")[1]
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.SendMsg("当前用户名被使用了\n")
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()
			user.Name = newName
			user.SendMsg("您已经成功更新用户名：" + user.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式 to|张三|消息内容
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			user.SendMsg("消息格式不正确，请使用 to|张三|消息内容\n")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			user.SendMsg("消息格式不正确，请使用 to|张三|消息内容\n")
			return
		}
		remoteUser, ok := user.server.OnlineMap[remoteName]
		if !ok {
			user.SendMsg("该用户不存在\n")
		}
		remoteUser.SendMsg(user.Name + "发送消息：" + content + "\n")

	} else {
		user.server.BroadCast(user, msg)
	}

}

// 给当前User对应的客户端发送消息
func (user *User) SendMsg(msg string) {
	user.conn.Write([]byte(msg))
}

// 创建用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// 启动监听当前User channel的方法，一旦有消息，就直接发送给对端客户端
	go user.ListenMessage()

	return user
}

// 监听当前User channel的方法，一旦有消息，就直接发送给对端客户端

func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
