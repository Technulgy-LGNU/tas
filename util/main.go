package util

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

// This little program finds all the localhost urls in the
// project and replaces them with the production urls

func main() {
	log.SetFlags(log.LstdFlags & log.Lshortfile)
	var (
		urls = map[string]string{
			"localhost:3001": "api.technulgy.com",
		}

		err error
	)

	fmt.Println("Starting the URL replacement program")
	fmt.Println("Urls to replace:")
	for k, v := range urls {
		fmt.Printf("Localhost: %s -> Production: %s\n", k, v)
	}

	// Replace the urls
	// Get all the files in the project
	files, err := os.ReadDir(os.Args[1])
	if err != nil {
		log.Fatalf("Error reading the directory: %d\n", err)
	}

	for _, file := range files {
		// Read the file
		data, err := os.ReadFile(file.Name())
		if err != nil {
			log.Fatalf("Error reading the file: %s\n", file.Name())
		}

		// Replace the urls
		for k, v := range urls {
			data = bytes.ReplaceAll(data, []byte(k), []byte(v))
		}
	}
}
