package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	receiveOrdersChan := make(chan order)
	validOrderChan := make(chan order)
	invalidOrderChan := make(chan invalidOrder)
	go receiveOrders(receiveOrdersChan)
	go validateOrders(receiveOrdersChan, validOrderChan, invalidOrderChan)

	wg.Add(1)
	go func(validOrderChan <-chan order) {
		order := <-validOrderChan
		fmt.Printf("valid order received: %v\n", order)
		wg.Done()
	}(validOrderChan)
	go func(invalidOrderChan <-chan invalidOrder) {
		order := <-invalidOrderChan
		fmt.Printf("inalid order received: %v, error: %v\n", order.order, order.err)
		wg.Done()
	}(invalidOrderChan)
	wg.Wait()
}

func validateOrders(inChan <-chan order, outChan chan<- order, errChan chan<- invalidOrder) {
	order := <-inChan
	if order.Quantity <= 0 {
		errChan <- invalidOrder{order: order, err: errors.New("quantity must be greater than zero")}
	} else {
		outChan <- order
	}
}

func receiveOrders(out chan<- order) {
	for _, rawOrder := range rawOrders {
		var newOrder order
		err := json.Unmarshal([]byte(rawOrder), &newOrder)
		if err != nil {
			log.Print(err)
			continue
		}
		out <- newOrder
	}
}

var rawOrders = []string{
	`{"productCode": 1111, "quantity": 5, "status": 1}`,
	`{"productCode": 2222, "quantity": 42.3, "status": 1}`,
	`{"productCode": 3333, "quantity": 19, "status": 1}`,
	`{"productCode": 4444, "quantity": 8, "status": 1}`,
}
