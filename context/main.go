package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		tick := time.Tick(500 * time.Millisecond)

		for range tick {
			fmt.Println("tick!")

			if ctx.Err() != nil {
				log.Println(ctx.Err())
				return
			}
		}
	}(ctx)

	wg.Wait()
}
