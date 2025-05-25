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

func (client *Client) menu() bool {
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	var flag int
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("=====请输入选择正确的模式======")
		return false
	}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for !client.menu() {
		}
		switch client.flag {
		case 1:
			// fmt.Println("公聊模式")
			client.PublicChat()
			break
		case 2:
			// fmt.Println("私聊模式")
			client.PrivateChat()
			break
		case 3:
			// fmt.Println("更新用户名")
			client.UpdateName()
			break
		}
	}
}

func (client *Client) SendMsg(msg string) bool {
	_, err := client.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("conn write err: ", err)
		return false
	}
	return true
}

func (client *Client) SelectUsers() {
	_, err := client.conn.Write([]byte("who\n"))
	if err != nil {
		fmt.Println("conn write err: ", err)
	}
}

func (client *Client) PrivateChat() {
	client.SelectUsers()

	fmt.Println(">>>>请输入聊天用户")
	var remoteName string
	fmt.Scanln(&remoteName)
	for remoteName != "exit" {
		fmt.Println(">>>>请输入聊天内容")
		var content string
		fmt.Scanln(&content)
		for content != "exit" {
			if len(content) != 0 {
				newMsg := "to|" + remoteName + "|" + content + "\n\n"
				rlt := client.SendMsg(newMsg)
				if !rlt {
					break
				}
			}

			content = ""
			fmt.Println(">>>>请输入聊天内容")
			fmt.Scanln(&content)
		}

		remoteName = ""
		client.SelectUsers()
		fmt.Println(">>>>请输入聊天用户")
		fmt.Scanln(&remoteName)
	}

	// fmt.Println(">>>>请输入聊天用户")
	// client.conn.Write([]byte("who\n"))
	// var receiver string
	// fmt.Scanln(&receiver)
	// for len(receiver) == 0 {
	// 	client.conn.Write([]byte("who\n"))
	// 	fmt.Scanln(&receiver)
	// }
	// if receiver == "exit" {
	// 	return
	// }
	// fmt.Println(">>>>请输入聊天内容")
	// var content string
	// fmt.Scanln(&content)
	// for content != "exit" {
	// 	if len(content) != 0 {
	// 		newMsg := "to|" + receiver + "|" + content + "\n"
	// 		client.SendMsg(newMsg)
	// 	}
	// 	content = ""
	// 	fmt.Scanln(&content)
	// }

}

func (client *Client) PublicChat() {
	var content string
	fmt.Println(">>>>请输入聊天内容")
	fmt.Scanln(&content)
	for content != "exit" {
		if len(content) != 0 {
			newMsg := content + "\n"
			_, err := client.conn.Write([]byte(newMsg))
			if err != nil {
				fmt.Println("conn write err: ", err)
				break
			}

			content = ""
			fmt.Scanln(&content)
		}
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>>请输入你的用户名")
	fmt.Scanln(&client.Name)

	newMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(newMsg))
	if err != nil {
		fmt.Println("conn write err: ", err)
		return false
	}
	return true
}

func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.conn)
	// for {
	// 	buf := make([]byte, 4096)
	// 	client.conn.Read(buf)
	// 	fmt.Println(buf)
	// }
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net dial err:", err)
		return nil
	}
	client.conn = conn
	return client
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip，默认值127.0.0.1")
	flag.IntVar(&serverPort, "p", 8888, "设置服务器端口，默认值8888")
}

func main() {
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("============== 连接创建失败")
		return
	}

	go client.DealResponse()

	fmt.Println(">>>>>>>>>>> 已连接")

	client.Run()
}
