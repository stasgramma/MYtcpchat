package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

var All []Mape

type Mape struct {
	Time string
	Sms  string
}

func sendAllMessages(conn net.Conn) {
	for _, msg := range All {
		conn.Write([]byte(fmt.Sprintf("[%s] %s\n", msg.Time, msg.Sms)))
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		sendAllMessages(conn)

		reader := bufio.NewReader(conn)
		msg, err := reader.ReadString('\n')
		msg = strings.TrimSpace(msg)
		if err != nil {
			fmt.Printf("Клиент отключился")
			conn.Close()
			break
		}

		All = append(All, Mape{
			Time: time.Now().Format("15:04:05"),
			Sms:  msg,
		})

		fmt.Printf("%s Client message: %s\n", time.Now().Format("15:04"), msg)
		fmt.Printf("%s Send message to client: %s from server\n", time.Now().Format("15:04"), msg)

		conn.Write([]byte(msg + " from server\n"))
	}

}
func main() {
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Ошибка подключения:", err)
			continue
		}

		fmt.Printf("%s Client connected from %s\n", time.Now().Format("15:04"), conn.RemoteAddr().String())
		go handleConnection(conn)
	}
}
