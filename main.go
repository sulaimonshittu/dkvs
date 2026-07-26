package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/go-chi/chi/v5"
)

type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)
}

type FileTransactionLogger struct {
	events       chan<- Event
	errors       <-chan error
	lastSequence uint64
	file         *os.File
}

func (log *FileTransactionLogger) WritePut(key, value string) {
	log.events <- Event{
		EventType: EventPut,
		key:       key,
		value:     value,
	}
}

func (log *FileTransactionLogger) WriteDelete(key string) {
	log.events <- Event{
		EventType: EventDelete,
		key:       key,
	}
}

func (log *FileTransactionLogger) err() <-chan error {
	return log.errors
}

func (log *FileTransactionLogger) Run() {
	events := make(chan Event, 20)
	log.events = events

	errs := make(chan error, 1)
	log.errors = errs

	go func() {
		for event := range events {
			log.lastSequence++
			_, err := fmt.Fprintf(
				log.file,
				"%d\t%d\t%s\t%s\n",
				log.lastSequence, event.EventType, event.key, event.value,
			)
			if err != nil {
				errs <- err
				return
			}
		}
	}()
}

func NewFileTransactionLogger(filename string) (TransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %w", err)
	}

	return &FileTransactionLogger{file: file}, nil
}

type EventType int

const (
	_ EventType = iota
	EventPut
	EventDelete
)

type Event struct {
	Sequence uint64
	EventType
	key   string
	value string
}

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
	r.Delete("v1/key/{key}", deleteHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
