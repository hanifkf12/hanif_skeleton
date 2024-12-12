package contract

import "github.com/hanifkf12/hanif_skeleton/internal/appctx"

type UseCase interface {
	Serve(data appctx.Data) appctx.Response
}
