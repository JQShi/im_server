package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	conn net.Conn

	C      chan string
	server *Server
}

func (user *User) Online() {
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()
	fmt.Println("user ", user.Name, " online")
	user.DoMessage("已上线")
}

func (user *User) Offline() {
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()
	fmt.Println("user ", user.Name, " offline")
	user.DoMessage("已下线")
}

func (user *User) DoMessage(msg string) {
	if msg == "who" {
		user.server.mapLock.Lock()
		for _, u := range user.server.OnlineMap {
			newMsg := "[" + u.Addr + "]" + u.Name + " is online"
			user.SendMessage(newMsg)
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.SendMessage("用户名[" + newName + "]已存在")
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()
			user.Name = newName
			user.SendMessage("您已更新用户名:" + newName)
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		arr := strings.Split(msg, "|")
		if len(arr) != 3 {
			user.SendMessage("消息格式不正确，请使用\"to|tom|你好\"格式")
			return
		}
		remoteName := arr[1]
		if remoteName == "" {
			user.SendMessage("消息格式不正确，请使用\"to|tom|你好\"格式")
			return
		}
		content := arr[2]
		if content == "" {
			user.SendMessage("消息内容为空，请重新发送")
			return
		}
		remoteUser, ok := user.server.OnlineMap[remoteName]
		if !ok {
			user.SendMessage("用户名不存在")
			return
		}
		newMsg := user.Name + " -> you: " + content
		fmt.Println("send to ", remoteName, " ", newMsg)
		remoteUser.SendMessage(newMsg)
	} else {
		user.server.BroadCast(user, msg)
	}
}

func (user *User) SendMessage(msg string) {
	user.conn.Write([]byte(msg + "\n"))
}

func (user *User) ListenMessage() {
	// for {
	// if msg, ok := <-user.C; ok {
	// 	user.SendMessage(msg)
	// } else {
	// 	fmt.Println("user.chan is closed")
	// }
	// msg := <-user.C
	// user.SendMessage(msg)
	// fmt.Println(user.Name, " is listening [", msg, "]")
	// }
	for msg := range user.C {
		user.SendMessage(msg)
	}
}

func NewUser(conn net.Conn, srv *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		conn:   conn,
		C:      make(chan string),
		server: srv,
	}
	go user.ListenMessage()
	return user
}
