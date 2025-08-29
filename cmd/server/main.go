package main

import (
	"bufio"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var All []Mape

type Mape struct {
	Id   uint `gorm:"primaryKey"`
	Time string
	Ip   string
	Sms  string
	User string
}

func sendAllMessages(conn net.Conn) {
	for _, msg := range All {
		conn.Write([]byte(fmt.Sprintf(" %s  : [%s]  %s  %s\n", msg.User, msg.Time, msg.Ip, msg.Sms)))
	}
}

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("chat.db"), &gorm.Config{})
	if err != nil {
		panic("Не удалось подключиться к базе")
	}

	db.AutoMigrate(&Mape{})

	var messages []Mape
	db.Find(&messages)
	All = messages
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	var username string
	sendAllMessages(conn)

	for {

		msg, err := reader.ReadString('\n')
		msg = strings.TrimSpace(msg)
		if err != nil {
			fmt.Printf("Клиент отключился")
			conn.Close()
			break
		}
		if strings.HasPrefix(msg, "NAME:") {
			username = strings.TrimSpace(strings.TrimPrefix(msg, "NAME:"))
			fmt.Printf("Клиент %s установил имя: %s\n", conn.RemoteAddr().String(), username)
			continue
		}

		msgObj := Mape{
			User: username,

			Time: time.Now().Format("15:04:05"),
			Ip:   conn.RemoteAddr().String(),
			Sms:  msg,
		}
		All = append(All, msgObj)

		db.Create(&msgObj)

		totalb := 0
		for _, bytee := range msg {
			totalb += utf8.RuneLen(bytee)
		}
		words := strings.Fields(msg)

		fmt.Printf("%s Client message: %s\n", time.Now().Format("15:04"), msg)
		fmt.Printf("Total bytes : %v\n ", totalb)
		fmt.Printf("Worlds quantity on message: %v\n ", len(words))

		switch words[0] {
		case "echo":
			result := strings.Join(words[1:], " ")
			conn.Write([]byte(result + "\n"))

		case "mul":
			a, _ := strconv.Atoi(words[1])
			b, _ := strconv.Atoi(words[2])
			sum := a * b
			conn.Write([]byte(fmt.Sprintf("%d\n", sum)))

		case "sum":
			a, _ := strconv.Atoi(words[1])
			b, _ := strconv.Atoi(words[2])
			sum := a + b
			conn.Write([]byte(fmt.Sprintf("%d\n", sum)))

		default:
			conn.Write([]byte(msg + " from server\n"))

		}

		fmt.Printf("%s Send message to client: %s from server\n", time.Now().Format("15:04"), msg)

	}
}

func main() {
	initDB()
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
