package main

import (
	"errors"
	"sync"
)

type LockMap struct {
	sync.RWMutex
	m map[string]string
}

var ErrorNonExistentKey = errors.New("key doesn't exist")

var store = LockMap{
	m: map[string]string{},
}

func Put(key, value string) error {
	store.Lock()
	defer store.Unlock()
	store.m[key] = value

	return nil
}

func Get(key string) (string, error) {
	store.RLock()
	defer store.RUnlock()
	value, ok := store.m[key]
	if !ok {
		return "", ErrorNonExistentKey
	}
	return value, nil
}

func Delete(key string) error {
	store.Lock()
	defer store.Unlock()
	delete(store.m, key)
	return nil
}
