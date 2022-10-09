package memberlist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"

	"github.com/Blinkuu/qms/internal/core/domain"
	"github.com/Blinkuu/qms/pkg/dto"
	"github.com/Blinkuu/qms/pkg/log"
)

const (
	ClientName          = "memberlist-client"
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

func (c *Client) Members(ctx context.Context, addrs []string) ([]domain.Instance, error) {
	if len(addrs) < 1 {
		return nil, errors.New("empty address list")
	}

	for _, addr := range addrs {
		url := fmt.Sprintf("http://%s/memberlist", addr)
		r, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create new request with context: %w", err)
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

		memberlistResponseBody := dto.ResponseBody[dto.MemberlistResponseBody]{}
		if err := json.NewDecoder(res.Body).Decode(&memberlistResponseBody); err != nil {
			c.logger.Warn("failed to decode response body", "err", err)
			continue
		}

		if memberlistResponseBody.Status != dto.StatusOK {
			c.logger.Warn("invalid status code", "statusCode", memberlistResponseBody.Status)
		}

		return memberlistResponseBody.Result.Members, nil
	}

	return nil, errors.New("all attempts failed")
}
