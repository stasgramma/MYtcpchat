package db

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

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

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("chat.db"), &gorm.Config{})
	if err != nil {
		panic("Не удалось подключиться к базе")
	}
	fmt.Println("База подключена")
	DB.AutoMigrate(&User{}, &Message{})
}
