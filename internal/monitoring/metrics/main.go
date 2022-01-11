package metrics

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var ()

type MonitoringServer struct {
	wg  *sync.WaitGroup
	srv *http.Server
}

func StartMonitoringServer(logger *zap.Logger, cfg Config) (*MonitoringServer, error) {
	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)
	srv := startHttpServer(logger, cfg, httpServerExitDone)
	return &MonitoringServer{
		wg:  httpServerExitDone,
		srv: srv,
	}, nil
}

func (m *MonitoringServer) Stop() error {
	context, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := m.srv.Shutdown(context); err != nil {
		return err
	}

	m.wg.Wait()

	return nil
}

func startHttpServer(logger *zap.Logger, cfg Config, wg *sync.WaitGroup) *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port)}

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		defer wg.Done()

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	return srv
}
