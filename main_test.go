package main

import (
	"log"
	"testing"
)

func Test_zipPath(t *testing.T) {
	path, err := zipPath(`E:\audio\asmr`)
	if err != nil {
		panic(err)
	}

	log.Printf("path %s", path)
}

func Test_backup(t *testing.T) {
	oc := CreateOSSClient()
	backup(`E:\audio\asmr`, oc)
}

func Test_cleanOld(t *testing.T) {
	oc := CreateOSSClient()
	cleanOld(oc)
}
