package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

	receiveOrdersChan, _ := receiveOrders(rawOrders)
	validOrderChan, invalidOrderChan := validateOrders(receiveOrdersChan)

	wg.Add(1)
	processOrders(validOrderChan, invalidOrderChan, func() { wg.Done() })
	wg.Wait()
}

func receiveOrders(rawOrders []string) (<-chan order, <-chan error) {
	out := make(chan order)
	errChan := make(chan error)

	go func() {
		for _, rawOrder := range rawOrders {
			var newOrder order
			err := json.Unmarshal([]byte(rawOrder), &newOrder)
			if err != nil {
				errChan <- err
			} else {
				out <- newOrder
			}
		}

		close(out)
		close(errChan)
	}()

	return out, errChan
}

func validateOrders(inChan <-chan order) (<-chan order, <-chan invalidOrder) {
	validOrderChan := make(chan order)
	invalidOrderChan := make(chan invalidOrder)

	go func() {
		for order := range inChan {
			if order.Quantity <= 0 {
				invalidOrderChan <- invalidOrder{order: order, err: errors.New("quantity must be greater than zero")}
			} else {
				validOrderChan <- order
			}
		}

		close(validOrderChan)
		close(invalidOrderChan)
	}()

	return validOrderChan, invalidOrderChan
}

func processOrders(
	validOrderChan <-chan order,
	invalidOrderChan <-chan invalidOrder,
	onComplete func(),
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

	onComplete()
}
