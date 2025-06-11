package request

import (
	"encoding/json"
	"time"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/codes"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/response"
)

type WorkOrderBatch struct {
	bls.BaseMessage

	execute.Request // execute request is embedded

	RequestID string    `json:"request_id,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"` // Execution request timestamp is a factor for PBFT.
}

func (w WorkOrderBatch) Response(c codes.Code, res execute.Result) *response.WorkOrderBatch {

	return &response.WorkOrder{
		BaseMessage: bls.BaseMessage{TraceInfo: w.TraceInfo},
		Code:        c,
		RequestID:   w.RequestID,
		Result: execute.NodeResult{
			Result: res,
		},
	}
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
