package db

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

const (
	defaultMaxConnections = 10
)

type DB struct {
	DB             *sqlx.DB
	maxConnections int
	RetryIntervals []time.Duration
}

func NewClient() *DB {
	return &DB{
		maxConnections: defaultMaxConnections,
		RetryIntervals: []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
}

func (c *DB) Open(source string) error {
	var err error

	fmt.Println("connecting to db")
	c.DB, err = sqlx.Connect("pgx", source)
	if err != nil {
		return err
	}

	err = c.DB.Ping()
	if err != nil {
		fmt.Println("sql ping failed")
		return err
	}
	fmt.Println("connected to db")
	c.DB.SetMaxOpenConns(c.maxConnections)
	c.DB.SetMaxIdleConns(c.maxConnections)

	return nil
}

// Close closes PostgreSQL connection.
func (c *DB) Close() error {
	fmt.Println("connection to db closed")
	return c.DB.Close()
}

// Close closes PostgreSQL connection.
func (c *DB) CreateTable() error {
	_, err := c.DB.ExecContext(context.Background(),
		` CREATE TABLE IF NOT EXISTS gauge (
					 id serial PRIMARY KEY,
					 name varchar(450) NOT NULL UNIQUE,
					 value double precision NOT NULL DEFAULT '0.00'
				);
				CREATE TABLE IF NOT EXISTS counter (
					id serial PRIMARY KEY,
					name varchar(450) NOT NULL UNIQUE,
					delta bigint NOT NULL DEFAULT '0'
				);
			`)
	return err
}
