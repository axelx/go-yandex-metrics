package pg

import (
	"context"
	"errors"
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/models"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

const (
	defaultMaxConnections = 10
)

type PgStorage struct {
	DB             *sqlx.DB
	maxConnections int
}

func NewClient() *PgStorage {
	return &PgStorage{
		maxConnections: defaultMaxConnections,
	}
}

func (c *PgStorage) Open(source string) error {
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
func (c *PgStorage) Close() error {
	fmt.Println("connection to db closed")
	return c.DB.Close()
}

// Close closes PostgreSQL connection.
func (c *PgStorage) CreateTable() error {
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

func (c *PgStorage) GetDBMetric(typeMetric, nameMetric string) (models.Metrics, error) {
	err := errors.New("не найдена метрика")
	mt := models.Metrics{}
	switch typeMetric {
	case "gauge":
		row := c.DB.QueryRowContext(context.Background(), ` SELECT value FROM gauge WHERE name = $1`, nameMetric)

		var value float64
		err = row.Scan(&value)
		if err != nil {
			return mt, err
		}

		mt = models.Metrics{MType: typeMetric, ID: nameMetric, Value: &value}
		return mt, nil
	case "counter":
		row := c.DB.QueryRowContext(context.Background(), ` SELECT delta FROM counter WHERE name = $1`, nameMetric)

		var delta int64
		err = row.Scan(&delta)
		if err != nil {
			fmt.Println("panic err", err)
			return mt, err
		}
		mt = models.Metrics{MType: typeMetric, ID: nameMetric, Delta: &delta}
		return mt, nil
	default:
		return mt, err
	}
}

func (c *PgStorage) SetDBMetric(typeMetric, nameMetric string, value *float64, delta *int64) error {

	err := errors.New("не найдена метрика")
	switch typeMetric {
	case "gauge":
		_, err := c.DB.ExecContext(context.Background(),
			`INSERT INTO gauge (name, value) VALUES ($1, $2)
					ON CONFLICT (name) DO UPDATE SET value = $2;`, nameMetric, value)
		if err != nil {
			return err
		}

		return nil
	case "counter":
		_, err := c.DB.ExecContext(context.Background(),
			`INSERT INTO counter (name, delta) VALUES ($1, $2)
					ON CONFLICT (name) DO UPDATE SET delta = counter.delta +  $2;`, nameMetric, delta)
		if err != nil {
			return err
		}

		return nil
	default:
		return err
	}
}
func (c *PgStorage) SetBatchMetrics(metrics []models.Metrics) error {
	ctx := context.Background()

	tx, err := c.DB.Begin()
	if err != nil {
		return err
	}
	for _, v := range metrics {
		switch v.MType {
		case "gauge":
			fmt.Println(v, "gauge")
			_, err := tx.ExecContext(ctx,
				"INSERT INTO gauge (name, value) VALUES ($1, $2) "+
					" ON CONFLICT (name) DO UPDATE SET value = $2", v.ID, v.Value)
			if err != nil {
				fmt.Println(v, "gauge err", err)
				tx.Rollback()

				return err
			}
		case "counter":
			fmt.Println(v, "counter")
			_, err := tx.ExecContext(ctx,
				`INSERT INTO counter (name, delta) VALUES ($1, $2)
						ON CONFLICT (name) DO UPDATE SET delta = counter.delta +  $2;`, v.ID, v.Delta)
			if err != nil {
				fmt.Println(v, "counter err", err)
				tx.Rollback()
				return err
			}
		default:
		}
	}
	return tx.Commit()
}

func (c *PgStorage) GetDBMetrics(typeMetric string) interface{} {
	res := map[string]interface{}{}
	var metric struct {
		Name  string
		Value float64
		Delta int64
	}
	switch typeMetric {
	case "gauge":
		rows, err := c.DB.QueryContext(context.Background(), ` SELECT name, value FROM gauge`)
		if err != nil {
			return nil
		}

		for rows.Next() {
			err = rows.Scan(&metric.Name, &metric.Value)
			if err != nil {
				return nil
			}
			res[metric.Name] = metric.Value
		}
		err = rows.Err()
		if err != nil {
			return nil
		}

		return res
	case "counter":
		rows, err := c.DB.QueryContext(context.Background(), ` SELECT name, delta FROM counter`)
		if err != nil {
			return nil
		}

		for rows.Next() {
			err = rows.Scan(&metric.Name, &metric.Delta)
			if err != nil {
				return nil
			}
			res[metric.Name] = metric.Delta
		}
		err = rows.Err()
		if err != nil {
			return nil
		}
		return res
	default:
		return nil
	}
}
