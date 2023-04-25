package main

import (
	"log"
	"os"
	"sync"
)

func main() {
	waitGroup([]string{
		"text1.txt",
		"text2.txt",
		"text3.txt",
		"error.txt",
	})
}

func waitGroup(files []string) {
	var wg sync.WaitGroup

	for _, file := range files {
		path := file
		wg.Add(1)
		go func() {
			defer wg.Done()

			data, err := os.ReadFile(path)
			if err != nil {
				log.Print(err)
			} else {
				log.Print(string(data))
			}
		}()
	}

	wg.Wait()
}
