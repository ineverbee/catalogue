package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockServer struct {
}

func (ms *MockServer) GetAllCats(f *Filters) ([]Cat, error) {
	return nil, nil
}

func (ms *MockServer) Set(c *Cat) error {
	return nil
}

func TestHandlers(t *testing.T) {
	s = &MockServer{}

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	errorHandler(errorHandle).ServeHTTP(rr, req)
	require.Equal(t, http.StatusNotFound, rr.Code)

	req = httptest.NewRequest("GET", "/cats", nil)
	rr = httptest.NewRecorder()
	errorHandler(catsHandle).ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	req = httptest.NewRequest("GET", "/cats?attribute=tail_length&order=desc&offset=5&limit=3", nil)
	rr = httptest.NewRecorder()
	errorHandler(catsHandle).ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	req = httptest.NewRequest("GET", "/cats?attribute=bad", nil)
	rr = httptest.NewRecorder()
	errorHandler(catsHandle).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	req = httptest.NewRequest("GET", "/cats?attribute=123", nil)
	rr = httptest.NewRecorder()
	errorHandler(catsHandle).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	req = httptest.NewRequest("GET", "/cats?order=smth", nil)
	rr = httptest.NewRecorder()
	errorHandler(catsHandle).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	req = httptest.NewRequest("GET", "/cats?offset=NaN", nil)
	rr = httptest.NewRecorder()
	errorHandler(catsHandle).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	req = httptest.NewRequest("GET", "/cats?limit=NaN", nil)
	rr = httptest.NewRecorder()
	errorHandler(catsHandle).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	req = httptest.NewRequest("POST", "/cat", strings.NewReader("{\"name\": \"Taylor\", \"color\": \"red & white\", \"tail_length\": 15, \"whiskers_length\": 12}"))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	errorHandler(catHandle).ServeHTTP(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	req = httptest.NewRequest("POST", "/cat", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	errorHandler(catHandle).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	req = httptest.NewRequest("POST", "/cat", strings.NewReader("{\"name\": \"Taylor\", \"color\": \"red & white\", \"tail_length\": -15, \"whiskers_length\": 12}"))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	errorHandler(catHandle).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

}
