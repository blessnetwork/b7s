package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (a *API) BatchExecutionResult(ctx echo.Context) error {

	var req BatchResultRequest
	err := ctx.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("could not unpack request: %w", err))
	}

	res, err := a.Node.GetBatchResults(ctx.Request().Context(), req.Id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("could not retrieve batch execution result: %w", err))
	}

	out := BatchExecutionResult{
		RequestId: res.RequestID,
		Code:      res.Code.String(),
		Message:   res.ErrorMessage,
		Chunks:    res.Chunks,
	}

	return ctx.JSON(http.StatusOK, out)
}
