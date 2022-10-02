package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grafana/dskit/services"
	"github.com/stretchr/testify/require"

	"github.com/Blinkuu/qms/pkg/dto"
)

type mockPingService struct {
	services.NamedService
}

func (m mockPingService) Ping(_ context.Context) string {
	return "pong"
}

func TestNewPingHTTPHandler(t *testing.T) {
	// Given
	s := &mockPingService{}

	// When
	h := NewPingHTTPHandler(s)

	// Then
	require.NotNil(t, h)
}

func TestPingHTTPHandler_Ping(t *testing.T) {
	// Given
	s := mockPingService{}
	httpHandler := &PingHTTPHandler{service: s}
	handler := httpHandler.Ping()
	respRecorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))

	// When
	handler.ServeHTTP(respRecorder, req)

	// Then
	var resp dto.ResponseBody[dto.PingResponseBody]
	require.NoError(t, json.NewDecoder(respRecorder.Body).Decode(&resp))
	require.Equal(t, http.StatusOK, respRecorder.Code)
	require.Equal(t, dto.StatusOK, resp.Status)
	require.Equal(t, dto.PingResponseBody{Msg: "pong"}, resp.Result)
}
