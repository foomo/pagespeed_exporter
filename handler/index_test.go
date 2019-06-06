package handler

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewIndexHandler(t *testing.T) {
	handler := NewIndexHandler()

	require.HTTPSuccess(t, handler.ServeHTTP, "GET", "/", nil)
	require.HTTPBodyContains(t, handler.ServeHTTP, "GET", "/", nil, string(indexPage))
}
