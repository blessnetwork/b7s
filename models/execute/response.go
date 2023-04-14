package execute

import (
	"time"

	"github.com/blocklessnetworking/b7s/models/codes"
)

// Result describes an execution result.
type Result struct {
	Code      codes.Code    `json:"code"`
	Result    RuntimeOutput `json:"result"`
	RequestID string        `json:"request_id"`
	Usage     Usage         `json:"usage,omitempty"`
}

// RuntimeOutput describes the output produced by the Blockless Runtime during exection.
type RuntimeOutput struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Log      string `json:"-"` // TODO: Check do we want to send this over the wire too?
}

// Usage represents the resource usage information for a particular execution.
type Usage struct {
	WallClockTime time.Duration `json:"wall_clock_time,omitempty"`
	CPUUserTime   time.Duration `json:"cpu_user_time,omitempty"`
	CPUSysTime    time.Duration `json:"cpu_sys_time,omitempty"`
	MemoryMaxKB   int64         `json:"memory_max_kb,omitempty"`
}
