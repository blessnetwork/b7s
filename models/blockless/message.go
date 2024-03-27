package blockless

// TODO: Remove unused/messages that don't make sense here.

// Message types in the Blockless protocol.
const (
	MessageHealthCheck             = "MsgHealthCheck"
	MessageExecute                 = "MsgExecute"
	MessageExecuteResult           = "MsgExecuteResult"
	MessageExecuteError            = "MsgExecuteError"
	MessageExecuteTimeout          = "MsgExecuteTimeout"
	MessageExecuteUnknown          = "MsgExecuteUnknown"
	MessageExecuteInvalid          = "MsgExecuteInvalid"
	MessageExecuteNotFound         = "MsgExecuteNotFound"
	MessageExecuteNotSupported     = "MsgExecuteNotSupported"
	MessageExecuteNotImplemented   = "MsgExecuteNotImplemented"
	MessageExecuteNotAuthorized    = "MsgExecuteNotAuthorized"
	MessageExecuteNotPermitted     = "MsgExecuteNotPermitted"
	MessageExecuteNotAvailable     = "MsgExecuteNotAvailable"
	MessageExecuteNotReady         = "MsgExecuteNotReady"
	MessageExecuteNotConnected     = "MsgExecuteNotConnected"
	MessageExecuteNotInitialized   = "MsgExecuteNotInitialized"
	MessageExecuteNotConfigured    = "MsgExecuteNotConfigured"
	MessageExecuteNotInstalled     = "MsgExecuteNotInstalled"
	MessageExecuteNotUpgraded      = "MsgExecuteNotUpgraded"
	MessageRollCall                = "MsgRollCall"
	MessageRollCallResponse        = "MsgRollCallResponse"
	MessageExecuteResponse         = "MsgExecuteResponse"
	MessageInstallFunction         = "MsgInstallFunction"
	MessageInstallFunctionResponse = "MsgInstallFunctionResponse"
	MessageFormCluster             = "MsgFormCluster"
	MessageFormClusterResponse     = "MsgFormClusterResponse"
	MessageDisbandCluster          = "MsgDisbandCluster"
)

type Message interface {
	Type() string
}
