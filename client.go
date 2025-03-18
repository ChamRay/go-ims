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
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
	}

	// 连接服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.ServerIp, client.ServerPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	return client
}

func (client *Client) menu() bool {
	for {
		fmt.Println("1. 显示在线用户列表")
		fmt.Println("2. 私聊")
		fmt.Println("3. 群聊")
		fmt.Println("4. 更改用户名")
		fmt.Println("5. 退出")

		var key int
		fmt.Scanln(&key)

		if key >= 0 && key <= 3 {
			client.flag = key
			return true
		}
		return false
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println("请输入用户名：")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error:", err)
		return false

	}
	return true
}

func (client *Client) QueryUserList() {
	msg := "who\n"
	_, err := client.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("conn.Write error:", err)
		return
	}
}

func (client *Client) PrivateChat() {
	// 查询在线用户列表
	client.QueryUserList()
	fmt.Println("请输入聊天对象：")
	var remoteName string
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		if remoteName == client.Name {
			fmt.Println("不能跟自己聊天")
			continue
		}
		fmt.Println("请输入聊天内容：")
		var chatMsg string
		fmt.Scanln(&chatMsg)
		for chatMsg == "exit" {
			sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
			if len(sendMsg) > 0 {
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write error:", err)
					return
				}
			}
		}

	}

}

func (client *Client) GroupChat() {

	var chatMsg string
	fmt.Println("请输入聊天内容，exit退出")
	fmt.Scanln(&chatMsg)

	for chatMsg == "exit" {
		// 发送给服务器
		client.flag = 5
		if len(chatMsg) > 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write error:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println("请输入聊天内容，exit退出")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) Exit() {
	msg := "0"
	_, err := client.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("conn.Write error:", err)
		return
	}
	client.conn.Close()
}

func (client *Client) DealResponse() {
	// 一旦client.conn 有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) Run() {
	for client.flag != 0 {
		if client.menu() != true {
			var key = client.flag
			// 根据不同的模式处理不同的业务
			switch key {
			case 1: // 显示在线用户列表
				// 发送who指令
				client.QueryUserList()
				break
			case 2: // 私聊
				client.PrivateChat()
				break
			case 3: // 群聊
				client.GroupChat()
				break
			case 4: // 更改用户名
				client.UpdateName()
				break
			case 5: // 退出
				client.Exit()
				return
			default:
				fmt.Println("请输入合法的选项")
			}
		}

	}
}

var serverIp string
var serverPort int

func init() {
	// 通过外部设置连接地址和端口
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址（默认127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口号（默认8888）")
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("连接服务器失败")
		return
	}
	fmt.Println("连接服务器成功")

	go client.DealResponse()
	// 启动客户端的业务
	client.Run()
}
