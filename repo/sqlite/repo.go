package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/connorkuehl/factoid/service"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db}
}

func (r *Repo) Facts(ctx context.Context) ([]service.Fact, error) {
	db := New(r.db)
	result, err := db.GetFacts(ctx)
	if err != nil {
		return nil, ErrToDomainErr(err)
	}

	facts := make([]service.Fact, 0, len(result))
	for _, f := range result {
		facts = append(facts, ModelToDomain(f))
	}

	return facts, nil
}

func (r *Repo) Fact(ctx context.Context, id int64) (service.Fact, error) {
	db := New(r.db)
	result, err := db.GetFact(ctx, id)
	return ModelToDomain(result), ErrToDomainErr(err)
}

func (r *Repo) RandomFact(ctx context.Context) (service.Fact, error) {
	db := New(r.db)
	result, err := db.GetRandomFact(ctx)
	return ModelToDomain(result), ErrToDomainErr(err)
}

func (r *Repo) CreateFact(ctx context.Context, content, source string) (service.Fact, error) {
	db := New(r.db)
	result, err := db.CreateFact(ctx, CreateFactParams{
		Content: content,
		Source:  sql.NullString{String: source, Valid: true},
	})
	return ModelToDomain(result), ErrToDomainErr(err)
}

func (r *Repo) DeleteFact(ctx context.Context, id int64) error {
	db := New(r.db)
	return ErrToDomainErr(db.DeleteFact(ctx, id))
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
