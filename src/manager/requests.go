package manager

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	HTTPClient *http.Client
	wg         sync.WaitGroup
	mu         sync.Mutex
}

func NewRequester() *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout: 1 * time.Second,
		},
		wg: sync.WaitGroup{},
		mu: sync.Mutex{},
	}
}

type Requester interface {
	URLRequester(ctx context.Context, urls []string, requests chan map[int]io.ReadCloser, errCh chan error)
}

// URLRequester business logic that works with url requests
func (c *Client) URLRequester(ctx context.Context, urls []string, requests chan map[int]io.ReadCloser, errCh chan error) {
	res := make(map[int]io.ReadCloser, len(urls))

	// create requests to other api's
	for i, url := range urls {
		c.wg.Add(1)
		go func(i int, url string) {
			defer c.wg.Done()
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				errCh <- err
				return
			}
			result, err := c.HTTPClient.Do(req.WithContext(ctx))
			if err != nil {
				errCh <- err
				return
			}
			defer result.Body.Close()
			c.mu.Lock()
			res[i] = result.Body
			c.mu.Unlock()
		}(i, url)
	}
	c.wg.Wait()
	requests <- res
}
