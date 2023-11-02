// Модуль metrics собирает метрики системы в рантайме и отправляет их по установленному урлу
package metrics

import (
	"context"
	"google.golang.org/grpc/encoding/gzip"
	"log"
	"time"

	"google.golang.org/grpc"

	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/models"
	pb "github.com/axelx/go-yandex-metrics/internal/proto"
	"github.com/axelx/go-yandex-metrics/internal/service"
	"google.golang.org/grpc/credentials/insecure"
)

func sendRequestMetricGRPC(ctx context.Context, addr string, metrics []models.Metrics) error {

	// Set up a connection to the server.
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("gRPC agent sendRequestMetricGRPC: did not connect: grpc.Dial, ", "about ERR"+err.Error())
		return err
	}
	defer conn.Close()
	c := pb.NewMetricsClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	for _, m := range metrics {

		mG := &pb.UpdateMetricRequest{Metric: &pb.Metric{
			ID:    m.ID,
			MType: string(m.MType),
			Delta: service.UnPointer(m.Delta),
			Value: service.UnPointer(m.Value),
		}}
		updateM(ctx, c, mG)
	}
	return nil
}

func sendRequestMetricsGRPC(ctx context.Context, addr string, metrics []models.Metrics) error {

	// Set up a connection to the server.
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("gRPC agent sendRequestMetricGRPC: did not connect: grpc.Dial, ", "about ERR"+err.Error())
		return err
	}
	defer conn.Close()
	c := pb.NewMetricsClient(conn)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	ms := []*pb.Metric{}

	for _, m := range metrics {
		ms = append(ms, &pb.Metric{
			ID:    m.ID,
			MType: string(m.MType),
			Delta: service.UnPointer(m.Delta),
			Value: service.UnPointer(m.Value),
		})
	}
	mss := &pb.UpdateMetricsRequest{Metric: ms}
	updateMS(ctx, c, mss)

	return nil
}

func getM(ctx context.Context, c pb.MetricsClient, in *pb.GetMetricRequest) error {

	r, err := c.GetMetric(ctx, in)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return err
	}

	logger.Info("gRPC agent ", "getM  r.Metric.ID"+r.Metric.ID)
	return nil
}

func updateM(ctx context.Context, c pb.MetricsClient, in *pb.UpdateMetricRequest) error {
	compressor := grpc.UseCompressor(gzip.Name)
	r, err := c.UpdateMetric(ctx, in, compressor)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return err
	}
	logger.Info("gRPC agent ", "updateM  r.Metric.ID"+r.Metric.ID)
	return nil
}

func updateMS(ctx context.Context, c pb.MetricsClient, in *pb.UpdateMetricsRequest) error {
	compressor := grpc.UseCompressor(gzip.Name)
	_, err := c.UpdateMetrics(ctx, in, compressor)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return err
	}
	logger.Info("gRPC agent ", "updateMS")
	return nil

}
