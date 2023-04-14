package mocks

import (
	"errors"
	"io"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"

	"github.com/blocklessnetworking/b7s/models/blockless"
	"github.com/blocklessnetworking/b7s/models/codes"
	"github.com/blocklessnetworking/b7s/models/execute"
)

// Global variables that can be used for testing. They are valid non-nil values for commonly needed types.
var (
	NoopLogger = zerolog.New(io.Discard)

	GenericError = errors.New("dummy error")

	GenericPeerID = peer.ID([]byte{0x0, 0x24, 0x8, 0x1, 0x12, 0x20, 0x56, 0x77, 0x86, 0x82, 0x76, 0xa, 0xc5, 0x9, 0x63, 0xde, 0xe4, 0x31, 0xfc, 0x44, 0x75, 0xdd, 0x5a, 0x27, 0xee, 0x6b, 0x94, 0x13, 0xed, 0xe2, 0xa3, 0x6d, 0x8a, 0x1d, 0x57, 0xb6, 0xb8, 0x91})

	GenericAddress = "/ip4/127.0.0.1/tcp/9000/p2p/12D3KooWRp3AVk7qtc2Av6xiqgAza1ZouksQaYcS2cvN94kHSCoa"

	GenericString = "test"

	GenericUUID = uuid.UUID{0xd1, 0xc2, 0x44, 0xaf, 0xa3, 0x1d, 0x48, 0x87, 0x93, 0x9d, 0xd6, 0xc7, 0xf, 0xe, 0x4f, 0xd0}

	GenericExecutionResult = execute.Result{
		Code: codes.Unknown,
		Result: execute.RuntimeOutput{
			Stdout:   "generic-execution-result",
			Stderr:   "generic-execution-log",
			ExitCode: 0,
		},
		RequestID: GenericUUID.String(),
	}

	GenericExecutionRequest = execute.Request{
		FunctionID: "generic-function-id",
		Method:     "wasm",
		Parameters: []execute.Parameter{
			{
				Name:  "generic-param-name",
				Value: "generic-param-value",
			},
		},
	}

	GenericManifest = blockless.FunctionManifest{
		ID:          "generic-id",
		Name:        "generic-name",
		Description: "generic-description",
		Function: blockless.Function{
			ID:      "function-id",
			Name:    "function-name",
			Runtime: "generic-runtime",
		},
		Deployment: blockless.Deployment{
			CID:      "generic-cid",
			Checksum: "1234567890",
			URI:      "generic-uri",
		},
		FSRootPath: "/var/tmp/blockless/",
		Entry:      "/var/tmp/blockless/app.wasm",
	}
)
