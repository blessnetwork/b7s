package mocks

import (
	"context"
	"testing"

	"github.com/blessnetwork/b7s/models/codes"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/request"
	"github.com/blessnetwork/b7s/models/response"
)

// APINode implements the `Node` interface expected by the API.
type APINode struct {
	ExecuteFunctionFunc        func(context.Context, execute.Request, string) (codes.Code, string, execute.ResultMap, execute.Cluster, error)
	ExecuteFunctionBatchFunc   func(context.Context, request.ExecuteBatch) (*response.ExecuteBatch, error)
	ExecutionResultFunc        func(id string) (execute.ResultMap, bool)
	PublishFunctionInstallFunc func(ctx context.Context, uri string, cid string, subgroup string) error
}

func BaselineNode(t *testing.T) *APINode {
	t.Helper()

	node := APINode{
		ExecuteFunctionFunc: func(context.Context, execute.Request, string) (codes.Code, string, execute.ResultMap, execute.Cluster, error) {

			// TODO: Add a generic cluster info
			return GenericExecutionResult.Code, GenericUUID.String(), GenericExecutionResultMap, execute.Cluster{}, nil
		},
		ExecuteFunctionBatchFunc: func(context.Context, request.ExecuteBatch) (*response.ExecuteBatch, error) {
			// TODO: Return success by default.
			return nil, GenericError
		},
		ExecutionResultFunc: func(id string) (execute.ResultMap, bool) {
			return GenericExecutionResultMap, true
		},
		PublishFunctionInstallFunc: func(ctx context.Context, uri string, cid string, subgroup string) error {
			return nil
		},
	}

	return &node
}

func (n *APINode) ExecuteFunction(ctx context.Context, req execute.Request, subgroup string) (codes.Code, string, execute.ResultMap, execute.Cluster, error) {
	return n.ExecuteFunctionFunc(ctx, req, subgroup)
}

func (n *APINode) ExecuteFunctionBatch(ctx context.Context, req request.ExecuteBatch) (*response.ExecuteBatch, error) {
	return n.ExecuteFunctionBatchFunc(ctx, req)
}

func (n *APINode) ExecutionResult(id string) (execute.ResultMap, bool) {
	return n.ExecutionResultFunc(id)
}

func (n *APINode) PublishFunctionInstall(ctx context.Context, uri string, cid string, subgroup string) error {
	return n.PublishFunctionInstallFunc(ctx, uri, cid, subgroup)
}
