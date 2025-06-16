package request

import (
	"encoding/json"

	"github.com/blessnetwork/b7s/models/bls"
)

type WorkOrderBatch struct {
	bls.BaseMessage

	Template  ExecutionRequestTemplate `json:"template,omitempty"`
	StrandID  string                   `json:"strand_id,omitempty"`
	Arguments [][]string               `json:"arguments,omitempty"`
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
