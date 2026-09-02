package main

import (
	"jobFlow/database"
	"log"
)

func main() {
	err := database.InitDB("database.db")
	if err != nil {
		log.Println(err)
	}
}
