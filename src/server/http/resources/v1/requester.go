package v1

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/JetBrainer/ApiService/src/manager"
)

const clientConnections = 100

var ErrTooManyURLs = errors.New("too many url's")

type URLResource struct {
	connections chan struct{}
	urlManager  manager.Requester
}

func NewURLResource(cruder manager.Requester) *URLResource {
	return &URLResource{
		connections: make(chan struct{}, clientConnections),
		urlManager:  cruder,
	}
}

// RequestLimiter limit's number of connections
func (cr *URLResource) RequestLimiter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request %s\n", r.RemoteAddr)
		cr.connections <- struct{}{}
		defer func() {
			<-cr.connections
		}()
		next(w, r)
	}
}

// URLHandler multiplexing request
func (cr *URLResource) URLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		cr.URLList(w, r)
	}
}

// URLList handle url's that coming from client
func (cr *URLResource) URLList(w http.ResponseWriter, r *http.Request) {
	var urls []string
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	if err := json.NewDecoder(r.Body).Decode(&urls); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	errCh := make(chan error)
	responses := make(chan map[int]io.ReadCloser)

	if err := Validation(urls); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	go func() {
		cr.urlManager.URLRequester(ctx, urls, responses, errCh)
	}()
	select {
	case err := <-errCh:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	case <-ctx.Done():
		return
	case res := <-responses:
		log.Println(res[0])
		if err := json.NewEncoder(w).Encode(&res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
}

// Validation of url's count and message
func Validation(urls []string) error {
	if len(urls) > 20 {
		return ErrTooManyURLs
	}

	for _, address := range urls {
		_, err := url.Parse(address)
		if err != nil {
			return err
		}
	}

	return nil
}