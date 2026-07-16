package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Circuit func(ctx context.Context) (string, error)

func Breaker(circuit Circuit, threshold int) Circuit {
	var failures int
	var last = time.Now()
	var m sync.RWMutex

	return func(ctx context.Context) (string, error) {
		m.RLock()
		d := failures - threshold
		if d >= 0 {
			shouldRetryAt := last.Add((2 << d) * time.Second)
			if !time.Now().After(shouldRetryAt) {
				m.RUnlock()
				return "", errors.New("service unavailable")
			}
		}
		m.RUnlock()

		response, err := circuit(ctx)
		m.Lock()
		defer m.Unlock()

		last = time.Now()
		if err != nil {
			failures++
			return response, err
		}

		failures = 0
		return response, nil
	}
}
