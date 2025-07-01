package request

import (
	"encoding/json"

	"github.com/blessnetwork/b7s/consensus"
	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/codes"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/response"
)

var _ (json.Marshaler) = (*RollCall)(nil)

// RollCall describes the `MessageRollCall` message payload.
type RollCall struct {
	bls.BaseMessage
	FunctionID string              `json:"function_id,omitempty"`
	RequestID  string              `json:"request_id,omitempty"`
	Consensus  consensus.Type      `json:"consensus"`
	Attributes *execute.Attributes `json:"attributes,omitempty"`
	// This field is used to discriminate against workers that do not support batch executions.
	// Workers that are aware of it should relay this flag back to us, while workers that are not
	// will ignore/omit it.
	Batch bool `json:"batch,omitempty"`
}

func (r RollCall) Response(c codes.Code) *response.RollCall {
	return &response.RollCall{
		BaseMessage:  bls.BaseMessage{TraceInfo: r.TraceInfo},
		FunctionID:   r.FunctionID,
		RequestID:    r.RequestID,
		Code:         c,
		BatchSupport: r.Batch,
	}
}

func (RollCall) Type() string { return bls.MessageRollCall }

func (r RollCall) MarshalJSON() ([]byte, error) {
	type Alias RollCall
	rec := struct {
		Alias
		Type string `json:"type"`
	}{
		Alias: Alias(r),
		Type:  r.Type(),
	}
	return json.Marshal(rec)
}
