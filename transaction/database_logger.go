package transaction

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresTransactionLogger struct {
	events chan<- Event
	errors <-chan error
	db     *sql.DB
}

type PostgresConfig struct {
	DbName   string
	Host     string
	User     string
	Password string
}

func NewPostgresTransactionLogger(config PostgresConfig) (TransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s", config.Host, config.DbName, config.User, config.Password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to open db: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	logger := &PostgresTransactionLogger{db: db}

	exists, err := logger.verifyTableExists()
	if err != nil {
		return nil, fmt.Errorf("failed to verify table exists: %w", err)
	}

	if !exists {
		if err = logger.createTable(); err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}

	return logger, nil
}

func (log *PostgresTransactionLogger) WritePut(key, value string) {
	log.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (log *PostgresTransactionLogger) WriteDelete(key string) {
	log.events <- Event{EventType: EventDelete, Key: key}
}

func (log *PostgresTransactionLogger) Err() <-chan error {
	return log.errors
}

func (log *PostgresTransactionLogger) Run() {
	events := make(chan Event, 20)
	log.events = events

	errs := make(chan error, 1)
	log.errors = errs

	go func() {
		query := `INSERT INTO transactions
			(event_type, key, value)
			VALUES ($1, $2, $3)`

		for event := range events {
			_, err := log.db.Exec(query, event.EventType, event.Key, event.Value)
			if err != nil {
				errs <- err
			}
		}
	}()
}

func (log *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		defer close(outEvent)
		defer close(outError)

		query := `SELECT sequence, event_type, key, value
				FROM transactions
				ORDER BY sequence`

		rows, err := log.db.Query(query)
		if err != nil {
			outError <- fmt.Errorf("sql query error: %w", err)
			return
		}

		defer rows.Close()
		e := Event{}

		for rows.Next() {
			err = rows.Scan(&e.Sequence, &e.EventType, &e.Key, &e.Value)

			if err != nil {
				outError <- fmt.Errorf("error reading row: %w", err)
				return
			}

			outEvent <- e
		}
		err = rows.Err()
		if err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
		}
	}()

	return outEvent, outError
}

func (log *PostgresTransactionLogger) verifyTableExists() (bool, error) {
	tableName := "transactions"
	query := `SELECT to_regclass($1) IS NOT NULL;`

	var exists bool
	err := log.db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("Failed to verify table exists: %w", err)
	}

	return true, nil
}

func (log *PostgresTransactionLogger) createTable() error {
	query := `
	CREATE TABLE transactions (
		sequence INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		event_type SMALLINT NOT NULL UNIQUE,
		key  NOT NULL,
		value  NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	err, _ := log.db.Exec(query)
	if err != nil {
		return fmt.Errorf("Failed to create table: %w", err)
	}
	return nil
}
