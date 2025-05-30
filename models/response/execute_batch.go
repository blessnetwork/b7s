package response

import (
	"encoding/json"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/codes"
	"github.com/libp2p/go-libp2p/core/peer"
)

var _ (json.Marshaler) = (*ExecuteBatch)(nil)

// Execute describes the response to the `MessageExecuteBatch` message.
type ExecuteBatch struct {
	bls.BaseMessage
	RequestID string                       `json:"request_id,omitempty"`
	Code      codes.Code                   `json:"code,omitempty"`
	Strands   map[string]NodeStrandResults `json:"strands,omitempty"`

	// Used to communicate the reason for failure to the user.
	ErrorMessage string `json:"message,omitempty"`
}

type NodeStrandResults struct {
	Peer    peer.ID      `json:"peer,omitempty"`
	Results BatchResults `json:"results,omitempty"`
}

func (e *ExecuteBatch) WithResults(strands map[string]NodeStrandResults) *ExecuteBatch {
	e.Strands = strands
	return e
}

func (e *ExecuteBatch) WithErrorMessage(err error) *ExecuteBatch {
	e.ErrorMessage = err.Error()
	return e
}

func (ExecuteBatch) Type() string { return bls.MessageExecuteResponse }

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
