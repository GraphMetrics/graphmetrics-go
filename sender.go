package graphmetrics

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/graphmetrics/logger-go"
	"github.com/hashicorp/go-retryablehttp"

	"github.com/graphmetrics/graphmetrics-go/internal"
)

type Sender struct {
	client    *retryablehttp.Client
	wg        *sync.WaitGroup
	url       string
	apiKey    string
	userAgent string

	metricsChan chan *internal.UsageMetrics
	stopChan    chan interface{}
	stopTimeout time.Duration

	logger logger.Logger
}

func NewSender(cfg *Configuration) *Sender {
	c := retryablehttp.NewClient()
	c.RetryWaitMax = 1 * time.Minute
	c.RetryMax = 8 // Will retry for ~5 minutes
	c.Logger = internal.NewRetryableLogger(cfg.getLogger())
	return &Sender{
		client:      c,
		wg:          &sync.WaitGroup{},
		url:         fmt.Sprintf("%s://%s/reporting/metrics", cfg.getProtocol(), cfg.getEndpoint()),
		apiKey:      cfg.ApiKey,
		userAgent:   fmt.Sprintf("sdk/go/%s", internal.GetModuleVersion()),
		metricsChan: make(chan *internal.UsageMetrics),
		stopChan:    make(chan interface{}),
		stopTimeout: cfg.getStopTimeout(),
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
		req.Header.Set("user-agent", "sdk/js/${PACKAGE.version}")
		req.Header.Set("x-api-key", s.apiKey)
		_, err = s.client.Do(req)
		if err != nil {
			s.logger.Error("unable to send reporting request", map[string]interface{}{
				"error": err,
			})
		}
	}()
}

func (s *Sender) Stop() error {
	s.logger.Debug("stopping sender", nil)

	// Send remaining metrics
	s.stopChan <- nil
	close(s.metricsChan)
	for sketch := range s.metricsChan {
		s.send(sketch)
	}

	// Wait or timeout
	// Note: this can create goroutine leak, but we don't care since Stop is only call on server exit
	c := make(chan struct{})
	go func() {
		defer close(c)
		s.wg.Wait()
	}()
	select {
	case <-c:
		return nil
	case <-time.After(s.stopTimeout):
		s.logger.Error("sending remaining metrics timed out", nil)
		return errors.New("sending remaining metrics timed out")
	}
}

func marshalGzip(w io.Writer, i interface{}) error {
	gz := gzip.NewWriter(w)
	if err := json.NewEncoder(gz).Encode(i); err != nil {
		return err
	}
	return gz.Close()
}
