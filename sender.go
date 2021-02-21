package graphmetrics

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/graphmetrics/graphmetrics-go/internal/logging"
	"github.com/graphmetrics/graphmetrics-go/internal/models"
	"github.com/graphmetrics/graphmetrics-go/internal/version"

	"github.com/graphmetrics/logger-go"
	"github.com/hashicorp/go-retryablehttp"
)

type Sender struct {
	client    *retryablehttp.Client
	wg        *sync.WaitGroup
	apiKey    string
	userAgent string

	metricsUrl     string
	definitionsUrl string
	stopTimeout    time.Duration

	logger logger.Logger
}

func NewSender(cfg *Configuration) *Sender {
	c := retryablehttp.NewClient()
	c.RetryWaitMax = 1 * time.Minute
	c.RetryMax = 8 // Will retry for ~5 minutes
	c.Logger = logging.NewRetryableLogger(cfg.getLogger())

	baseUrl := fmt.Sprintf("%s://%s/reporting", cfg.getProtocol(), cfg.getEndpoint())
	return &Sender{
		client:    c,
		wg:        &sync.WaitGroup{},
		apiKey:    cfg.ApiKey,
		userAgent: fmt.Sprintf("sdk/go/%s", version.GetModuleVersion()),

		metricsUrl:     fmt.Sprintf("%s/metrics", baseUrl),
		definitionsUrl: fmt.Sprintf("%s/definitions", baseUrl),
		stopTimeout:    cfg.getStopTimeout(),

		logger: cfg.getLogger(),
	}
}

func (s *Sender) SendMetrics(metrics *models.UsageMetrics) {
	s.send(metrics, s.metricsUrl)
}

func (s *Sender) SendDefinitions(definitions *models.UsageDefinitions) {
	s.send(definitions, s.definitionsUrl)
}

func (s *Sender) send(data interface{}, url string) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		// Prepare payload
		r, w := io.Pipe()
		go func() {
			err := marshalGzip(w, data)
			_ = w.CloseWithError(err)
		}()

		// Send request
		req, err := retryablehttp.NewRequest("POST", url, r)
		if err != nil {
			s.logger.Error("unable to create reporting request", map[string]interface{}{
				"error": err,
				"url":   url,
			})
			return
		}
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("user-agent", s.userAgent)
		req.Header.Set("x-api-key", s.apiKey)
		_, err = s.client.Do(req)
		if err != nil {
			s.logger.Error("unable to send reporting request", map[string]interface{}{
				"error": err,
				"url":   url,
			})
		}
	}()
}

func (s *Sender) Stop() error {
	s.logger.Debug("stopping sender", nil)

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
		s.logger.Error("sending remaining requests timed out", nil)
		return errors.New("sending remaining requests timed out")
	}
}

func marshalGzip(w io.Writer, i interface{}) error {
	gz := gzip.NewWriter(w)
	if err := json.NewEncoder(gz).Encode(i); err != nil {
		return err
	}
	return gz.Close()
}
