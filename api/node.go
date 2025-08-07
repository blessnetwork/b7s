package api

import (
	"context"

	"github.com/blessnetwork/b7s/models/codes"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/request"
	"github.com/blessnetwork/b7s/models/response"
)

// TODO: ExecutionFunctionBatch makes a detour from the established approach
// by directly using the request/response types. Consider if
// other handlers should do the same, bringing down REST API handlers closer
// to their p2p counterpart, which is what REST API is trying to emulate.

type Node interface {
	ExecuteFunction(ctx context.Context, req execute.Request, subgroup string) (code codes.Code, requestID string, results execute.ResultMap, peers execute.Cluster, err error)
	ExecuteFunctionBatch(ctx context.Context, req request.ExecuteBatch) (*response.ExecuteBatch, error)
	ExecutionResult(id string) (execute.ResultMap, bool)
	PublishFunctionInstall(ctx context.Context, uri string, cid string, subgroup string) error
}
