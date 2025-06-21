package request

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/hashicorp/go-multierror"
)

type WorkOrderBatch struct {
	bls.BaseMessage

	Template ExecutionRequestTemplate `json:"template,omitempty"`
	// Technically workers don't need to know the request ID.
	// But for easier troubleshooting, at least for now, it's okay.
	RequestID        string     `json:"request_id,omitempty"`
	StrandID         string     `json:"strand_id,omitempty"`
	Arguments        [][]string `json:"arguments,omitempty"`
	ConcurrencyLimit uint       `json:"concurrency_limit,omitempty"`
}

func (WorkOrderBatch) Type() string { return bls.MessageWorkOrderBatch }

func (w WorkOrderBatch) MarshalJSON() ([]byte, error) {
	type Alias WorkOrderBatch
	rec := struct {
		Alias
		Type string `json:"type"`
	}{
		Alias: Alias(w),
		Type:  w.Type(),
	}
	return json.Marshal(rec)
}

func (w WorkOrderBatch) Valid() error {

	var multierr *multierror.Error

	err := execute.Request{
		FunctionID: w.Template.FunctionID,
		Method:     w.Template.Method,
		Config:     w.Template.Config,
	}.Valid()
	if err != nil {
		multierr = multierror.Append(multierr, fmt.Errorf("execution requst is invalid: %w", err))
	}

	if w.RequestID == "" {
		multierr = multierror.Append(multierr, errors.New("request ID is required"))
	}

	if w.StrandID == "" {
		multierr = multierror.Append(multierr, errors.New("strand ID is required"))
	}

	if len(w.Arguments) == 0 {
		multierr = multierror.Append(multierr, errors.New("arguments are required"))
	}

	return multierr.ErrorOrNil()
}
