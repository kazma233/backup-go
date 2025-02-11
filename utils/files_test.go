package utils

import (
	"log"
	"testing"
)

func Test_zipPath(t *testing.T) {
	path, err := ZipPath(`E:\audio\asmr`, "test.zip")
	if err != nil {
		panic(err)
	}

	log.Printf("path %s", path)
}
