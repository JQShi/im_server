package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	ip   string
	port int

	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		ip:        ip,
		port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (srv *Server) ListenMessage() {
	for {
		msg := <-srv.Message
		srv.mapLock.Lock()
		for _, user := range srv.OnlineMap {
			user.C <- msg
		}
		srv.mapLock.Unlock()
	}
}

func (srv *Server) BroadCast(user *User, msg string) {
	newMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	srv.Message <- newMsg
}

func (srv *Server) Handle(conn net.Conn) {
	// fmt.Println("connection created...")
	user := NewUser(conn, srv)
	user.Online()

	isLive := make(chan bool)
	go func() {

		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn read error: ", err)
				return
			}

			msg := string(buf[:n-1])
			user.DoMessage(msg)

			isLive <- true
		}
	}()

	for {
		select {
		case <-isLive:
		case <-time.After(1000 * time.Second):
			user.SendMessage("你已下线")
			close(user.C)
			fmt.Println("close [", user.Name, "] chan")
			user.conn.Close() //断开客户端的连接，服务端会读取到0个字节
			fmt.Println("close user [", user.Name, "] conn")
			return // runtime.Goexit()
		}
	}
}

func (srv *Server) Start() {

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", srv.ip, srv.port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	go srv.ListenMessage()

	fmt.Println("server is listening at ", srv.port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			return
		}
		go srv.Handle(conn)
	}

}
