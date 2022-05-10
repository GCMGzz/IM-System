package main

import (
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

//	创建一个用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

// 用户上线业务
func (u *User) Online() {
	//用户上线后 加入到onlineMap中
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	// 广播当前用户上线信息
	u.server.BroadCast(u, "上线啦")
}

// 下线业务
func (u *User) Offline() {
	// 用户下线后 从onlineMap中删除
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	// 广播当前用户上线信息
	u.server.BroadCast(u, "下线咯")
}

//给当前User的客户端发消息
func (this *User) sendMessage(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息的业务
func (u *User) DoMessgae(msg string) {
	if msg == "who" {
		// 查询当前在线的用户
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":在线...\r"
			u.sendMessage(onlineMsg)
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 改名格式： rename|张三
		newName := strings.Split(msg, "|")[1]
		// 判断名字是否已经存在
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.sendMessage("用户名已经存在\n")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap["newName"] = u
			u.server.mapLock.Unlock()

			u.Name = newName
			u.sendMessage("成功修改用户名为：" + u.Name + "\n")
		}
	} else {
		u.server.BroadCast(u, msg)
	}
}

// 监听当前user channel的方法，一旦有消息直接发送给客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}

}
