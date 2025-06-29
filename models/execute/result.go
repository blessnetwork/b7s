package execute

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/models/codes"
)

// NodeResult is an annotated execution result.
type NodeResult struct {
	Result

	Signature string         `json:"signature,omitempty"` // Signed digest of the response.
	PBFT      PBFTResultInfo `json:"pbft,omitempty"`
	Metadata  any            `json:"metadata,omitempty"`
}

// Result describes an execution result.
type Result struct {
	Code   codes.Code    `json:"code"`
	Result RuntimeOutput `json:"result"`
	Usage  Usage         `json:"usage,omitempty"`
}

// Cluster represents the set of peers that executed the request.
type Cluster struct {
	Main  peer.ID   `json:"main,omitempty"`
	Peers []peer.ID `json:"peers,omitempty"`
}

// RuntimeOutput describes the output produced by the Bless Runtime during execution.
type RuntimeOutput struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Log      string `json:"-"`
}

// Usage represents the resource usage information for a particular execution.
type Usage struct {
	WallClockTime time.Duration `json:"wall_clock_time,omitempty"`
	CPUUserTime   time.Duration `json:"cpu_user_time,omitempty"`
	CPUSysTime    time.Duration `json:"cpu_sys_time,omitempty"`
	MemoryMaxKB   int64         `json:"memory_max_kb,omitempty"`
}

type PBFTResultInfo struct {
	View             *uint     `json:"view,omitempty"`
	RequestTimestamp time.Time `json:"request_timestamp,omitzero"`
	Replica          peer.ID   `json:"replica,omitempty"`
}

// ResultMap contains execution results from multiple peers.
type ResultMap map[peer.ID]NodeResult

// MarshalJSON provides means to correctly handle JSON serialization/deserialization.
// See:
//
//	https://github.com/libp2p/go-libp2p/pull/2156
//	https://github.com/libp2p/go-libp2p-resource-manager/pull/67#issuecomment-1176820561
func (m ResultMap) MarshalJSON() ([]byte, error) {

	em := make(map[string]NodeResult, len(m))
	for p, v := range m {
		em[p.String()] = v
	}

	return json.Marshal(em)
}

func (r *NodeResult) Sign(key crypto.PrivKey) error {

	cp := *r
	// Exclude some of the fields from the signature.
	cp.Signature = ""

	payload, err := json.Marshal(cp)
	if err != nil {
		return fmt.Errorf("could not get byte representation of the record: %w", err)
	}

	sig, err := key.Sign(payload)
	if err != nil {
		return fmt.Errorf("could not sign digest: %w", err)
	}

	r.Signature = hex.EncodeToString(sig)
	return nil
}

func (r NodeResult) VerifySignature(key crypto.PubKey) error {

	cp := r
	// Exclude some of the fields from the signature.
	cp.Signature = ""

	payload, err := json.Marshal(cp)
	if err != nil {
		return fmt.Errorf("could not get byte representation of the record: %w", err)
	}

	sig, err := hex.DecodeString(r.Signature)
	if err != nil {
		return fmt.Errorf("could not decode signature from hex: %w", err)
	}

	ok, err := key.Verify(payload, sig)
	if err != nil {
		return fmt.Errorf("could not verify signature: %w", err)
	}

	if !ok {
		return errors.New("invalid signature")
	}

	return nil
}
