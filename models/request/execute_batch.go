package request

import (
	"encoding/json"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/codes"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/response"
)

type ExecuteBatch struct {
	bls.BaseMessage

	Topic                  string                   `json:"topic,omitempty"`
	Template               ExecutionRequestTemplate `json:"template,omitempty"`
	Arguments              [][]string               `json:"arguments,omitempty"`
	WorkerConcurrencyLimit uint                     `json:"worker_concurrency_limit,omitempty"`
}

func (e ExecuteBatch) Response(c codes.Code, id string) *response.ExecuteBatch {
	return &response.ExecuteBatch{
		BaseMessage: bls.BaseMessage{TraceInfo: e.TraceInfo},
		RequestID:   id,
		Code:        c,
	}
}

type ExecutionRequestTemplate struct {
	FunctionID string         `json:"function_id,omitempty"`
	Method     string         `json:"method,omitempty"`
	Config     execute.Config `json:"config,omitempty"`
}

func (e ExecuteBatch) RollCall(id string) *RollCall {

	return &RollCall{
		BaseMessage: bls.BaseMessage{TraceInfo: e.TraceInfo},
		RequestID:   id,
		FunctionID:  e.Template.FunctionID,
		Attributes:  e.Template.Config.Attributes,
	}
}

func (e ExecuteBatch) WorkOrderBatch(requestID string, chunkID string, arguments ...[]string) *WorkOrderBatch {

	// TBD: Implement.
	w := &WorkOrderBatch{
		BaseMessage:      bls.BaseMessage{TraceInfo: e.TraceInfo},
		RequestID:        requestID,
		ChunkID:          chunkID,
		Template:         e.Template,
		Arguments:        arguments,
		ConcurrencyLimit: e.WorkerConcurrencyLimit,
	}
	return w
}

func (ExecuteBatch) Type() string { return bls.MessageExecuteBatch }

func (e ExecuteBatch) MarshalJSON() ([]byte, error) {
	type Alias ExecuteBatch
	rec := struct {
		Alias
		Type string `json:"type"`
	}{
		Alias: Alias(e),
		Type:  e.Type(),
	}
	return json.Marshal(rec)
}
