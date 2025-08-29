package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
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
	go func() {
		serverReader := bufio.NewReader(conn)
		for {
			msg, err := serverReader.ReadString('\n')
			msg = strings.TrimSpace(msg)
			if err != nil {
				log.Printf("err = %v\n", err)
				return
			}
			fmt.Printf("Reseave message %s\n", msg)
		}
	}()
	time.Sleep(200 * time.Millisecond)

	fmt.Print("Введите ваше имя: ")
	nameReader := bufio.NewReader(os.Stdin)
	name, _ := nameReader.ReadString('\n')
	name = strings.TrimSpace(name)
	reader := bufio.NewReader(os.Stdin)

	conn.Write([]byte("NAME:" + name + "\n"))
	for {
		time.Sleep(200 * time.Millisecond)

		fmt.Print("Введите сообщение: ")
		rawmsg, _ := reader.ReadString('\n')
		msg := strings.TrimSpace(rawmsg)

		if msg == "exit" {
			fmt.Printf("Подключение прервано\n")
			conn.Close()
			break
		}
		fmt.Printf("Send message: %s\n", msg)

		conn.Write([]byte(rawmsg))
	}
}
