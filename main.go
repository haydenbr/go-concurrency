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
	reservedOrderChan := reserveInventory(validOrderChan)
	filledOrdersChan := fillOrders(reservedOrderChan)

	wg.Add(1)
	go processRecords(processRecordsParams[invalidOrder]{
		Records: invalidOrderChan,
		ProcessRecord: func(o invalidOrder) {
			fmt.Printf("invalid order received: %v, error: %v\n", o.order, o.err)
		},
		OnComplete: wg.Done,
	})

	wg.Add(1)
	go processRecords(processRecordsParams[order]{
		Records: filledOrdersChan,
		ProcessRecord: func(o order) {
			fmt.Printf("order filled: %v\n", o)
		},
		OnComplete: wg.Done,
	})

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

func reserveInventory(in <-chan order) <-chan order {
	out := make(chan order)

	go func() {
		for o := range in {
			o.Status = reserved
			out <- o
		}
		close(out)
	}()

	return out
}

func fillOrders(in <-chan order) <-chan order {
	out := make(chan order)

	go processRecords(processRecordsParams[order]{
		Records: in,
		ProcessRecord: func(o order) {
			o.Status = filled
			out <- o
		},
		OnComplete: func() { close(out) },
	})

	return out
}

type processRecordsParams[TRecord any] struct {
	Records       <-chan TRecord
	ProcessRecord func(order TRecord)
	OnComplete    func()
}

func processRecords[TRecord any](params processRecordsParams[TRecord]) {
	for record := range params.Records {
		params.ProcessRecord(record)
	}

	params.OnComplete()
}
