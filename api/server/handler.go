package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(10, 30)

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// StatusError represents an error with an associated HTTP status code.
type StatusError struct {
	Code int
	Err  error
}

// Allows StatusError to satisfy the error interface.
func (se StatusError) Error() string {
	return se.Err.Error()
}

// Returns our HTTP status code.
func (se StatusError) Status() int {
	return se.Code
}

func ConfigureHandlers(r *http.ServeMux) {
	r.Handle("/", loggingHandler(limit(errorHandler(errorHandle))))
	r.Handle("/cats", loggingHandler(limit(errorHandler(catsHandle))))
	r.Handle("/cat", loggingHandler(limit(errorHandler(catHandle))))
	r.Handle("/ping", loggingHandler(limit(errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		b, err := json.Marshal("Cats Service. Version 0.1")
		if err != nil {
			return err
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		return nil
	}))))
}

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Before executing the handler.
		start := time.Now()
		log.Printf("Strated %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		// After executing the handler.
		log.Printf("Completed %s in %v", r.URL.Path, time.Since(start))
	})
}

type errorHandler func(http.ResponseWriter, *http.Request) error

func (f errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := f(w, r)
	if err != nil {
		switch e := err.(type) {
		case Error:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			log.Printf("HTTP %d - %s", e.Status(), e)
			http.Error(w, e.Error(), e.Status())
		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			log.Printf("HTTP - %s", e)
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
		}
	}
}

func errorHandle(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "Resource Not Found"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	w.Write(jsonResp)
	return nil
}

func catsHandle(w http.ResponseWriter, r *http.Request) error {
	filters := &Filters{}
	values := r.URL.Query()
	if values.Has("attribute") {
		if val := values.Get("attribute"); attributes[val] {
			filters.Attribute = val
		} else {
			return &StatusError{http.StatusBadRequest, fmt.Errorf("error: attribute value cannot be %s", val)}
		}
	}
	if values.Has("order") {
		if val := values.Get("order"); orders[val] {
			filters.Order = val
		} else {
			return &StatusError{http.StatusBadRequest, fmt.Errorf("error: order value cannot be %s", val)}
		}
	}
	if val := values.Get("offset"); val != "" {
		o, err := strconv.Atoi(val)
		if err != nil {
			return &StatusError{http.StatusBadRequest, fmt.Errorf("error: offset value cannot be %s", val)}
		}
		filters.Offset = o
	}
	if val := values.Get("limit"); val != "" {
		l, err := strconv.Atoi(val)
		if err != nil {
			return &StatusError{http.StatusBadRequest, fmt.Errorf("error: limit value cannot be %s", val)}
		}
		filters.Limit = l
	}
	resp, err := s.GetAllCats(filters)
	if err != nil {
		return err
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
	return nil
}

func catHandle(w http.ResponseWriter, r *http.Request) error {
	cat := new(Cat)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(cat)
	if err != nil {
		return &StatusError{http.StatusBadRequest, err}
	}
	if cat.TailLength <= 0 || cat.WhiskersLength <= 0 {
		return &StatusError{http.StatusBadRequest, fmt.Errorf("error: tail or whiskers length cannot be <= 0")}
	}
	err = s.Set(cat)
	if err != nil {
		return err
	}
	resp := http.StatusText(http.StatusCreated)
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
	return nil
}
