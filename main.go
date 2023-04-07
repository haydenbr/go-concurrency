package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

var rawOrders = []string{
	`{"productCode": 1111, "quantity": -5, "status": 1}`,
	`{"productCode": 2222, "quantity": 42.3, "status": 1}`,
	`{"productCode": 3333, "quantity": 19, "status": 1}`,
	`{"productCode": 4444, "quantity": 8, "status": 1}`,
}

func main() {
	var wg sync.WaitGroup

	receiveOrdersChan, _ := receiveOrders(rawOrders)
	validOrderChan, invalidOrderChan := validateOrders(receiveOrdersChan)
	reservedOrderChan := reserveInventory(validOrderChan)
	filledOrdersWg := fillOrders(reservedOrderChan)

	wg.Add(1)
	go processRecords(processRecordsParams[invalidOrder]{
		Records: invalidOrderChan,
		ProcessRecord: func(o invalidOrder) {
			fmt.Printf("invalid order received: %v, error: %v\n", o.order, o.err)
		},
		OnComplete: wg.Done,
	})

	filledOrdersWg.Wait()
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

	go processRecords(processRecordsParams[order]{
		Records: inChan,
		ProcessRecord: func(o order) {
			if o.Quantity <= 0 {
				invalidOrderChan <- invalidOrder{order: o, err: errors.New("quantity must be greater than zero")}
			} else {
				validOrderChan <- o
			}
		},
		OnComplete: func() {
			close(validOrderChan)
			close(invalidOrderChan)
		},
	})

	return validOrderChan, invalidOrderChan
}

func reserveInventory(in <-chan order) <-chan order {
	out := make(chan order)
	var wg sync.WaitGroup

	const workers = 3
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go processRecords(processRecordsParams[order]{
			Records: in,
			ProcessRecord: func(o order) {
				o.Status = reserved
				out <- o
			},
			OnComplete: wg.Done,
		})
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func fillOrders(in <-chan order) *sync.WaitGroup {
	var wg sync.WaitGroup
	const workers = 3

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go processRecords(processRecordsParams[order]{
			Records: in,
			ProcessRecord: func(o order) {
				o.Status = filled
				fmt.Printf("order filled: %v\n", o)
			},
			OnComplete: wg.Done,
		})
	}

	return &wg
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
