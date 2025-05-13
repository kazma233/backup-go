package utils

import (
	"log"
	"testing"
)

func Test_zipPath(t *testing.T) {
	path, err := ZipPath(`D:\program\idea`, "test.zip", func(filePath string, processed, total int64, percentage float64) {
		log.Printf("zip %s: %d/%d (%.2f%%)", filePath, processed, total, percentage)
	}, func(total int64) {
		log.Printf("zip done, total: %d", total)
	})
	if err != nil {
		panic(err)
	}

	log.Printf("path %s", path)
}
