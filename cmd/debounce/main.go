package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Circuit func(context.Context) (string, error)

const (
	duration = time.Second
)

func main() {
	ctx := context.Background()
	debounce := Debounce(
		func(context.Context) (string, error) {
			return "blabla", nil
		},
		duration,
	)

	for i := 0; i < 10; i++ {
		res, err := debounce(ctx)
		if err != nil {
			fmt.Printf("failed to get value: %s\n", err.Error())
			continue
		}

		fmt.Println(res)
		time.Sleep(300 * time.Millisecond)
	}
}

func Debounce(circuit Circuit, duration time.Duration) Circuit {
	var res string
	var err error
	var threshold time.Time
	var m sync.Mutex

	return func(ctx context.Context) (string, error) {
		m.Lock()
		defer m.Unlock()

		if time.Now().Before(threshold) {
			return res, err
		}

		fmt.Println("try get new value ...")
		res, err = circuit(ctx)
		threshold = time.Now().Add(duration)

		return res, err
	}
}
