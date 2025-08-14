package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Trying to connect to 127.0.0.1:3000...")

	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		fmt.Println("Не удалось подключиться к серверу:", err)
		return
	}
	fmt.Println("Connected!")

	defer conn.Close()
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Введите сообщение: ")
		msg, _ := reader.ReadString('\n')
		if msg == "exit\n" || msg == "exit \n" {
			fmt.Printf("Подключение прервано\n")
			conn.Close()
			break
		}
		fmt.Printf("Send message: %s", msg)

		conn.Write([]byte(msg))
		serverReader := bufio.NewReader(conn)
		response, _ := serverReader.ReadString('\n')
		fmt.Printf("Received message: %s", response)
	}
}
