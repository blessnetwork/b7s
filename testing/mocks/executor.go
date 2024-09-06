package mocks

import (
	"context"
	"testing"

	"github.com/blocklessnetwork/b7s/models/blockless"
	"github.com/blocklessnetwork/b7s/models/execute"
)

var _ (blockless.Executor) = (*Executor)(nil)

type Executor struct {
	ExecFunctionFunc func(context.Context, string, execute.Request) (execute.Result, error)
}

func BaselineExecutor(t *testing.T) *Executor {
	t.Helper()

	executor := Executor{
		ExecFunctionFunc: func(context.Context, string, execute.Request) (execute.Result, error) {
			return GenericExecutionResult, nil
		},
	}

	return &executor
}

func (e *Executor) ExecuteFunction(ctx context.Context, requestID string, req execute.Request) (execute.Result, error) {
	return e.ExecFunctionFunc(ctx, requestID, req)
}
