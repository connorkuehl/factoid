package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"time"

	"github.com/connorkuehl/factoid/internal/promlabels"
	"github.com/connorkuehl/factoid/internal/service"
)

//go:embed schema.sql
var schema string

func Schema() string {
	return schema
}

type Metrics interface {
	UpstreamResponsesInc(component promlabels.Upstream, status promlabels.RequestStatus)
	UpstreamRequestsInc(component promlabels.Upstream)
	UpstreamRequestLatency(component promlabels.Upstream, status promlabels.RequestStatus, latency time.Duration)
}

type Repo struct {
	db      *sql.DB
	metrics Metrics
}

func NewRepo(db *sql.DB, metrics Metrics) *Repo {
	return &Repo{
		db:      db,
		metrics: metrics,
	}
}

func (r *Repo) Facts(ctx context.Context) ([]service.Fact, error) {
	r.metrics.UpstreamRequestsInc(promlabels.UpstreamRepo)
	start := time.Now()
	status := promlabels.RequestSuccess

	defer func() {
		r.metrics.UpstreamRequestLatency(promlabels.UpstreamRepo, status, time.Since(start))
		r.metrics.UpstreamResponsesInc(promlabels.UpstreamRepo, status)
	}()

	db := New(r.db)
	result, err := db.GetFacts(ctx)
	if err != nil {
		status = promlabels.RequestFail
		return nil, ErrToDomainErr(err)
	}

	facts := make([]service.Fact, 0, len(result))
	for _, f := range result {
		facts = append(facts, ModelToDomain(f))
	}

	return facts, nil
}

func (r *Repo) Fact(ctx context.Context, id int64) (service.Fact, error) {
	r.metrics.UpstreamRequestsInc(promlabels.UpstreamRepo)
	start := time.Now()
	status := promlabels.RequestSuccess

	defer func() {
		r.metrics.UpstreamRequestLatency(promlabels.UpstreamRepo, status, time.Since(start))
		r.metrics.UpstreamResponsesInc(promlabels.UpstreamRepo, status)
	}()

	db := New(r.db)
	result, err := db.GetFact(ctx, id)
	if err != nil {
		status = promlabels.RequestFail
	}
	return ModelToDomain(result), ErrToDomainErr(err)
}

func (r *Repo) RandomFact(ctx context.Context) (service.Fact, error) {
	r.metrics.UpstreamRequestsInc(promlabels.UpstreamRepo)
	start := time.Now()
	status := promlabels.RequestSuccess

	defer func() {
		r.metrics.UpstreamRequestLatency(promlabels.UpstreamRepo, status, time.Since(start))
		r.metrics.UpstreamResponsesInc(promlabels.UpstreamRepo, status)
	}()

	db := New(r.db)
	result, err := db.GetRandomFact(ctx)
	if err != nil {
		status = promlabels.RequestFail
	}

	return ModelToDomain(result), ErrToDomainErr(err)
}

func (r *Repo) CreateFact(ctx context.Context, content, source string) (service.Fact, error) {
	r.metrics.UpstreamRequestsInc(promlabels.UpstreamRepo)
	start := time.Now()
	status := promlabels.RequestSuccess

	defer func() {
		r.metrics.UpstreamRequestLatency(promlabels.UpstreamRepo, status, time.Since(start))
		r.metrics.UpstreamResponsesInc(promlabels.UpstreamRepo, status)
	}()

	db := New(r.db)
	result, err := db.CreateFact(ctx, CreateFactParams{
		Content: content,
		Source:  sql.NullString{String: source, Valid: true},
	})
	if err != nil {
		status = promlabels.RequestFail
	}

	return ModelToDomain(result), ErrToDomainErr(err)
}

func (r *Repo) DeleteFact(ctx context.Context, id int64) error {
	r.metrics.UpstreamRequestsInc(promlabels.UpstreamRepo)
	start := time.Now()
	status := promlabels.RequestSuccess

	defer func() {
		r.metrics.UpstreamRequestLatency(promlabels.UpstreamRepo, status, time.Since(start))
		r.metrics.UpstreamResponsesInc(promlabels.UpstreamRepo, status)
	}()

	db := New(r.db)
	err := db.DeleteFact(ctx, id)
	if err != nil {
		status = promlabels.RequestFail
	}

	return ErrToDomainErr(err)
}

func ModelToDomain(f Fact) service.Fact {
	return service.Fact{
		ID:        f.ID,
		CreatedAt: f.CreatedAt.Time,
		UpdatedAt: f.UpdatedAt.Time,
		DeletedAt: f.DeletedAt.Time,
		Content:   f.Content,
		Source:    f.Source.String,
	}
}

func ErrToDomainErr(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return service.ErrNotFound
	}
	return err
}
