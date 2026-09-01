package main

import (
	"jobFlow/database"
	"log"
)

func main() {
	err := database.InitDB()
	if err != nil {
		log.Println(err)
	}
}
