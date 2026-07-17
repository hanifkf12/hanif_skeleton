package appctx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewResponseUsesRequestTime(t *testing.T) {
	first := NewResponse()
	time.Sleep(time.Millisecond)
	second := NewResponse()

	require.True(t, second.Timestamp.After(first.Timestamp))
	require.Equal(t, time.UTC, first.Timestamp.Location())
}
