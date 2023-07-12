package pg

import (
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

const (
	defaultMaxConnections = 10
)

type Client struct {
	db             *sqlx.DB
	maxConnections int
}

func NewClient() *Client {
	return &Client{
		maxConnections: defaultMaxConnections,
	}
}

func (c *Client) Open(source string) error {
	var err error

	fmt.Println("connecting to db")
	c.db, err = sqlx.Connect("pgx", source)
	if err != nil {
		return err
	}

	err = c.db.Ping()
	if err != nil {
		fmt.Println("sql ping failed")
		return err
	}
	fmt.Println("connected to db")
	c.db.SetMaxOpenConns(c.maxConnections)
	c.db.SetMaxIdleConns(c.maxConnections)

	return nil
}

// Close closes PostgreSQL connection.
func (c *Client) Close() error {
	fmt.Println("connection to db closed")
	return c.db.Close()
}
