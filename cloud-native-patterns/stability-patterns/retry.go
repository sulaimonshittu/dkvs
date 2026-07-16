package main

import (
	"context"
	"time"
)

type Effector func(context.Context) (string, error)

func Retry(effector Effector, maxRetries int, delay time.Duration) Effector {
	return func(ctx context.Context) (string, error) {
		for r := 0; ; r++ {
			response, err := effector(ctx)
			if err == nil || r >= maxRetries {
				return response, err
			}
		}
	}
}
