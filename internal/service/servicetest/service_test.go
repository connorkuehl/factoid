package servicetest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"

	"github.com/connorkuehl/factoid/internal/promlabels"
	sqliterepo "github.com/connorkuehl/factoid/internal/repo/sqlite"
	"github.com/connorkuehl/factoid/internal/service"
)

type metricsStub struct{}

func (m metricsStub) UpstreamResponsesInc(component promlabels.Upstream, status promlabels.RequestStatus) {
}

func (m metricsStub) UpstreamRequestsInc(component promlabels.Upstream) {}

func (m metricsStub) UpstreamRequestLatency(component promlabels.Upstream, status promlabels.RequestStatus, latency time.Duration) {
}

func newTestDB(t *testing.T, facts ...service.Fact) (*sqliterepo.Repo, func()) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(sqliterepo.Schema())
	if err != nil {
		t.Fatal(err)
	}

	repo := sqliterepo.NewRepo(db, metricsStub{})

	for _, fact := range facts {
		_, err := repo.CreateFact(context.TODO(), fact.Content, fact.Source)
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
	}

	return repo, func() { db.Close() }
}

func makeFactTuples(facts ...service.Fact) map[string]string {
	m := make(map[string]string)
	for _, fact := range facts {
		m[fact.Content] = fact.Source
	}
	return m
}

