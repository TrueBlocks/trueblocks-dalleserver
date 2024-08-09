package dd

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var Waiting = 10

func CreateDalleDress(series, address, filePath string) {
	lockFilePath := filepath.Join("./pending", series+"-"+address+".lck")
	if _, err := os.Create(lockFilePath); err != nil {
		log.Printf("Failed to create lock file: %v", err)
		return
	}
	defer os.Remove(lockFilePath)
	for {
		Waiting--
		if Waiting <= 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	sourcePath := "./input/default.png"
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		log.Printf("Failed to create directory: %v", err)
		return
	}
	srcFile, err := os.Open(sourcePath)
	if err != nil {
		log.Printf("Failed to open source file: %v", err)
		return
	}
	defer srcFile.Close()
	dstFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("Failed to create destination file: %v", err)
		return
	}
	defer dstFile.Close()
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		log.Printf("Failed to copy file: %v", err)
		return
	}
	log.Printf("File created: %s", filePath)
	Waiting = 10
}
