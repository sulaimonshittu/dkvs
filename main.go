package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
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

func putHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	value, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	err = Put(key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
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

func getHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	value, err := Get(key)
	if errors.Is(err, ErrorNonExistentKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

func Delete(key string) error {
	store.Lock()
	defer store.Unlock()
	delete(store.m, key)
	return nil
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	err := Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	r := chi.NewRouter()
	r.Put("/v1/key/{key}", putHandler)
	r.Get("/v1/key/{key}", getHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
