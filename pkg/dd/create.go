package dd

import (
	"log"
	"os"
	"path/filepath"
)

func CreateDalleDress(series, address, filePath string) {
	// We protect ourselves from being overwritten if the user asks for the same image while
	// we're building it.
	lockFilePath := filepath.Join("./pending", series+"-"+address+".lck")
	if _, err := os.Create(lockFilePath); err != nil {
		log.Printf("Failed to create lock file: %v", err)
		return
	}
	defer os.Remove(lockFilePath)

	// Make sure we have a place to store the new image
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		log.Printf("Failed to create directory: %v", err)
		return
	}

	maker := NewMaker(series)
	maker.GenerateImage(address)
	// time.Sleep(1 * time.Second)
	// sourcePath := "./input/default.png"
	// srcFile, err := os.Open(sourcePath)
	// if err != nil {
	// 	log.Printf("Failed to open source file: %v", err)
	// 	return
	// }
	// defer srcFile.Close()
	// dstFile, err := os.Create(filePath)
	// if err != nil {
	// 	log.Printf("Failed to create destination file: %v", err)
	// 	return
	// }
	// defer dstFile.Close()
	// if _, err := io.Copy(dstFile, srcFile); err != nil {
	// 	log.Printf("Failed to copy file: %v", err)
	// 	return
	// }
	// log.Printf("File created: %s", filePath)
}
