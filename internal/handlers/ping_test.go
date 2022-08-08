package handlers

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockPingService struct {
}

func (m mockPingService) Ping() string {
	return "pong"
}

func TestNewHTTPHandler(t *testing.T) {
	// Given
	s := &mockPingService{}

	// When
	h := NewHTTPHandler(s)

	// Then
	require.NotNil(t, h)
}

func TestHTTPHandler_Ping(t *testing.T) {
	// Given
	s := mockPingService{}
	httpHandler := &HTTPHandler{service: s}
	handler := httpHandler.Ping()
	respRecorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))

	// When
	handler.ServeHTTP(respRecorder, req)

	// Then
	var resp response
	require.NoError(t, json.NewDecoder(respRecorder.Body).Decode(&resp))
	require.Equal(t, http.StatusOK, respRecorder.Code)
	require.Equal(t, StatusOK, resp.Status)
	require.Equal(t, "pong", resp.Result)
}