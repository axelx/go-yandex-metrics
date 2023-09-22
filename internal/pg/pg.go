package pg

import (
	"context"
	"errors"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/axelx/go-yandex-metrics/internal/models"
)

const (
	defaultMaxConnections = 10
)

type PgStorage struct {
	DB             *sqlx.DB
	maxConnections int
	RetryIntervals []time.Duration
	logger         *zap.Logger
}

func NewDBStorage(log *zap.Logger) *PgStorage {
	return &PgStorage{
		logger:         log,
		RetryIntervals: []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
}

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

func (c *PgStorage) GetDBMetric(typeMetric models.MetricType, nameMetric string) (models.Metrics, error) {
	err := errors.New("не найдена метрика")
	mt := models.Metrics{}
	switch typeMetric {
	case models.MetricGauge:
		row := c.DB.QueryRowContext(context.Background(), ` SELECT value FROM gauge WHERE name = $1`, nameMetric)

		var value float64
		err = row.Scan(&value)
		if err != nil {
			c.logger.Error("Error GetDBMetricс gauge:", zap.String("about ERR", err.Error()))
			return mt, err
		}

		mt = models.Metrics{MType: typeMetric, ID: nameMetric, Value: &value}
		return mt, nil
	case models.MetricCounter:
		row := c.DB.QueryRowContext(context.Background(), ` SELECT delta FROM counter WHERE name = $1`, nameMetric)

		var delta int64
		err = row.Scan(&delta)
		if err != nil {
			c.logger.Error("Error GetDBMetric сounter", zap.String("about ERR", err.Error()))
			return mt, err
		}
		mt = models.Metrics{MType: typeMetric, ID: nameMetric, Delta: &delta}
		return mt, nil
	default:
		return mt, err
	}
}

func (c *PgStorage) SetDBMetric(typeMetric models.MetricType, nameMetric string, value *float64, delta *int64) error {

	err := errors.New("не найдена метрика")
	switch typeMetric {
	case models.MetricGauge:
		_, err := c.DB.ExecContext(context.Background(),
			`INSERT INTO gauge (name, value) VALUES ($1, $2)
					ON CONFLICT (name) DO UPDATE SET value = $2;`, nameMetric, value)
		if err != nil {
			return err
		}

		return nil
	case models.MetricCounter:
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
		case models.MetricGauge:
			_, err := tx.ExecContext(ctx,
				"INSERT INTO gauge (name, value) VALUES ($1, $2) "+
					" ON CONFLICT (name) DO UPDATE SET value = $2", v.ID, v.Value)
			if err != nil {
				c.logger.Error("Error SetBatchMetrics gauge:", zap.String("about ERR", err.Error()))
				tx.Rollback()
				return err
			}
		case models.MetricCounter:
			_, err := tx.ExecContext(ctx,
				`INSERT INTO counter (name, delta) VALUES ($1, $2)
						ON CONFLICT (name) DO UPDATE SET delta = counter.delta +  $2;`, v.ID, v.Delta)
			if err != nil {
				c.logger.Error("Error SetBatchMetrics counter:", zap.String("about ERR", err.Error()))
				tx.Rollback()
				return err
			}
		default:
		}
	}
	return tx.Commit()
}

func (c *PgStorage) GetDBMetrics(typeMetric models.MetricType) interface{} {
	res := map[string]interface{}{}
	var metric struct {
		Name  string
		Value float64
		Delta int64
	}
	switch typeMetric {
	case models.MetricGauge:
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
	case models.MetricCounter:
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
