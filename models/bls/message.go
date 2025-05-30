package bls

import (
	"github.com/blessnetwork/b7s/telemetry/tracing"
)

type Message interface {
	Type() string
}

// Message types in the bls protocol.
const (
	MessageHealthCheck             = "MsgHealthCheck"
	MessageInstallFunction         = "MsgInstallFunction"
	MessageInstallFunctionResponse = "MsgInstallFunctionResponse"
	MessageRollCall                = "MsgRollCall"
	MessageRollCallResponse        = "MsgRollCallResponse"
	MessageExecute                 = "MsgExecute" // MessageExecute is the execution request, as expected by the head node.
	MessageExecuteResponse         = "MsgExecuteResponse"
	MessageWorkOrder               = "MsgWorkOrder" // MessageWorkOrder is the execution request, as expected by the worker node.
	MessageWorkOrderResponse       = "MsgWorkOrderResponse"
	MessageFormCluster             = "MsgFormCluster"
	MessageFormClusterResponse     = "MsgFormClusterResponse"
	MessageDisbandCluster          = "MsgDisbandCluster"

	MessageExecuteBatch   = "MsgExecuteBatch"
	MessageWorkOrderBatch = "MsgWorkOrderBatch"
)

type TraceableMessage interface {
	Message
	SaveTraceContext(tracing.TraceInfo)
}

type BaseMessage struct {
	tracing.TraceInfo
}

func (m *BaseMessage) SaveTraceContext(t tracing.TraceInfo) {
	m.TraceInfo = t
}
