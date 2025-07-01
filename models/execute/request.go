package execute

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
)

// RequestHash is an md5 checksum of the most important request data:
// CID/method.wasm <arg1> <arg2> ... <argN>
type RequestHash string

func (r Request) GetExecutionID() RequestHash {
	return ExecutionID(r.FunctionID, r.Method, r.Arguments)
}

func (r Request) FunctionInvocation() string {
	return FunctionInvocation(r.FunctionID, r.Method)
}

func ExecutionID(cid string, method string, arguments []string) RequestHash {

	// CID/method.wasm arg1 arg2 arg3
	sum := md5.Sum(
		fmt.Appendf([]byte{}, "%v %v", FunctionInvocation(cid, method), strings.Join(arguments, " ")),
	)
	hex := hex.EncodeToString(sum[:])

	return RequestHash(hex)
}

func FunctionInvocation(cid string, method string) string {
	return cid + "/" + method
}

func (h RequestHash) String() string {
	return string(h)
}

// Request describes an execution request.
type Request struct {
	FunctionID string   `json:"function_id"`
	Method     string   `json:"method"`
	Arguments  []string `json:"arguments,omitempty"`
	Config     Config   `json:"config"`

	// Optional signature of the request.
	Signature string `json:"signature,omitempty"`
}

func (r Request) Valid() error {

	var err *multierror.Error

	if r.FunctionID == "" {
		err = multierror.Append(err, errors.New("function ID is required"))
	}

	if r.Method == "" {
		err = multierror.Append(err, errors.New("method is required"))
	}

	return err.ErrorOrNil()
}

// Parameter represents an execution parameter, modeled as a key-value pair.
type Parameter struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// Config represents the configurable options for an execution request.
type Config struct {
	Runtime           BLSRuntimeConfig  `json:"runtime,omitempty"`
	Environment       []EnvVar          `json:"env_vars,omitempty"`
	Stdin             *string           `json:"stdin,omitempty"`
	Permissions       []string          `json:"permissions,omitempty"`
	ResultAggregation ResultAggregation `json:"result_aggregation,omitempty"`

	Attributes *Attributes `json:"attributes,omitempty"`

	// NodeCount specifies how many nodes should execute this request.
	NodeCount int `json:"number_of_nodes,omitempty"`

	// When should the execution timeout
	Timeout int `json:"timeout,omitempty"`

	// Consensus algorithm to use. Raft and PBFT are supported at this moment.
	ConsensusAlgorithm string `json:"consensus_algorithm,omitempty"`

	// Threshold (percentage) defines how many nodes should respond with a result to consider this execution successful.
	Threshold float64 `json:"threshold,omitempty"`
}

// EnvVar represents the name and value of the environment variables set for the execution.
type EnvVar struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type ResultAggregation struct {
	Enable     bool        `json:"enable,omitempty"`
	Type       string      `json:"type,omitempty"`
	Parameters []Parameter `json:"parameters,omitempty"`
}
