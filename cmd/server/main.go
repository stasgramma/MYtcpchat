package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"tcp/db"
	"time"
	"unicode/utf8"
)

func sendAllMessages(conn net.Conn) {
	var messages []db.Message
	db.DB.Preload("User").Find(&messages)

	for _, msg := range messages {
		var user db.User
		db.DB.First(&user, msg.UserID)
		conn.Write([]byte(fmt.Sprintf(" %s : [%s] %s %s\n",
			user.Username, msg.Time, msg.Ip, msg.Text)))
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	var currentUser string

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
			proposedName := strings.TrimSpace(strings.TrimPrefix(msg, "NAME:"))

			var existing db.User
			if err := db.DB.Where("username = ?", proposedName).First(&existing).Error; err == nil {
				conn.Write([]byte("Это имя уже занято, выберите другое\n"))
				continue
			}

			username = proposedName
			conn.Write([]byte("Имя установлено: " + username + "\n"))
		}

		totalb := 0
		for _, bytee := range msg {
			totalb += utf8.RuneLen(bytee)
		}
		words := strings.Fields(msg)

		fmt.Printf("%s Client message: %s\n", time.Now().Format("15:04"), msg)
		fmt.Printf("Total bytes : %v\n ", totalb)
		fmt.Printf("Worlds quantity on message: %v\n ", len(words))
		switch words[0] {
		case "/echo":
			result := strings.Join(words[1:], " ")
			conn.Write([]byte(result + "\n"))

		case "/mul":
			a, _ := strconv.Atoi(words[1])
			b, _ := strconv.Atoi(words[2])
			sum := a * b
			conn.Write([]byte(fmt.Sprintf("%d\n", sum)))

		case "/sum":
			a, _ := strconv.Atoi(words[1])
			b, _ := strconv.Atoi(words[2])
			sum := a + b
			conn.Write([]byte(fmt.Sprintf("%d\n", sum)))

		case "/setname":
			username := words[1]

			var existing db.User
			if err := db.DB.Where("username = ?", username).First(&existing).Error; err == nil {
				conn.Write([]byte("Пользователь уже существует\n"))
				return
			}

			newUser := db.User{Username: username}
			db.DB.Create(&newUser)
			currentUser = username
			conn.Write([]byte("Создан новый пользователь: " + username + "\n"))

		case "/connect":
			username := words[1]

			var user db.User
			if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
				conn.Write([]byte("Пользователь не найден. Сначала создайте через setname\n"))
			} else {
				currentUser = username
				conn.Write([]byte("Вы подключились как " + username + "\n"))
			}

		default:
			if currentUser != "" {
				var user db.User
				if err := db.DB.Where("username = ?", currentUser).First(&user).Error; err == nil {
					msgObj := db.Message{
						Text:   msg,
						Time:   time.Now().Format("15:04:05"),
						Ip:     conn.RemoteAddr().String(),
						UserID: user.ID,
					}
					db.DB.Create(&msgObj)
					conn.Write([]byte("Сообщение сохранено\n"))
				}
			}
			conn.Write([]byte(msg + " from server\n"))
		}

		fmt.Printf("%s Send message to client: %s from server\n", time.Now().Format("15:04"), msg)

	}
}

func main() {
	db.Init()
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
		return
	}
	fmt.Printf("Сервер запустился\n")

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
