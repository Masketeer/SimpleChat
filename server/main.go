package main

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"
	"./conf"
)

type client struct {
	conn        *net.TCPConn
	name        string
	waitingName bool
	online      bool
	onlineTime  time.Time
	offlineTime time.Time
}

type message struct {
	content string
	sender  *client
}

var clients map[string]*client
var chanMessage chan message
var chanNewClient chan *net.TCPConn
var chanClientContent chan message

var contentCache []string

func (c *client) send(content string) {
	if c.conn == nil {
		return
	}

	_, err := c.conn.Write([]byte(content))
	if err != nil {
		logError(err.Error())
	}
}

func handleClient(cli *client, ctx context.Context) {
	if cli == nil {
		logError("client is nil")
		return
	}

	conn := cli.conn
	if conn == nil {
		logError("conn is nil")
		return
	}

	defer cli.conn.Close()
	data := make([]byte, 128)
	var content string
	for {
		select {
		case <-ctx.Done():
			return
		default:
			count, err := conn.Read(data)
			if err != nil {
				logError(err.Error())

				if err.Error() == "EOF" {
					content = cli.name + "退出了聊天室"
					chanMessage <- message{
						content: content,
						sender:  cli,
					}
				}

				cli.online = false
				cli.offlineTime = time.Now()
				return
			}

			sData := string(data[:count])
			fmt.Println("receive message " + sData)

			chanClientContent <- message{
				content: sData,
				sender:  cli,
			}
		}
	}
}

func writeTask(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			break
		case msg := <-chanMessage:
			if len(clients) == 0 || msg.sender == nil {
				continue
			}

			for name, cli := range clients {
				if cli == nil || cli.conn == nil {
					continue
				}

				if msg.sender.name == name {
					continue
				}

				fmt.Printf("send message to %s\n", cli.name)
				cli.send(msg.content)
			}

		default:
			//do nothing
		}
	}
}

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:" + strconv.Itoa(int(conf.ServerConfig.Port)))
	if err != nil {
		logError(err.Error())
		return
	}

	var listener *net.TCPListener
	listener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		logError(err.Error())
		return
	}
	defer listener.Close()

	chanMessage = make(chan message, 10000)
	chanNewClient = make(chan *net.TCPConn, 10000)
	chanClientContent = make(chan message, 10000)

	clients = map[string]*client{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go writeTask(ctx)
	go accept(listener)

	for {
		select {
		case conn := <-chanNewClient:
			if conn == nil {
				continue
			}

			cli := newClient(conn)
			if cli == nil {
				logError("client is nil")
			}
			go handleClient(cli, ctx)
		case msg := <-chanClientContent:
			cli := msg.sender
			if cli == nil || cli.conn == nil {
				continue
			}

			var content string
			if !cli.waitingName {
				if _, ok := clients[msg.content]; ok {
					content = "该名称已被使用"
					cli.send(content)

					continue
				}

				cli.name = msg.content
				cli.waitingName = true
				clients[msg.content] = cli

				content = "[" + msg.content + "]" + "进入聊天室"
				chanMessage <- message{
					content: content,
					sender:  cli,
				}

				totalContent := ""
				for _, content := range contentCache {
					totalContent += content + "\n"
				}
				cli.send(totalContent)
				continue
			}

			if handleGM(msg.content, cli.conn) {
				continue
			}

			statistics(content)
			checkWord(content)
			content = "[" + cli.name + "]" + "说：" + msg.content
			contentCache = append(contentCache, content)
			if len(contentCache) > 50 {
				contentCache = contentCache[1:]
			}
			chanMessage <- message{
				content: content,
				sender:  cli,
			}
		}
	}
}

func checkWord(content string) string {
	//说明：这里可用前缀树或者KMP算法解决
	if len(conf.DirtyWords) == 0 {
		return content
	}

	for word := range conf.DirtyWords {
		content = strings.ReplaceAll(content, word, "***")
	}

	return content
}

func statistics(content string) {
	if content == "" {
		return
	}

	//这里可用线段树解决
}

func handleGM(content string, conn *net.TCPConn) bool {
	if content[0] != '/' {
		return false
	}

	commandList := strings.Split(content, " ")
	commandLength := len(commandList)
	fmt.Printf("gm command:%v\n", commandList)
	if commandLength == 0 {
		return false
	}

	switch commandList[0] {
	case "/popular":
		if commandLength <= 1 {
			return false
		}

	case "/stats":
		if commandLength <= 1 {
			return false
		}

		name := commandList[1]
		cli := clients[name]
		if cli == nil {
			return false
		}

		var timeDuration time.Duration
		if cli.online {
			timeDuration = time.Now().Sub(cli.onlineTime)
		} else {
			timeDuration = cli.offlineTime.Sub(cli.onlineTime)
		}

		_, err := conn.Write([]byte(timeDuration.String()))
		if err != nil {
			logError(err.Error())
			return false
		}
	}

	return true
}

func accept(listener *net.TCPListener) {
	if listener == nil {
		return
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			logError(err.Error())
			break
		}
		fmt.Printf("%s建立连接，进入聊天室\n", conn.RemoteAddr().String())
		_, err = conn.Write([]byte("请设置你的昵称："))
		if err != nil {
			logError(err.Error())
			continue
		}

		chanNewClient <- conn
	}
}

func logError(msg string) {
	_, _, line, _ := runtime.Caller(1)
	fmt.Printf("%v:%s\n", line, msg)
}

func newClient(conn *net.TCPConn) *client {
	if conn == nil {
		logError("conn is nil")
		return nil
	}

	//建立连接之后发送的第一条消息视为名字
	cli := &client{
		conn:       conn,
		name:       "",
		online:     true,
		onlineTime: time.Now(),
	}

	return cli
}