func TestGetFacts(t *testing.T) {
	tests := []struct {
		name        string
		preexisting []service.Fact
		wantStatus  int
		wantFacts   []service.Fact
	}{
		{
			name:       "none",
			wantStatus: http.StatusOK,
		},
		{
			name:        "one",
			wantStatus:  http.StatusOK,
			preexisting: []service.Fact{{Content: "a fun fact", Source: "a unit test"}},
			wantFacts:   []service.Fact{{Content: "a fun fact", Source: "a unit test"}},
		},
		{
			name:       "many",
			wantStatus: http.StatusOK,
			preexisting: []service.Fact{
				{Content: "1", Source: "Source 1"},
				{Content: "2", Source: "Source 2"},
				{Content: "3", Source: "Source 3"},
			},
			wantFacts: []service.Fact{
				{Content: "1", Source: "Source 1"},
				{Content: "2", Source: "Source 2"},
				{Content: "3", Source: "Source 3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, cleanup := newTestDB(t, tt.preexisting...)
			defer cleanup()

			svc := service.New(log.Logger, r)

			ts := httptest.NewServer(svc.Routes())
			defer ts.Close()

			rsp, err := ts.Client().Get(ts.URL + "/v1/facts")
			if err != nil {
				t.Fatal(err)
			}
			defer rsp.Body.Close()

			if rsp.StatusCode != tt.wantStatus {
				t.Fatalf("want http %d, got http %d",
					tt.wantStatus, rsp.StatusCode)
			}

			var response struct {
				Facts []service.Fact `json:"facts"`
			}

			err = json.NewDecoder(rsp.Body).Decode(&response)
			if err != nil {
				t.Fatal(err)
			}

			wantTuples := makeFactTuples(tt.wantFacts...)
			gotTuples := makeFactTuples(response.Facts...)

			if !reflect.DeepEqual(wantTuples, gotTuples) {
				t.Fatalf("want facts %v, got facts %v",
					wantTuples, gotTuples)
			}
		})
	}
}

func TestPostFactsInputValidation(t *testing.T) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	tests := []struct {
		name       string
		inputJSON  string
		wantStatus int
		wantError  string
	}{
		{
			name:       "bad JSON",
			inputJSON:  `{"content}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "bad request",
		},
		{
			name:       "missing content field",
			inputJSON:  `{"source": "the Internet"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "content field missing or blank",
		},
		{
			name:       "empty content field",
			inputJSON:  `{"content": ""}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "content field missing or blank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, cleanup := newTestDB(t)
			defer cleanup()

			svc := service.New(log.Logger, r)

			ts := httptest.NewServer(svc.Routes())
			defer ts.Close()

			rsp, err := ts.Client().Post(ts.URL+"/v1/facts", "application/json", strings.NewReader(tt.inputJSON))
			if err != nil {
				t.Fatal(err)
			}
			defer rsp.Body.Close()

			if rsp.StatusCode != tt.wantStatus {
				t.Fatalf("want http %d, got http %d",
					tt.wantStatus, rsp.StatusCode)
			}

			var errMessage errorResponse
			err = json.NewDecoder(rsp.Body).Decode(&errMessage)
			if err != nil {
				t.Fatal(err)
			}

			if errMessage.Error != tt.wantError {
				t.Fatalf("want err %q, got err %q",
					tt.wantError, errMessage.Error)
			}
		})
	}
}

func TestPostFacts(t *testing.T) {
	r, cleanup := newTestDB(t)
	defer cleanup()

	svc := service.New(log.Logger, r)

	ts := httptest.NewServer(svc.Routes())
	defer ts.Close()

	body := `{"content": "unit tests can be helpful", "source": "my opinion"}`
	rsp, err := ts.Client().Post(ts.URL+"/v1/facts", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer rsp.Body.Close()

	wantCode := http.StatusCreated
	if rsp.StatusCode != wantCode {
		t.Fatalf("want http %d, got http %d",
			wantCode, rsp.StatusCode)
	}

	var response struct {
		Fact service.Fact `json:"fact"`
	}
	if err := json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	wantContent := "unit tests can be helpful"

	if response.Fact.Content != wantContent {
		t.Fatalf("want content %q, got content %q",
			wantContent, response.Fact.Content)
	}

	wantSource := "my opinion"

	if response.Fact.Source != wantSource {
		t.Fatalf("want source %q, got source %q",
			wantSource, response.Fact.Source)
	}

	// And finally, fetch the exact fact that was just created to
	// ensure it's able to read what it's written.

	rsp, err = ts.Client().Get(fmt.Sprintf("%s/v1/fact/%d", ts.URL, response.Fact.ID))
	if err != nil {
		t.Fatal(err)
	}
	defer rsp.Body.Close()

	returned := response.Fact
	if err := json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response.Fact != returned {
		t.Fatalf("want %+v, got %+v", returned, response.Fact)
	}
}

func TestGetFactInputValidation(t *testing.T) {
	tests := []struct {
		name    string
		inputID string
		wantErr string
	}{
		{
			name:    "non-integer, and not 'rand'",
			inputID: "asdf",
			wantErr: "id must be an integer or 'rand'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, cleanup := newTestDB(t)
			defer cleanup()

			svc := service.New(log.Logger, r)

			ts := httptest.NewServer(svc.Routes())
			defer ts.Close()

			rsp, err := ts.Client().Get(fmt.Sprintf("%s/v1/fact/%s", ts.URL, tt.inputID))
			if err != nil {
				t.Fatal(err)
			}
			defer rsp.Body.Close()

			var errRsp struct {
				Error string `json:"error"`
			}

			if err := json.NewDecoder(rsp.Body).Decode(&errRsp); err != nil {
				t.Fatal(err)
			}

			if tt.wantErr != errRsp.Error {
				t.Fatalf("want error %q, got error %q",
					tt.wantErr, errRsp.Error)
			}
		})
	}
}

func TestGetFact(t *testing.T) {
	tests := []struct {
		name        string
		inputID     string
		preexisting []service.Fact
		wantFact    service.Fact
		wantErr     string
		wantStatus  int
	}{
		{
			name:        "exists",
			preexisting: []service.Fact{{Content: "a fact", Source: "a source"}},
			inputID:     "1",
			wantFact:    service.Fact{Content: "a fact", Source: "a source"},
			wantStatus:  http.StatusOK,
		},
		{
			name:       "does not exist",
			inputID:    "2",
			wantStatus: http.StatusNotFound,
			wantErr:    "not found",
		},
		{
			name:        "rand",
			preexisting: []service.Fact{{Content: "a fact", Source: "a source"}},
			inputID:     "rand",
			wantFact:    service.Fact{Content: "a fact", Source: "a source"},
			wantStatus:  http.StatusOK,
		},
		{
			name:       "rand against empty table",
			inputID:    "rand",
			wantStatus: http.StatusNotFound,
			wantErr:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, cleanup := newTestDB(t, tt.preexisting...)
			defer cleanup()

			svc := service.New(log.Logger, r)

			ts := httptest.NewServer(svc.Routes())
			defer ts.Close()

			rsp, err := ts.Client().Get(fmt.Sprintf("%s/v1/fact/%s", ts.URL, tt.inputID))
			if err != nil {
				t.Fatal(err)
			}
			defer rsp.Body.Close()

			if rsp.StatusCode != tt.wantStatus {
				t.Fatalf("want http %d, got http %d",
					tt.wantStatus, rsp.StatusCode)
			}

			var response struct {
				Fact  service.Fact `json:"fact"`
				Error string       `json:"error"`
			}

			if err := json.NewDecoder(rsp.Body).Decode(&response); err != nil {
				t.Fatal(err)
			}

			if response.Fact.Content != tt.wantFact.Content {
				t.Fatalf("want content %q, got content %q",
					tt.wantFact.Content, response.Fact.Content)
			}

			if response.Fact.Source != tt.wantFact.Source {
				t.Fatalf("want source %q, got source %q",
					tt.wantFact.Source, response.Fact.Content)
			}
		})
	}
}

func TestDeleteFactInputValidation(t *testing.T) {
	tests := []struct {
		name       string
		inputID    string
		wantStatus int
		wantError  string
	}{
		{
			name:       "non-integer",
			inputID:    "asdf",
			wantStatus: http.StatusBadRequest,
			wantError:  "id must be an integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, cleanup := newTestDB(t)
			defer cleanup()

			svc := service.New(log.Logger, r)

			ts := httptest.NewServer(svc.Routes())
			defer ts.Close()

			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/v1/fact/%s", ts.URL, tt.inputID), strings.NewReader(""))
			if err != nil {
				t.Fatal(err)
			}

			rsp, err := ts.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer rsp.Body.Close()

			if tt.wantStatus != rsp.StatusCode {
				t.Fatalf("want http %d, got http %d", tt.wantStatus, rsp.StatusCode)
			}

			var response struct {
				Error string `json:"error"`
			}

			if err := json.NewDecoder(rsp.Body).Decode(&response); err != nil {
				t.Fatal(err)
			}

			if tt.wantError != response.Error {
				t.Fatalf("want error %q, got error %q", tt.wantError, response.Error)
			}
		})
	}
}

func TestDeleteFact(t *testing.T) {
	r, cleanup := newTestDB(t, service.Fact{Content: "To be removed", Source: "No one"})
	defer cleanup()

	svc := service.New(log.Logger, r)

	ts := httptest.NewServer(svc.Routes())
	defer ts.Close()

	// The ID for that single row in the test DB above _should_ be 1.
	req, err := http.NewRequest(http.MethodDelete, ts.URL+"/v1/fact/1", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}

	rsp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rsp.Body.Close()

	wantStatus := http.StatusNoContent
	if wantStatus != rsp.StatusCode {
		t.Fatalf("want http %d, got http %d", wantStatus, rsp.StatusCode)
	}

	// Repeat the delete request, this time it shouldn't be found.
	rsp, err = ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rsp.Body.Close()

	wantStatus = http.StatusNotFound
	if wantStatus != rsp.StatusCode {
		t.Fatalf("want http %d, got http %d", wantStatus, rsp.StatusCode)
	}
}
