package request

import (
	"encoding/json"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/execute"
)

type ExecuteBatch struct {
	bls.BaseMessage

	Topic                  string          `json:"topic,omitempty"`
	Template               execute.Request `json:"template,omitempty"`
	Arguments              [][]string      `json:"arguments,omitempty"`
	WorkerConcurrencyLimit uint            `json:"concurrency_limit,omitempty"`
}

func (e ExecuteBatch) RollCall(id string) *RollCall {

	return &RollCall{
		BaseMessage: bls.BaseMessage{TraceInfo: e.TraceInfo},
		RequestID:   id,
		FunctionID:  e.Template.FunctionID,
		Attributes:  e.Template.Config.Attributes,
	}
}

func (e ExecuteBatch) WorkOrder(id string) any {

	// TBD: Implement.
	return nil
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
