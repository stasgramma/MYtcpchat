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

var All []Mape

type Mape struct {
	User string
	Id   uint `gorm:"primaryKey"`
	Time string
	Ip   string
	Sms  string
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

			nameExists := false
			for _, m := range All {
				if m.User == proposedName {
					nameExists = true
					break
				}
			}

			if nameExists {
				conn.Write([]byte("Это имя уже занято, выберите другое\n"))
				continue
			}

			username = proposedName
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

		case "setname":

			fullperson := words[1]
			sqlStmt := fmt.Sprintf(`
    CREATE TABLE IF NOT EXISTS %s (
        message TEXT
    );`, fullperson)

			if err := db.Exec(sqlStmt).Error; err != nil {
				fmt.Println("Ошибка создания таблицы:", err)
			} else {
				fmt.Println("Таблица создана для сообщений:", fullperson)
			}
			currentUser = fullperson
		case "connect":
			fullperson := words[1]

			var tableName string
			checkStmt := fmt.Sprintf(`SELECT name FROM sqlite_master WHERE type='table' AND name='%s';`, fullperson)
			db.Raw(checkStmt).Scan(&tableName)

			if tableName == "" {
				conn.Write([]byte("Пользователь не найден. Сначала создайте через setname\n"))
			} else {
				currentUser = fullperson
				conn.Write([]byte("Вы подключились как " + currentUser + "\n"))
			}

		default:

			if currentUser == "" {
			} else {
				insertStmt := fmt.Sprintf("INSERT INTO %s (message) VALUES (?)", currentUser)
				db.Exec(insertStmt, msg)
				conn.Write([]byte("Сообщение сохранено\n"))
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
