package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/blessnetwork/b7s/models/request"
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

	// Background context because we don't want our request to be cancelled if the HTTP request gets cancelled.
	res, err := a.Node.ExecuteFunctionBatch(context.Background(), exr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("batch execution failed: %w", err))
	}

	out := BatchExecutionResponse{
		RequestId: res.RequestID,
		Code:      res.Code.String(),
		Message:   res.ErrorMessage,
		Chunks:    res.Chunks,
	}

	return ctx.JSON(http.StatusOK, out)
}
