package response

import (
	"encoding/json"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/codes"
	"github.com/blessnetwork/b7s/models/execute"
)

var _ (json.Marshaler) = (*Execute)(nil)

// Execute describes the response to the `MessageExecute` message.
type Execute struct {
	bls.BaseMessage
	RequestID string            `json:"request_id,omitempty"`
	Code      codes.Code        `json:"code,omitzero"`
	Results   execute.ResultMap `json:"results,omitzero"`
	Cluster   execute.Cluster   `json:"cluster,omitzero"`

	// Used to communicate the reason for failure to the user.
	ErrorMessage string `json:"message,omitzero"`
}

func (e *Execute) WithResults(r execute.ResultMap) *Execute {
	e.Results = r
	return e
}

func (e *Execute) WithCluster(c execute.Cluster) *Execute {
	e.Cluster = c
	return e
}

func (e *Execute) WithErrorMessage(err error) *Execute {
	e.ErrorMessage = err.Error()
	return e
}

func (Execute) Type() string { return bls.MessageExecuteResponse }

func (e Execute) MarshalJSON() ([]byte, error) {
	type Alias Execute
	rec := struct {
		Alias
		Type string `json:"type"`
	}{
		Alias: Alias(e),
		Type:  e.Type(),
	}
	return json.Marshal(rec)
}
