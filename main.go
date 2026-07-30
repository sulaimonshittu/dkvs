package main

import (
	"dkvs/transaction"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var logger transaction.TransactionLogger

func initializeTransactionLog() error {
	var err error

	logger, err := transaction.NewPostgresTransactionLogger(transaction.PostgresConfig{
		Host:     "localhost",
		DbName:   "db-name",
		User:     "db-user",
		Password: "db-password",
	})

	if err != nil {
		return fmt.Errorf("failed to create event logger: %w", err)
	}
	events, errors := logger.ReadEvents()
	e := transaction.Event{}
	ok := true

	for ok && err == nil {
		select {
		case err, ok = <-errors:
		case e, ok = <-events:
			switch e.EventType {
			case transaction.EventDelete:
				err = Delete(e.Key)
			case transaction.EventPut:
				err = Put(e.Key, e.Value)
			}
		}
	}

	logger.Run()
	return err
}

func main() {
	err := initializeTransactionLog()
	if err != nil {
		log.Println(fmt.Errorf("Failed to initiate the Trasaction log: %w", err))
	}

	r := chi.NewRouter()
	r.Put("/v1/key/{key}", putHandler)
	r.Get("/v1/key/{key}", getHandler)
	r.Delete("v1/key/{key}", deleteHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
