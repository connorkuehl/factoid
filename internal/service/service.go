package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	log "golang.org/x/exp/slog"
)

type Option interface {
	Apply(s *Service)
}

type optionFunc func(s *Service)

func (opt optionFunc) Apply(s *Service) {
	opt(s)
}

func WithAuthorizer(auth string) optionFunc {
	return func(s *Service) { s.auth = auth }
}

type FactRepo interface {
	Facts(context.Context) ([]Fact, error)
	Fact(ctx context.Context, id int64) (Fact, error)
	RandomFact(context.Context) (Fact, error)
	CreateFact(ctx context.Context, contents, source string) (Fact, error)
	DeleteFact(ctx context.Context, id int64) error
}

type Service struct {
	facts FactRepo
	auth  string
}

func New(f FactRepo, opts ...Option) *Service {
	s := &Service{facts: f}
	for _, opt := range opts {
		opt.Apply(s)
	}
	return s
}

func (s *Service) FactsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		facts, err := s.facts.Facts(context.Background())
		if err != nil {
			log.Error("", "err", err)
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
			log.With("err", err).Error("")
			s.RespondErrorJSON(w, http.StatusBadRequest, errors.New("bad request"))
			return
		}

		if body.Content == "" {
			s.RespondErrorJSON(w, http.StatusBadRequest, errors.New("content field missing or blank"))
			return
		}

		f, err := s.facts.CreateFact(context.Background(), body.Content, body.Source)
		if err != nil {
			log.With(
				"create_fact_content", body.Content,
				"create_fact_source", body.Source,
				"err", err,
			).Error("")
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

	logger := log.With(
		"request_uri", r.RequestURI,
		"http_method", r.Method,
		"get_fact_param_id", idParam,
	)

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
				logger.With("err", err).Error("")

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
				logger.With("err", err).Error("")

				status = http.StatusInternalServerError
				err = errors.New("internal error")
			}
			s.RespondErrorJSON(w, status, err)
			return
		}

		err = s.facts.DeleteFact(context.Background(), id)
		if err != nil {
			logger.With("err", err).Error("")

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
