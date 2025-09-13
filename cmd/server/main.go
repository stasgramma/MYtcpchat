package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex"`
	Messages []Message
}

type Message struct {
	ID     uint `gorm:"primaryKey"`
	Text   string
	Time   string
	Ip     string
	UserID uint
}

func sendAllMessages(conn net.Conn) {
	var messages []Message
	db.Preload("User").Find(&messages)

	for _, msg := range messages {
		var user User
		db.First(&user, msg.UserID)
		conn.Write([]byte(fmt.Sprintf(" %s : [%s] %s %s\n",
			user.Username, msg.Time, msg.Ip, msg.Text)))
	}
}

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("chat.db"), &gorm.Config{})
	if err != nil {
		panic("Не удалось подключиться к базе")
	}

	db.AutoMigrate(&User{}, &Message{})
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

			var existing User
			if err := db.Where("username = ?", proposedName).First(&existing).Error; err == nil {
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

		case "setname":
			username := words[1]

			var existing User
			if err := db.Where("username = ?", username).First(&existing).Error; err == nil {
				conn.Write([]byte("Пользователь уже существует\n"))
				return
			}

			newUser := User{Username: username}
			db.Create(&newUser)
			currentUser = username
			conn.Write([]byte("Создан новый пользователь: " + username + "\n"))

		case "connect":
			username := words[1]

			var user User
			if err := db.Where("username = ?", username).First(&user).Error; err != nil {
				conn.Write([]byte("Пользователь не найден. Сначала создайте через setname\n"))
			} else {
				currentUser = username
				conn.Write([]byte("Вы подключились как " + username + "\n"))
			}

		default:
			if currentUser != "" {
				var user User
				if err := db.Where("username = ?", currentUser).First(&user).Error; err == nil {
					msgObj := Message{
						Text:   msg,
						Time:   time.Now().Format("15:04:05"),
						Ip:     conn.RemoteAddr().String(),
						UserID: user.ID,
					}
					db.Create(&msgObj)
					conn.Write([]byte("Сообщение сохранено\n"))
				}
			}
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
