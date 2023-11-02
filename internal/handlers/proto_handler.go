package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"

	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"github.com/axelx/go-yandex-metrics/internal/pg"
	pb "github.com/axelx/go-yandex-metrics/internal/proto"
	"github.com/axelx/go-yandex-metrics/internal/service"
)

// ProtoHandler data for gRPC server
type ProtoHandler struct {
	pb.UnimplementedMetricsServer
	DB         *sqlx.DB
	DBPostgres *pg.PgStorage
	Addr       string
}

func PBNew(db *sqlx.DB, NewDBStorage *pg.PgStorage, addr string) ProtoHandler {
	return ProtoHandler{
		DB:         db,
		DBPostgres: NewDBStorage,
		Addr:       addr,
	}
}

// ProtoHandler  GetMetric implements GetMetric
func (s *ProtoHandler) GetMetric(ctx context.Context, in *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	mType := models.MetricType(in.MType)
	metric, err := s.DBPostgres.GetDBMetric(mType, in.ID)

	var response pb.GetMetricResponse

	fmt.Println("grpc getMetric", metric, err)

	logger.Info("gRPC getMetrice:", "_")

	d := service.UnPointer(metric.Delta)
	v := service.UnPointer(metric.Value)

	response.Metric = &pb.Metric{ID: in.ID, MType: in.MType, Delta: d, Value: v}

	log.Printf("Received ID: %s, MType: %s", in.ID, in.MType)
	logger.Info("gRPC getMetrice:", "_")

	return &response, nil
}

// ProtoHandler  GetMetric implements UpdateMetric
func (s *ProtoHandler) UpdateMetric(ctx context.Context, in *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	var response pb.UpdateMetricResponse

	mType := models.MetricType(in.Metric.MType)
	metric, err := s.DBPostgres.GetDBMetric(mType, in.Metric.ID)
	if err != nil {
		logger.Error("gRPC UpdateMetric, s.DBPostgres.GetDBMetric:", "about ERR"+err.Error())
	}

	err = s.DBPostgres.SetDBMetric(
		mType,
		in.Metric.ID,
		service.Float64ToPointerFloat64(in.Metric.Value),
		service.Int64ToPointerInt64(in.Metric.Delta),
	)

	metric, err = s.DBPostgres.GetDBMetric(mType, in.Metric.ID)

	d := service.UnPointer(metric.Delta)
	v := service.UnPointer(metric.Value)

	response.Metric = &pb.Metric{ID: metric.ID, MType: in.Metric.MType, Delta: d, Value: v}

	logger.Info("gRPC UpdateMetric", "")
	return &response, nil
}

// ProtoHandler  GetMetric implements UpdateMetrics
func (s *ProtoHandler) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	var response pb.UpdateMetricsResponse
	ms := []models.Metrics{}

	for _, metric := range in.Metric {
		ms = append(ms, models.Metrics{
			ID:    metric.ID,
			MType: models.MetricType(metric.MType),
			Delta: service.ToPointer(metric.Delta),
			Value: service.ToPointer(metric.Value),
		})
	}

	err := s.DBPostgres.SetBatchMetrics(ms)
	if err != nil {
		logger.Error("gRPC UpdateMetric, s.DBPostgres.SetBatchMetrics:", "about ERR"+err.Error())
	}

	return &response, nil
}
