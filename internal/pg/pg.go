package pg

import (
	"context"
	"errors"
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"github.com/axelx/go-yandex-metrics/internal/pg/db"
	_ "github.com/jackc/pgx/stdlib"
	"time"
)

type PgStorage struct {
	Client *db.DB
	//maxConnections int
	RetryIntervals []time.Duration
}

func NewDBStorage(clientDB *db.DB) *PgStorage {
	return &PgStorage{
		Client:         clientDB,
		RetryIntervals: []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
}

func (c *PgStorage) GetDBMetric(typeMetric models.MetricType, nameMetric string) (models.Metrics, error) {
	fmt.Println("GetDBMetric:")
	err := errors.New("не найдена метрика")
	mt := models.Metrics{}
	switch typeMetric {
	case models.MetricGauge:
		row := c.Client.DB.QueryRowContext(context.Background(), ` SELECT value FROM gauge WHERE name = $1`, nameMetric)
		fmt.Println("GetDBMetric: gauge:", nameMetric)

		var value float64
		err = row.Scan(&value)
		if err != nil {
			fmt.Println("err GetDBMetricс g:", err)
			return mt, err
		}

		mt = models.Metrics{MType: typeMetric, ID: nameMetric, Value: &value}
		return mt, nil
	case models.MetricCounter:
		row := c.Client.DB.QueryRowContext(context.Background(), ` SELECT delta FROM counter WHERE name = $1`, nameMetric)
		fmt.Println("GetDBMetric: counter:", nameMetric)

		var delta int64
		err = row.Scan(&delta)
		if err != nil {
			fmt.Println(" err  GetDBMetric c:", err)
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
		_, err := c.Client.DB.ExecContext(context.Background(),
			`INSERT INTO gauge (name, value) VALUES ($1, $2)
					ON CONFLICT (name) DO UPDATE SET value = $2;`, nameMetric, value)
		if err != nil {
			return err
		}

		return nil
	case models.MetricCounter:
		_, err := c.Client.DB.ExecContext(context.Background(),
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

	tx, err := c.Client.DB.Begin()
	if err != nil {
		return err
	}
	for _, v := range metrics {
		switch v.MType {
		case models.MetricGauge:
			fmt.Println(v, *v.Value, "gauge")
			_, err := tx.ExecContext(ctx,
				"INSERT INTO gauge (name, value) VALUES ($1, $2) "+
					" ON CONFLICT (name) DO UPDATE SET value = $2", v.ID, v.Value)
			if err != nil {
				fmt.Println(v, "gauge err", err)
				tx.Rollback()

				return err
			}
		case models.MetricCounter:
			fmt.Println(v, *v.Delta, "counter")
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

func (c *PgStorage) GetDBMetrics(typeMetric models.MetricType) interface{} {
	res := map[string]interface{}{}
	var metric struct {
		Name  string
		Value float64
		Delta int64
	}
	switch typeMetric {
	case models.MetricGauge:
		rows, err := c.Client.DB.QueryContext(context.Background(), ` SELECT name, value FROM gauge`)
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
		rows, err := c.Client.DB.QueryContext(context.Background(), ` SELECT name, delta FROM counter`)
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
