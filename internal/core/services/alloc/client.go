package alloc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

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

func (c *Client) Alloc(ctx context.Context, addrs []string, namespace, resource string, tokens int64) (int64, bool, error) {
	for _, addr := range addrs {
		url := fmt.Sprintf("http://%s/api/v1/internal/alloc", addr)
		body := dto.AllocRequestBody{Namespace: namespace, Resource: resource, Tokens: tokens}
		var bodyBuffer bytes.Buffer
		if err := json.NewEncoder(&bodyBuffer).Encode(body); err != nil {
			return 0, false, fmt.Errorf("failed to encode alloc request body: %w", err)
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

		allocResponseBody := dto.ResponseBody[dto.AllocResponseBody]{}
		if err := json.NewDecoder(res.Body).Decode(&allocResponseBody); err != nil {
			c.logger.Warn("failed to decode response body", "err", err)
			continue
		}

		if allocResponseBody.Status != dto.StatusOK {
			c.logger.Warn("invalid status code", "statusCode", allocResponseBody.Status)
		}

		return allocResponseBody.Result.RemainingTokens, allocResponseBody.Result.OK, nil
	}

	return 0, false, errors.New("all attempts failed")
}

func (c *Client) Free(ctx context.Context, addrs []string, namespace, resource string, tokens int64) (int64, bool, error) {
	for _, addr := range addrs {
		url := fmt.Sprintf("http://%s/api/v1/internal/free", addr)
		body := dto.FreeRequestBody{Namespace: namespace, Resource: resource, Tokens: tokens}
		var bodyBuffer bytes.Buffer
		if err := json.NewEncoder(&bodyBuffer).Encode(body); err != nil {
			return 0, false, fmt.Errorf("failed to encode free request body: %w", err)
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

		freeResponseBody := dto.ResponseBody[dto.FreeResponseBody]{}
		if err := json.NewDecoder(res.Body).Decode(&freeResponseBody); err != nil {
			c.logger.Warn("failed to decode response body", "err", err)
			continue
		}

		if freeResponseBody.Status != dto.StatusOK {
			c.logger.Warn("invalid status code", "statusCode", freeResponseBody.Status)
		}

		return freeResponseBody.Result.RemainingTokens, freeResponseBody.Result.OK, nil
	}

	return 0, false, errors.New("all attempts failed")
}
