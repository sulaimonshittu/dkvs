package transaction

import (
	"bufio"
	"fmt"
	"os"
)

type FileTransactionLogger struct {
	events       chan<- Event
	errors       <-chan error
	lastSequence uint64
	file         *os.File
}

func NewFileTransactionLogger(filename string) (TransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %w", err)
	}

	return &FileTransactionLogger{file: file}, nil
}

func (log *FileTransactionLogger) WritePut(key, value string) {
	log.events <- Event{
		EventType: EventPut,
		Key:       key,
		Value:     value,
	}
}

func (log *FileTransactionLogger) WriteDelete(key string) {
	log.events <- Event{
		EventType: EventDelete,
		Key:       key,
	}
}

func (log *FileTransactionLogger) Err() <-chan error {
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
				log.lastSequence, event.EventType, event.Key, event.Value,
			)
			if err != nil {
				errs <- err
				return
			}
		}
	}()
	log.file.Close()
}

func (log *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	scanner := bufio.NewScanner(log.file)
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		var e Event

		defer close(outEvent)
		defer close(outError)

		for scanner.Scan() {
			line := scanner.Text()

			if _, err := fmt.Sscanf(line, "%d\t%d\t%s\t%s", &e.Sequence, &e.EventType, &e.Key, &e.Value); err != nil {
				outError <- fmt.Errorf("input parse error: %w", err)
				return
			}
			if log.lastSequence >= e.Sequence {
				outError <- fmt.Errorf("transaction numbers out of sequence")
				return
			}
			log.lastSequence = e.Sequence
			outEvent <- e
		}
		if scanner.Err() != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", scanner.Err())
			return
		}
	}()
	return outEvent, outError
}
