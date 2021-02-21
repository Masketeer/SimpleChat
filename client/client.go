package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
)

func writeTask(conn net.Conn, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			break
		default:
			input := bufio.NewReader(os.Stdin)
			message, _, _ := input.ReadLine()
			_, err := conn.Write(message)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}

func main() {
	port := 8888
	if len(os.Args) > 1 {
		number, err := strconv.Atoi(os.Args[1])
		if err == nil {
			port = number
		}
	}
	conn, err := net.Dial("tcp", "127.0.0.1:" + strconv.Itoa(port))
	if err != nil {
		fmt.Println(err.Error())
	}

	if conn == nil {
		fmt.Println("conn is nil")
		return
	}

	defer conn.Close()
	ctx, cancel := context.WithCancel(context.Background())
	go writeTask(conn, ctx)

	buf := make([]byte, 1024)
	for {
		var count int
		count, err = conn.Read(buf)
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		fmt.Println(string(buf[:count]))
	}

	if cancel != nil {
		cancel()
	}
}