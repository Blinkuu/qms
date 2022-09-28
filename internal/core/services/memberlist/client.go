package memberlist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/Blinkuu/qms/internal/core/domain"
	"github.com/Blinkuu/qms/pkg/dto"
	"github.com/Blinkuu/qms/pkg/log"
)

type Client struct {
	logger log.Logger
	client *http.Client
}

func NewClient(logger log.Logger) *Client {
	return &Client{
		logger: logger,
		client: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

func (c *Client) Members(ctx context.Context, addrs []string) ([]domain.Instance, error) {
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
