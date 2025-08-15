package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

var All = map[string]string{}

func sendAllMessages(conn net.Conn) {
	for _, msg := range All {
		conn.Write([]byte(msg + "\n"))
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {

		reader := bufio.NewReader(conn)
		msg, err := reader.ReadString('\n')

		if err != nil {
			fmt.Printf("Клиент отключился")
			conn.Close()
			break
		}

		All[time.Now().Format("15:04:05")] = msg

		fmt.Printf("%s Client message: %s", time.Now().Format("15:04"), msg)
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
		go sendAllMessages(conn)

		fmt.Printf("%s Client connected from %s\n", time.Now().Format("15:04"), conn.RemoteAddr().String())
		go handleConnection(conn)
	}
}
