package main

import "errors"

var ErrorNonExistentKey = errors.New("key doesn't exist")

var store = map[string]string{}

func Put(key, value string) error {
	store[key] = value

	return nil
}

func Get(key string) (string, error) {
	value, ok := store[key]
	if !ok {
		return "", ErrorNonExistentKey
	}
	return value, nil
}

func Delete(key string) error {
	delete(store, key)
	return nil
}

func main() {

}
