package graphmetrics

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/graphmetrics/graphmetrics-go/internal"
)

type Sender struct {
	client *retryablehttp.Client
	wg     *sync.WaitGroup
	url    string
	apiKey string

	metricsChan chan *internal.UsageMetrics
	stopChan    chan interface{}

	logger Logger
}

func NewSender(cfg *Configuration) *Sender {
	return &Sender{
		client:      retryablehttp.NewClient(),
		wg:          &sync.WaitGroup{},
		url:         fmt.Sprintf("https://%s/reporting/metrics", cfg.getEndpoint()),
		apiKey:      cfg.ApiKey,
		metricsChan: make(chan *internal.UsageMetrics),
		stopChan:    make(chan interface{}),
		logger:      cfg.getLogger(),
	}
}

func (s *Sender) Start() {
	for {
		select {
		case <-s.stopChan:
			return
		case metrics := <-s.metricsChan:
			s.send(metrics)
		}
	}
}

func (s *Sender) Send(metrics *internal.UsageMetrics) {
	s.metricsChan <- metrics
}

func (s *Sender) send(metrics *internal.UsageMetrics) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		// Prepare payload
		r, w := io.Pipe()
		go func() {
			err := marshalGzip(w, metrics)
			_ = w.CloseWithError(err)
		}()

		// Send request
		req, err := retryablehttp.NewRequest("POST", s.url, r)
		if err != nil {
			s.logger.Error("unable to create reporting request", map[string]interface{}{
				"error": err,
			})
			return
		}
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("x-api-key", s.apiKey)
		_, err = s.client.Do(req)
		if err != nil {
			s.logger.Error("unable to send reporting request", map[string]interface{}{
				"error": err,
			})
		}
	}()
}

func (s *Sender) Stop() {
	s.stopChan <- nil
	for sketch := range s.metricsChan {
		s.send(sketch)
	}
	// TODO: Add a timeout for this wait
	s.wg.Wait()
}

func marshalGzip(w io.Writer, i interface{}) error {
	gz := gzip.NewWriter(w)
	if err := json.NewEncoder(gz).Encode(i); err != nil {
		return err
	}
	return gz.Close()
}
