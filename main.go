package main

import (
	"context"
	"database/sql"
	_ "embed"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"

	"github.com/connorkuehl/factoid/internal/repo/sqlite"
	"github.com/connorkuehl/factoid/internal/service"
)

func main() {
	var config struct {
		readAddr   string
		writeAddr  string
		sqlitePath string
	}

	flag.StringVar(&config.readAddr, "api-public", ":8080", "public API address")
	flag.StringVar(&config.writeAddr, "api-private", ":8081", "private API address")
	flag.StringVar(&config.sqlitePath, "db-sqlite", ":memory:", "path to SQLite DB")
	flag.Parse()

	logger := log.With().Str("component", "service").Logger()
	logger.Info().Str("db-sqlite", config.sqlitePath).Msg("properties")

	reg := prometheus.NewRegistry()
	metrics := newMetrics(reg)

	db, _ := sql.Open("sqlite", config.sqlitePath)
	defer db.Close()

	if config.sqlitePath == ":memory:" {
		if _, err := db.Exec(sqlite.Schema()); err != nil {
			panic(err)
		}
	}

	service := service.New(
		logger,
		sqlite.NewRepo(db, metrics),
	)

	mux := service.Routes()
	mux.Handler(http.MethodGet, "/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	server := http.Server{
		Addr:    config.readAddr,
		Handler: withMetrics(metrics, mux),
	}

	serve := func(s *http.Server, name string) {
		logger := log.With().
			Str(name+"_addr", s.Addr).
			Logger()
		logger.Info().Msg("listening")
		logger.Error().Err(s.ListenAndServe()).Msg("")
	}

	go serve(&server, "http")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, os.Interrupt)
	defer cancel()

	<-ctx.Done()

	log.Info().Msg("attempting graceful shutdown, send SIGINT again to cancel")

	ctx, cancel = signal.NotifyContext(context.Background(), syscall.SIGINT, os.Interrupt)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	server.Shutdown(ctx)

	log.Info().Err(ctx.Err()).Msg("shut down")
}
