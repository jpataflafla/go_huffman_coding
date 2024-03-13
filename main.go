package main

import (
	"log"
)

func main() {
	db, err := NewSimplePostgressDB()
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewApiServer(":8080", db)
	server.Run()
}
