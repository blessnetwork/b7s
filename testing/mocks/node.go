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
	ExecuteFunctionFunc             func(context.Context, execute.Request, string) (codes.Code, string, execute.ResultMap, execute.Cluster, error)
	StartFunctionBatchExecutionFunc func(context.Context, request.ExecuteBatch) (string, error)
	GetBatchResultsFunc             func(context.Context, string) (*response.ExecuteBatch, error)
	ExecutionResultFunc             func(id string) (execute.ResultMap, bool)
	PublishFunctionInstallFunc      func(ctx context.Context, uri string, cid string, subgroup string) error
}

func BaselineNode(t *testing.T) *APINode {
	t.Helper()

	node := APINode{
		ExecuteFunctionFunc: func(context.Context, execute.Request, string) (codes.Code, string, execute.ResultMap, execute.Cluster, error) {

			var (
				code    = GenericExecutionResult.Code
				uuid    = GenericUUID.String()
				result  = GenericExecutionResultMap
				cluster = execute.Cluster{
					Main:  GenericPeerIDs[0],
					Peers: GenericPeerIDs[:4],
				}
			)

			return code, uuid, result, cluster, nil
		},
		StartFunctionBatchExecutionFunc: func(context.Context, request.ExecuteBatch) (string, error) {
			return "", nil
		},
		GetBatchResultsFunc: func(context.Context, string) (*response.ExecuteBatch, error) {
			return nil, nil
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

func (n *APINode) StartFunctionBatchExecution(ctx context.Context, req request.ExecuteBatch) (string, error) {
	return n.StartFunctionBatchExecutionFunc(ctx, req)
}

func (n *APINode) GetBatchResults(ctx context.Context, id string) (*response.ExecuteBatch, error) {
	return n.GetBatchResultsFunc(ctx, id)
}

func (n *APINode) ExecutionResult(id string) (execute.ResultMap, bool) {
	return n.ExecutionResultFunc(id)
}

func (n *APINode) PublishFunctionInstall(ctx context.Context, uri string, cid string, subgroup string) error {
	return n.PublishFunctionInstallFunc(ctx, uri, cid, subgroup)
}
