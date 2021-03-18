package jsend

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/random-guys/go-siber/json"
	"github.com/random-guys/go-siber/jwt"
	"github.com/random-guys/go-siber/requests"
	"github.com/rs/zerolog"
)

func TestRecoverer(t *testing.T) {
	router := chi.NewRouter()
	logOut := &bytes.Buffer{}

	router.Use(requests.AttachLogger(zerolog.New(logOut).With().Logger()))
	router.With(Recoverer("production")).Get("/", func(w http.ResponseWriter, r *http.Request) {
		panic(Err{
			Code:    http.StatusBadRequest,
			Message: "Request is bad",
		})
	})

	router.With(
		requests.Timeout(time.Second),
		Recoverer("production"),
	).Get("/timeout", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			panic(context.DeadlineExceeded)
		case <-time.After(time.Second * 2):
			_, _ = w.Write([]byte(""))
		}
	})

	router.With(Recoverer("production")).Get("/panic", func(_ http.ResponseWriter, _ *http.Request) {
		panic(errors.New("failure"))
	})

	t.Run("recovers from jsend.Err", func(t *testing.T) {
		defer logOut.Reset()

		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		router.ServeHTTP(res, req)

		resp := Err{}
		err := json.Unmarshal(res.Body.Bytes(), &resp)
		if err != nil {
			t.Fatal(err)
		}

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected the status code to be %d, got %d", http.StatusBadRequest, resp.Code)
		}
	})

	t.Run("recovers from context timeout", func(t *testing.T) {
		defer logOut.Reset()

		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/timeout", nil)
		router.ServeHTTP(res, req)

		if res.Code != http.StatusGatewayTimeout {
			t.Errorf("Expected the status code to be %d, got %d", http.StatusGatewayTimeout, res.Code)
		}

		logs := responseLog{}
		_ = json.Unmarshal(logOut.Bytes(), &logs)

		if logs.Error != context.DeadlineExceeded.Error() {
			t.Errorf("Expected logger to log error as %s, got %s", context.DeadlineExceeded, logs.Error)
		}
	})

	t.Run("recovers from a random error", func(t *testing.T) {
		defer logOut.Reset()

		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/panic", nil)
		router.ServeHTTP(res, req)

		if res.Code != http.StatusInternalServerError {
			t.Errorf("Expected the status code to be %d, got %d", http.StatusInternalServerError, res.Code)
		}

		logs := responseLog{}
		_ = json.Unmarshal(logOut.Bytes(), &logs)

		if logs.Error != "failure" {
			t.Errorf(`Expected logger to log "failure", got %s`, logs.Error)
		}
	})
}

func TestHeadless(t *testing.T) {
	router := chi.NewRouter()
	router.Use(Recoverer("production"))
	router.With(Headless(store)).Get("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("."))
	})

	t.Run("panics with 401", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		router.ServeHTTP(res, req)

		if res.Code != http.StatusUnauthorized {
			t.Errorf("Expected the status code to be %d, got %d", http.StatusUnauthorized, res.Code)
		}
	})

	t.Run("panics with 401", func(t *testing.T) {
		type session struct {
			Name string
		}

		token, err := jwt.EncodeStruct([]byte(secret), time.Minute, session{"Premium"})
		if err != nil {
			t.Fatal(err)
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", scheme+" "+token)
		router.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Errorf("Expected the status code to be %d, got %d", http.StatusOK, res.Code)
		}
	})
}

func TestSecureHeadless(t *testing.T) {
	router := chi.NewRouter()
	router.Use(Recoverer("production"))
	router.With(SecureHeadless(store)).Get("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("."))
	})

	t.Run("panics with 401", func(t *testing.T) {
		type session struct {
			Name string
		}

		token, err := jwt.EncodeEmbedded([]byte(secret), time.Minute, session{"Premium"})
		if err != nil {
			t.Fatal(err)
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", scheme+" "+token)
		router.ServeHTTP(res, req)

		if res.Code != http.StatusUnauthorized {
			t.Errorf("Expected the status code to be %d, got %d", http.StatusOK, res.Code)
		}
	})
}
