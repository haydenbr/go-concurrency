package main

import (
	"log"
	"os"

	"golang.org/x/sync/errgroup"
)

func main() {
	processFiles([]string{
		"text1.txt",
		"text2.txt",
		"text3.txt",
		"error.txt",
	})
}

func processFiles(files []string) {
	var wg = errgroup.Group{}

	for _, file := range files {
		path := file

		wg.Go(func() error {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			log.Print(string(data))
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		log.Print(err)
	}
}
