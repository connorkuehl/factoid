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

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"

	"github.com/connorkuehl/factoid/repo/sqlite"
	"github.com/connorkuehl/factoid/service"
)

//go:embed repo/sqlite/schema.sql
var schema string

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

	db, _ := sql.Open("sqlite", config.sqlitePath)
	defer db.Close()

	if config.sqlitePath == ":memory:" {
		if _, err := db.Exec(schema); err != nil {
			panic(err)
		}
	}

	service := service.New(
		logger,
		sqlite.NewRepo(db),
	)

	reg := prometheus.NewRegistry()
	metrics := newMetrics(reg)

	public := httprouter.New()
	public.HandlerFunc(http.MethodGet, "/v1/facts", service.FactsHandler)
	public.HandlerFunc(http.MethodGet, "/v1/fact/:id", service.FactHandler)

	private := httprouter.New()
	private.HandlerFunc(http.MethodPost, "/v1/facts", service.FactsHandler)
	private.HandlerFunc(http.MethodDelete, "/v1/fact/:id", service.FactHandler)
	private.Handler(http.MethodGet, "/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	publicServer := http.Server{
		Addr:    config.readAddr,
		Handler: withMetrics(metrics, public),
	}

	privateServer := http.Server{
		Addr:    config.writeAddr,
		Handler: withMetrics(metrics, private),
	}

	serve := func(s *http.Server, name string) {
		logger := log.With().
			Str(name+"_addr", s.Addr).
			Logger()
		logger.Info().Msg("listening")
		logger.Error().Err(s.ListenAndServe()).Msg("")
	}

	go serve(&publicServer, "http_public")
	go serve(&privateServer, "http_private")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, os.Interrupt)
	defer cancel()

	<-ctx.Done()

	log.Info().Msg("attempting graceful shutdown, send SIGINT again to cancel")

	ctx, cancel = signal.NotifyContext(context.Background(), syscall.SIGINT, os.Interrupt)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	privateServer.Shutdown(ctx)
	publicServer.Shutdown(ctx)

	log.Info().Err(ctx.Err()).Msg("shut down")
}
