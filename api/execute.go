package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/node/aggregate"
)

// ExecuteFunction implements the REST API endpoint for function execution.
func (a *API) ExecuteFunction(ctx echo.Context) error {

	// Unpack the API request.
	var req ExecutionRequest
	err := ctx.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("could not unpack request: %w", err))
	}

	exr := execute.Request{
		Config:     req.Config,
		FunctionID: req.FunctionId,
		Method:     req.Method,
		// TODO: Init arguments here, update API models.
	}

	var args []string

	// Maintain backwards compatibility - if someone provided function arguments via the `parameters` field.
	if len(req.Parameters) > 0 {
		for _, p := range req.Parameters {
			args = append(args, p.Value)
		}
	}
	exr.Arguments = args

	err = exr.Valid()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
	}

	// Get the execution result.
	code, id, results, cluster, err := a.Node.ExecuteFunction(ctx.Request().Context(), exr, req.Topic)
	if err != nil {
		a.Log.Warn().Str("function", req.FunctionId).Err(err).Msg("node failed to execute function")
	}

	// Transform the node response format to the one returned by the API.
	res := ExecutionResponse{
		Code:      string(code),
		RequestId: id,
		Results:   aggregate.Aggregate(results),
		Cluster:   cluster,
	}

	// Communicate the reason for failure in these cases.
	if errors.Is(err, bls.ErrRollCallTimeout) || errors.Is(err, bls.ErrExecutionNotEnoughNodes) {
		res.Message = err.Error()
	}

	// Send the response.
	return ctx.JSON(http.StatusOK, res)
}
