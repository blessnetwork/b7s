package response

import (
	"encoding/json"

	"github.com/blessnetwork/b7s/models/bls"
)

type WorkOrderBatch struct {
	bls.BaseMessage

	// TODO: TBD
}

func (WorkOrderBatch) Type() string { return bls.MessageWorkOrderResponse }

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
