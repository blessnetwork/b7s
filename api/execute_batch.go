package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/blessnetwork/b7s/models/request"
	"github.com/blessnetwork/b7s/telemetry/tracing"
	"github.com/labstack/echo/v4"
)

func (a *API) ExecuteFunctionBatch(ctx echo.Context) error {

	var req BatchExecutionRequest
	err := ctx.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("could not unpack request: %w", err))
	}

	exr := request.ExecuteBatch{
		Template: request.ExecutionRequestTemplate{
			FunctionID: req.Template.FunctionId,
			Method:     req.Template.Method,
			Config:     req.Template.Config,
		},
		Topic:     req.Topic,
		Arguments: req.Arguments,
	}

	// Start a background context but include trace info.
	ectx := tracing.TraceContext(context.Background(), tracing.GetTraceInfo(ctx.Request().Context()))

	// Background context because we don't want our request to be cancelled if the HTTP request gets cancelled.
	id, err := a.Node.StartFunctionBatchExecution(ectx, exr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("batch execution failed: %w", err))
	}

	return ctx.JSON(http.StatusOK,
		BatchExecutionResponse{
			RequestId: id,
		},
	)
}
