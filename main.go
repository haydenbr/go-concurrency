package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
)

var rawOrders = []string{
	`{"productCode": 1111, "quantity": 5, "status": 1}`,
	`{"productCode": 2222, "quantity": 42.3, "status": 1}`,
	`{"productCode": 3333, "quantity": 19, "status": 1}`,
	`{"productCode": 4444, "quantity": 8, "status": 1}`,
}

func main() {
	var wg sync.WaitGroup
	receiveOrdersChan := make(chan order)
	validOrderChan := make(chan order)
	invalidOrderChan := make(chan invalidOrder)
	go receiveOrders(receiveOrdersChan, rawOrders)
	go validateOrders(receiveOrdersChan, validOrderChan, invalidOrderChan)

	wg.Add(1)
	/*
		wait group keeps the process running until we're done
		normally, we might have an infinite loop or something in the main routine
	*/

	go func(
		validOrderChan <-chan order,
		invalidOrderChan <-chan invalidOrder,
	) {
	loop:
		for {
			select {
			case order, ok := <-validOrderChan:
				if ok {
					fmt.Printf("valid order received: %v\n", order)
				} else {
					break loop
				}
			case invalidOrder, ok := <-invalidOrderChan:
				if ok {
					fmt.Printf("inalid order received: %v, error: %v\n", invalidOrder.order, invalidOrder.err)
				} else {
					break loop
				}
			}
		}

		wg.Done()
	}(validOrderChan, invalidOrderChan)

	wg.Wait()
}

func receiveOrders(out chan<- order, rawOrders []string) {
	for _, rawOrder := range rawOrders {
		var newOrder order
		err := json.Unmarshal([]byte(rawOrder), &newOrder)
		if err != nil {
			log.Print(err)
			continue
		}
		out <- newOrder
	}
	close(out)
}

func validateOrders(inChan <-chan order, outChan chan<- order, errChan chan<- invalidOrder) {
	for order := range inChan {
		if order.Quantity <= 0 {
			errChan <- invalidOrder{order: order, err: errors.New("quantity must be greater than zero")}
		} else {
			outChan <- order
		}
	}
	close(outChan)
	close(errChan)
}
