// server для получения и отображения метрик которые отправляет агент
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"

	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/handlers"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/pg"
	"github.com/axelx/go-yandex-metrics/internal/proto"
	"github.com/axelx/go-yandex-metrics/internal/storage"
)

var errDB = errors.New("")

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version:= %s\n", buildVersion)
	fmt.Printf("Build date:= %s\n", buildDate)
	fmt.Printf("Build commit:= %s\n", buildCommit)

	conf := config.NewConfigServer()
	if err := logger.Initialize(conf.LogLevel); err != nil {
		fmt.Println(err)
	}
	logger.Info("Running server", "address"+conf.String())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metricStorage := storage.New(conf.FileStoragePath, conf.StoreInternal, conf.Restore)

	//подключаемся к базе
	NewDBStorage := pg.NewDBStorage()
	NewDBStorage.DB, errDB = sqlx.Connect("pgx", conf.DatabaseDSN)
	if errDB != nil {
		logger.Error("Error not connect to db", "about ERR"+errDB.Error())
		go metricStorage.UpdateFile(ctx)
		if conf.FileStoragePath != "" {
			metricStorage.RestoreFromFile()
		}
	} else {
		NewDBStorage.DB.SetMaxOpenConns(10)

		defer func() {
			NewDBStorage.DB.Close()
		}()

		if NewDBStorage.DB != nil {
			NewDBStorage.CreateTable()
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	gRPCsrv := handlers.PBNew(NewDBStorage.DB, NewDBStorage, ":50051")
	go grpcServer(gRPCsrv)

	hd := handlers.New(&metricStorage, NewDBStorage.DB, NewDBStorage, conf.HashKey, conf.CryptoKey, conf.TrustedSubnet)
	var srv = http.Server{
		Addr:    conf.RunAddr,
		Handler: hd.Router(),
	}

	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigint
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error("HTTP server Shutdown:", "about ERR"+errDB.Error())
		}
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logger.Error("HTTP server ListenAndServe:", "about ERR"+errDB.Error())
	}

	<-idleConnsClosed
	NewDBStorage.DB.Close()
	wg.Done()
	wg.Wait()

	fmt.Println("Server Shutdown gracefully")
}

func grpcServer(gsrv handlers.ProtoHandler) {
	lis, err := net.Listen("tcp", gsrv.Addr)
	if err != nil {
		logger.Error("gRPC failed to listen:", "about ERR"+errDB.Error())
	}
	s := grpc.NewServer()
	go_yandex_metrics.RegisterMetricsServer(s, &gsrv)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		logger.Error("gRPC failed to serve:", "about ERR"+errDB.Error())
	}
}
