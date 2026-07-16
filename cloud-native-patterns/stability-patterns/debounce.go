package main

import (
	"context"
	"sync"
	"time"
)

func DebounceFirst(circuit Circuit, d time.Duration) Circuit {
	var threshold time.Time
	var result string
	var err error
	var m sync.Mutex

	return func(ctx context.Context) (string, error) {
		m.Lock()
		defer m.Unlock()

		if time.Now().Before(threshold) {
			return result, err
		}
		result, err = circuit(ctx)
		threshold = time.Now().Add(d)

		return result, err
	}
}

func DebounceFirstContext(circuit Circuit, d time.Duration) Circuit {
	var threshold time.Time
	var m sync.Mutex
	var lastCtx context.Context
	var lastCancel context.CancelFunc

	return func(ctx context.Context) (string, error) {
		m.Lock()

		if time.Now().Before(threshold) {
			lastCancel()
		}
		lastCtx, lastCancel = context.WithCancel(ctx)
		threshold = time.Now().Add(d)

		m.Unlock()

		result, err := circuit(lastCtx)
		return result, err
	}
}

func DebounceLast(circuit Circuit, d time.Duration) Circuit {
	var m sync.Mutex
	var timer *time.Timer
	var cctx context.Context
	var cancel context.CancelFunc

	return func(ctx context.Context) (string, error) {
		m.Lock()

		if timer != nil {
			timer.Stop()
			cancel()
		}

		cctx, cancel = context.WithCancel(ctx)
		ch := make(chan struct {
			result string
			err    error
		}, 1)
		timer = time.AfterFunc(d, func() {
			r, e := circuit(cctx)
			ch <- struct {
				result string
				err    error
			}{result: r, err: e}
		})
		m.Unlock()

		select {
		case res := <-ch:
			return res.result, res.err
		case <-cctx.Done():
			return "", cctx.Err()
		}
	}
}
