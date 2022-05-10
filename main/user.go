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
	} else if len(msg) > 4 && msg[:3] == " to|" {
		// 私聊消息格式： to|张三|消息内容
		//1 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.sendMessage("消息格式不正确，请使用 to|张三|你好啊 \n")
			return
		}
		//2 根据用户名，得到对方的User对象
		remotUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.sendMessage("私聊用户不存在\n")
			return
		}
		//3 获取消息内容， 通过对方的User将消息发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.sendMessage("无消息内容，清确认重发\n")
			return
		}
		remotUser.sendMessage(u.Name + "对你说" + content)
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
