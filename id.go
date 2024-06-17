package main

import (
	"log"
	"os"
)

var ID string

func InitID() {
	ID = getId()
	log.Printf("ID is %s", ID)
}

func getId() string {
	id := Config.ID
	if id != "" {
		return id
	}

	hostname, err := os.Hostname()
	if err != nil {
		panic("id not found")
	}

	return hostname
}
