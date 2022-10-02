package rate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"

	"github.com/Blinkuu/qms/pkg/dto"
	"github.com/Blinkuu/qms/pkg/log"
)

const (
	ClientName          = "rate-client"
	defaultRetryWaitMin = 100 * time.Millisecond
	defaultRetryWaitMax = 500 * time.Millisecond
	defaultRetryMax     = 3
)

type Client struct {
	logger log.Logger
	client *http.Client
}

func NewClient(logger log.Logger) *Client {
	client := retryablehttp.Client{
		HTTPClient:   cleanhttp.DefaultPooledClient(),
		Logger:       logger,
		RetryWaitMin: defaultRetryWaitMin,
		RetryWaitMax: defaultRetryWaitMax,
		RetryMax:     defaultRetryMax,
		CheckRetry:   retryablehttp.DefaultRetryPolicy,
		Backoff:      retryablehttp.DefaultBackoff,
	}

	return &Client{
		logger: logger,
		client: client.StandardClient(),
	}
}

func (c *Client) Allow(ctx context.Context, addrs []string, namespace, resource string, tokens int64) (time.Duration, bool, error) {
	for _, addr := range addrs {
		url := fmt.Sprintf("http://%s/api/v1/internal/allow", addr)
		body := dto.AllowRequestBody{Namespace: namespace, Resource: resource, Tokens: tokens}
		var bodyBuffer bytes.Buffer
		if err := json.NewEncoder(&bodyBuffer).Encode(body); err != nil {
			return 0, false, fmt.Errorf("failed to encode allow request body: %w", err)
		}

		r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &bodyBuffer)
		if err != nil {
			return 0, false, fmt.Errorf("failed to create new request with context: %w", err)
		}

		res, err := c.client.Do(r)
		if err != nil {
			c.logger.Warn("failed to do request", "err", err)
			continue
		}
		defer func() {
			if err := res.Body.Close(); err != nil {
				c.logger.Warn("failed to close response body: %w", err)
			}
		}()

		if res.StatusCode != http.StatusOK {
			c.logger.Warn("invalid http status code", "statusCode", res.StatusCode)
			continue
		}

		allowResponseBody := dto.ResponseBody[dto.AllowResponseBody]{}
		if err := json.NewDecoder(res.Body).Decode(&allowResponseBody); err != nil {
			c.logger.Warn("failed to decode response body", "err", err)
			continue
		}

		if allowResponseBody.Status != dto.StatusOK {
			c.logger.Warn("invalid status code", "statusCode", allowResponseBody.Status)
		}

		return time.Duration(allowResponseBody.Result.WaitTime), allowResponseBody.Result.OK, nil
	}

	return 0, false, errors.New("all attempts failed")
}
