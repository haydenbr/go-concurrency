package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
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
		// done := ctx.Done()
		// loop:
		// 	for {
		// 		select {
		// 		case <-tick:
		// 			fmt.Println("tick!")
		// 		case <-done:
		// 			wg.Done()
		// 			break loop
		// 		}
		// 	}
	}(ctx)

	time.Sleep(2 * time.Second)
	cancel()

	wg.Wait()
}
