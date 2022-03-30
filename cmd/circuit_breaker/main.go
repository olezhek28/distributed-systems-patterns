package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Circuit func(context.Context) (string, error)

const (
	baseTimeout      = time.Second
	failureThreshold = 3
)

func main() {
	circuit := Breaker(
		func(ctx context.Context) (string, error) {
			return "", errors.New("some error")
			// return "good result", nil
		},
		failureThreshold,
		baseTimeout,
	)

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		time.Sleep(3 * time.Second)
		res, err := circuit(ctx)
		if err != nil {
			fmt.Println("failed to do:", err, "i=", i)
			continue
		}
		fmt.Println("res=", res, "i=", i)
	}
}

func Breaker(circuit Circuit, failureThreshold uint, baseTimeout time.Duration) Circuit {
	failuresCount := 0
	lastAttempt := time.Now()
	var m sync.RWMutex

	return func(ctx context.Context) (string, error) {
		m.RLock()

		diff := failuresCount - int(failureThreshold)

		if diff >= 0 {
			shouldRetryAt := lastAttempt.Add(baseTimeout << diff)
			if !time.Now().After(shouldRetryAt) {
				m.RUnlock()
				return "", errors.New("service unreachable")
			}
		}

		m.RUnlock()

		res, err := circuit(ctx)

		m.Lock()
		defer m.Unlock()

		lastAttempt = time.Now()

		if err != nil {
			failuresCount++
			return "", err
		}

		failuresCount = 0
		return res, nil
	}
}
