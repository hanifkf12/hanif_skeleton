package usecase

import (
	"testing"

	"github.com/hanifkf12/hanif_skeleton/internal/appctx"
	"github.com/hanifkf12/hanif_skeleton/internal/appctx/appctxtest"
	"github.com/stretchr/testify/require"
)

func TestHealthDoesNotDependOnApplicationTables(t *testing.T) {
	service := NewHealth()

	resp := service.Serve(appctx.Data{Request: appctxtest.New()})

	require.Equal(t, appctx.StatusOK, resp.Code)
	require.True(t, resp.Status)
}
