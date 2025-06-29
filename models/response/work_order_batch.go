package response

import (
	"encoding/json"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/execute"
)

type WorkOrderBatch struct {
	bls.BaseMessage

	// NOTE: Worker might not even need to be aware of the overarching request ID.
	// It will help with debugging right now so let's leave it be.
	RequestID string

	// NOTE: We have redundancy here as strand ID is <request-id>:<uuid>.
	// However, this might change too in the future.
	StrandID string

	Results BatchResults
}

type BatchResults map[execute.RequestHash]*BatchFunctionResult

type BatchFunctionResult struct {
	execute.NodeResult

	FunctionInvocation string   `json:"function_invocation,omitempty"`
	Arguments          []string `json:"arguments,omitempty"`
}

func (WorkOrderBatch) Type() string { return bls.MessageWorkOrderBatchResponse }

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
