package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
)

type FactRepo interface {
	Facts(context.Context) ([]Fact, error)
	Fact(ctx context.Context, id int64) (Fact, error)
	RandomFact(context.Context) (Fact, error)
	CreateFact(ctx context.Context, contents, source string) (Fact, error)
	DeleteFact(ctx context.Context, id int64) error
}

type Service struct {
	logger zerolog.Logger
	facts  FactRepo
}

func New(l zerolog.Logger, f FactRepo) *Service {
	return &Service{
		logger: l,
		facts:  f,
	}
}

func (s *Service) FactsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		facts, err := s.facts.Facts(context.Background())
		if err != nil {
			s.logger.Error().Err(err).Msg("")
			s.RespondErrorJSON(w, http.StatusInternalServerError, err)
			return
		}

		s.RespondJSON(w, http.StatusOK, map[string]any{"facts": facts})

	case http.MethodPost:
		var body struct {
			Content string `json:"content"`
			Source  string `json:"source"`
		}

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			s.logger.Error().Err(err).Msg("")
			s.RespondErrorJSON(w, http.StatusBadRequest, errors.New("bad request"))
			return
		}

		if body.Content == "" {
			s.RespondErrorJSON(w, http.StatusBadRequest, errors.New("content field missing or blank"))
			return
		}

		logger := s.logger.With().
			Str("create_fact_content", body.Content).
			Str("create_fact_source", body.Source).
			Logger()

		f, err := s.facts.CreateFact(context.Background(), body.Content, body.Source)
		if err != nil {
			logger.Error().Err(err).Msg("")
			s.RespondErrorJSON(w, http.StatusInternalServerError, errors.New("internal error"))
			return
		}

		s.RespondJSON(w, http.StatusCreated, map[string]any{"fact": f})
		return
	}
}

func (s *Service) FactHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	idParam := params.ByName("id")

	logger := s.logger.With().
		Str("request_uri", r.RequestURI).
		Str("http_method", r.Method).
		Str("get_fact_param_id", idParam).
		Logger()

	var id int64
	var err error
	id, err = strconv.ParseInt(idParam, 10, 64)

	switch r.Method {
	case http.MethodGet:
		getFact := func(ctx context.Context) (Fact, error) {
			return s.facts.Fact(ctx, id)
		}

		if idParam == "rand" {
			getFact = func(ctx context.Context) (Fact, error) {
				return s.facts.RandomFact(ctx)
			}
			err = nil
		}

		if err != nil {
			s.RespondErrorJSON(w, http.StatusBadRequest, errors.New("id must be an integer or 'rand'"))
			return
		}

		f, err := getFact(context.Background())
		if err != nil {
			status := http.StatusNotFound
			if !errors.Is(err, ErrNotFound) {
				logger.Error().Err(err).Msg("")

				status = http.StatusInternalServerError
				err = errors.New("internal error")
			}
			s.RespondErrorJSON(w, status, err)
			return
		}

		s.RespondJSON(w, http.StatusOK, map[string]any{"fact": f})

	case http.MethodDelete:
		// This belongs to the strconv.ParseInt call above the switch statement.
		if err != nil {
			s.RespondErrorJSON(w, http.StatusBadRequest, errors.New("id must be an integer"))
			return
		}

		ctx := context.Background()

		// Since s.facts.SoftDeleteFact doesn't return a count of
		// rows affected.
		_, err = s.facts.Fact(ctx, id)
		if err != nil {
			status := http.StatusNotFound
			if !errors.Is(err, ErrNotFound) {
				logger.Error().Err(err).Msg("")

				status = http.StatusInternalServerError
				err = errors.New("internal error")
			}
			s.RespondErrorJSON(w, status, err)
			return
		}

		err = s.facts.DeleteFact(context.Background(), id)
		if err != nil {
			logger.Error().Err(err).Msg("")

			err = errors.New("internal error")
			s.RespondErrorJSON(w, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Service) unimplemented(w http.ResponseWriter, _ *http.Request) {
	s.RespondErrorJSON(w, http.StatusNotImplemented, errors.New("not implemented"))
}
