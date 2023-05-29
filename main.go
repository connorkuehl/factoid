package main

import (
	"context"
	"database/sql"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"

	"github.com/connorkuehl/factoid/internal/repo/sqlite"
	"github.com/connorkuehl/factoid/internal/service"
)

func main() {
	var config struct {
		addr       string
		sqlitePath string
	}

	flag.StringVar(&config.addr, "addr", ":8080", "address to listen on")
	flag.StringVar(&config.sqlitePath, "db-sqlite", ":memory:", "path to SQLite DB")
	flag.Parse()

	logger := log.With().Str("component", "service").Logger()
	logger.Info().Str("db-sqlite", config.sqlitePath).Msg("properties")

	db, _ := sql.Open("sqlite", config.sqlitePath)
	defer db.Close()

	if config.sqlitePath == ":memory:" {
		if _, err := db.Exec(sqlite.Schema()); err != nil {
			panic(err)
		}
	}

	service := service.New(
		logger,
		sqlite.NewRepo(db),
	)

	mux := service.Routes()

	server := http.Server{
		Addr:    config.addr,
		Handler: mux,
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
