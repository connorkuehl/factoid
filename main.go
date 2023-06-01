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

	log "golang.org/x/exp/slog"
	_ "modernc.org/sqlite"

	"github.com/connorkuehl/factoid/internal/repo/sqlite"
	"github.com/connorkuehl/factoid/internal/service"
)

func main() {
	var config struct {
		addr       string
		sqlitePath string
		auth       string
	}

	flag.StringVar(&config.addr, "addr", ":8080", "address to listen on")
	flag.StringVar(&config.sqlitePath, "db-sqlite", ":memory:", "path to SQLite DB")
	flag.StringVar(&config.auth, "authorization", "", "secret for write-operations, disabled by default!")
	flag.Parse()

	logger := log.With("component", "service")
	logger.With("db-sqlite", config.sqlitePath).Info("")

	db, _ := sql.Open("sqlite", config.sqlitePath)
	defer db.Close()

	if config.sqlitePath == ":memory:" {
		if _, err := db.Exec(sqlite.Schema()); err != nil {
			panic(err)
		}
	}

	service := service.New(
		sqlite.NewRepo(db),
	)

	mux := service.Routes()

	server := http.Server{
		Addr:    config.addr,
		Handler: mux,
	}

	serve := func(s *http.Server) {
		logger := log.With("http_addr", s.Addr)
		logger.Info("listening")
		if err := s.ListenAndServe(); err != nil {
			logger.With("err", err).Error("")
		}
	}

	go serve(&server)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, os.Interrupt)
	defer cancel()

	<-ctx.Done()

	log.Info("attempting graceful shutdown, send SIGINT again to cancel")

	ctx, cancel = signal.NotifyContext(context.Background(), syscall.SIGINT, os.Interrupt)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.With("err", server.Shutdown(ctx)).Error("")
	}
	log.Info("reached shutdown")
}
