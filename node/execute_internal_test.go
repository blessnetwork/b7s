package node

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/stretchr/testify/require"

	"github.com/blocklessnetworking/b7s/host"
	"github.com/blocklessnetworking/b7s/models/blockless"
	"github.com/blocklessnetworking/b7s/models/execute"
	"github.com/blocklessnetworking/b7s/models/request"
	"github.com/blocklessnetworking/b7s/models/response"
	"github.com/blocklessnetworking/b7s/testing/mocks"
)

func TestNode_WorkerExecute(t *testing.T) {

	const (
		functionID     = "dummy-function-id"
		functionMethod = "dummy-function-method"
	)

	executionRequest := request.Execute{
		Type:       blockless.MessageExecute,
		FunctionID: functionID,
		Method:     functionMethod,
		Parameters: []execute.Parameter{},
		Config:     execute.Config{},
	}

	payload := serialize(t, executionRequest)

	t.Run("handles correct execution", func(t *testing.T) {
		t.Parallel()

		var (
			requestID string
		)

		node := createNode(t, blockless.WorkerNode)

		// Use a custom executor to verify all execution parameters are correct.
		executor := mocks.BaselineExecutor(t)
		executor.ExecFunctionFunc = func(reqID string, req execute.Request) (execute.Result, error) {
			require.NotEmpty(t, reqID)
			require.Equal(t, executionRequest.FunctionID, req.FunctionID)
			require.Equal(t, executionRequest.Method, req.Method)
			require.ElementsMatch(t, executionRequest.Parameters, req.Parameters)
			require.Equal(t, executionRequest.Config, req.Config)

			requestID = reqID
			res := mocks.GenericExecutionResult
			res.RequestID = requestID

			return res, nil
		}
		node.executor = executor

		// Create a host that will serve as a receiver of the execution response.
		receiver, err := host.New(mocks.NoopLogger, loopback, 0)
		require.NoError(t, err)

		hostAddNewPeer(t, node.host, receiver)

		var wg sync.WaitGroup
		wg.Add(1)

		receiver.SetStreamHandler(blockless.ProtocolID, func(stream network.Stream) {
			defer wg.Done()
			defer stream.Close()

			from := stream.Conn().RemotePeer()
			require.Equal(t, node.host.ID(), from)

			var received response.Execute
			getStreamPayload(t, stream, &received)

			require.Equal(t, blockless.MessageExecuteResponse, received.Type)

			// We should receive the response the baseline executor will return.
			expected := mocks.GenericExecutionResult
			require.Equal(t, requestID, received.RequestID)
			require.Equal(t, expected.Code, received.Code)
			require.Equal(t, expected.Result, received.Result)
		})

		err = node.processExecute(context.Background(), receiver.ID(), payload)
		require.NoError(t, err)

		wg.Wait()
	})
	t.Run("handles execution failure", func(t *testing.T) {
		t.Parallel()

		var (
			faultyExecutionResult = execute.Result{
				Code:      response.CodeError,
				Result:    "something horrible has happened",
				RequestID: mocks.GenericUUID.String(),
			}
		)

		node := createNode(t, blockless.WorkerNode)

		// Use a custom executor to verify all execution parameters are correct.
		executor := mocks.BaselineExecutor(t)
		executor.ExecFunctionFunc = func(requestID string, req execute.Request) (execute.Result, error) {
			return faultyExecutionResult, mocks.GenericError
		}
		node.executor = executor

		// Create a host that will serve as a receiver of the execution response.
		receiver, err := host.New(mocks.NoopLogger, loopback, 0)
		require.NoError(t, err)

		hostAddNewPeer(t, node.host, receiver)

		var wg sync.WaitGroup
		wg.Add(1)

		receiver.SetStreamHandler(blockless.ProtocolID, func(stream network.Stream) {
			defer wg.Done()
			defer stream.Close()

			from := stream.Conn().RemotePeer()
			require.Equal(t, node.host.ID(), from)

			var received response.Execute
			getStreamPayload(t, stream, &received)

			require.Equal(t, blockless.MessageExecuteResponse, received.Type)

			require.Equal(t, faultyExecutionResult.RequestID, received.RequestID)
			require.Equal(t, faultyExecutionResult.Code, received.Code)
			require.Equal(t, faultyExecutionResult.Result, received.Result)
		})

		err = node.processExecute(context.Background(), receiver.ID(), payload)
		require.NoError(t, err)

		wg.Wait()
	})
	t.Run("handles function store errors", func(t *testing.T) {
		t.Parallel()

		node := createNode(t, blockless.WorkerNode)

		// Error retrieving function.
		fstore := mocks.BaselineFunctionHandler(t)
		fstore.GetFunc = func(string, string, bool) (*blockless.FunctionManifest, error) {
			return nil, mocks.GenericError
		}
		node.function = fstore

		// Create a host that will serve as a receiver of the execution response.
		receiver, err := host.New(mocks.NoopLogger, loopback, 0)
		require.NoError(t, err)

		hostAddNewPeer(t, node.host, receiver)

		var wg sync.WaitGroup
		wg.Add(1)

		receiver.SetStreamHandler(blockless.ProtocolID, func(stream network.Stream) {
			defer wg.Done()
			defer stream.Close()

			from := stream.Conn().RemotePeer()
			require.Equal(t, node.host.ID(), from)

			var received response.Execute
			getStreamPayload(t, stream, &received)

			require.Equal(t, blockless.MessageExecuteResponse, received.Type)

			require.Equal(t, received.Code, response.CodeError)
		})

		err = node.processExecute(context.Background(), receiver.ID(), payload)
		require.NoError(t, err)

		wg.Wait()

		// Function is not installed.
		fstore.GetFunc = func(string, string, bool) (*blockless.FunctionManifest, error) {
			return nil, blockless.ErrNotFound
		}
		node.function = fstore

		wg.Add(1)

		receiver.SetStreamHandler(blockless.ProtocolID, func(stream network.Stream) {
			defer wg.Done()
			defer stream.Close()

			from := stream.Conn().RemotePeer()
			require.Equal(t, node.host.ID(), from)

			var received response.Execute
			getStreamPayload(t, stream, &received)

			require.Equal(t, blockless.MessageExecuteResponse, received.Type)

			require.Equal(t, received.Code, response.CodeNotFound)
		})

		err = node.processExecute(context.Background(), receiver.ID(), payload)
		require.NoError(t, err)

		wg.Wait()
	})
	t.Run("handles malformed request", func(t *testing.T) {
		t.Parallel()

		const (
			// JSON without closing brace.
			malformedJSON = `{
				"type": "MsgExecute",
				"function_id": "dummy-function-id",
				"method": "dummy-function-method",
				"config": {}`
		)

		node := createNode(t, blockless.WorkerNode)

		err := node.processExecute(context.Background(), mocks.GenericPeerID, []byte(malformedJSON))
		require.Error(t, err)
	})
}

func TestNode_HeadExecute(t *testing.T) {

	const (
		functionID     = "dummy-function-id"
		functionMethod = "dummy-function-method"
	)

	executionRequest := request.Execute{
		Type:       blockless.MessageExecute,
		FunctionID: functionID,
		Method:     functionMethod,
		Parameters: []execute.Parameter{},
		Config:     execute.Config{},
	}

	payload := serialize(t, executionRequest)

	t.Run("handles roll call timeout", func(t *testing.T) {
		t.Parallel()

		node := createNode(t, blockless.HeadNode)

		ctx := context.Background()
		_, err := node.subscribe(ctx)
		require.NoError(t, err)

		// Create a host that will receive the execution response.
		receiver, err := host.New(mocks.NoopLogger, loopback, 0)
		require.NoError(t, err)

		hostAddNewPeer(t, node.host, receiver)

		var wg sync.WaitGroup

		wg.Add(1)
		receiver.SetStreamHandler(blockless.ProtocolID, func(stream network.Stream) {
			defer wg.Done()
			defer stream.Close()

			from := stream.Conn().RemotePeer()
			require.Equal(t, node.host.ID(), from)

			var received response.Execute
			getStreamPayload(t, stream, &received)

			require.Equal(t, blockless.MessageExecuteResponse, received.Type)
			require.Equal(t, response.CodeTimeout, received.Code)
		})

		// Since no one will respond to a roll call, this is bound to time out.
		err = node.processExecute(ctx, receiver.ID(), payload)
		require.NoError(t, err)

		wg.Wait()
	})
	t.Run("handles correct execution", func(t *testing.T) {
		t.Parallel()

		const (
			topic           = DefaultTopic
			executionResult = "dummy-execution-result"
		)

		var (
			requestID string
		)

		ctx, cancel := context.WithCancel(context.Background())

		node := createNode(t, blockless.HeadNode)
		node.listenDirectMessages(ctx)

		defer cancel()
		_, err := node.subscribe(ctx)
		require.NoError(t, err)

		// Create a host that will simulate a worker.
		// It will listen to a roll call request and reply,
		// as well as feign execution.
		mockWorker, err := host.New(mocks.NoopLogger, loopback, 0)
		require.NoError(t, err)

		_, subscription, err := mockWorker.Subscribe(ctx, topic)
		require.NoError(t, err)

		hostAddNewPeer(t, node.host, mockWorker)

		// Connect to the node so they exchange topic subscription info.
		info := hostGetAddrInfo(t, node.host)
		err = mockWorker.Connect(ctx, *info)

		// Mock worker will feign execution.
		mockWorker.SetStreamHandler(blockless.ProtocolID, func(stream network.Stream) {
			defer stream.Close()

			var req request.Execute
			getStreamPayload(t, stream, &req)

			from := stream.Conn().RemotePeer()
			require.Equal(t, node.host.ID(), from)

			require.Equal(t, blockless.MessageExecute, req.Type)

			res := response.Execute{
				Type:      blockless.MessageExecuteResponse,
				RequestID: requestID,
				Result:    executionResult,
				Code:      response.CodeOK,
			}
			payload := serialize(t, res)
			err = mockWorker.SendMessage(ctx, node.host.ID(), payload)
			require.NoError(t, err)
		})

		// Create a host that will receive the execution response.
		receiver, err := host.New(mocks.NoopLogger, loopback, 0)
		require.NoError(t, err)

		hostAddNewPeer(t, node.host, receiver)

		var receiverWG sync.WaitGroup

		receiverWG.Add(1)
		receiver.SetStreamHandler(blockless.ProtocolID, func(stream network.Stream) {
			defer receiverWG.Done()
			defer stream.Close()

			from := stream.Conn().RemotePeer()
			require.Equal(t, node.host.ID(), from)

			var res response.Execute
			getStreamPayload(t, stream, &res)
			require.Equal(t, blockless.MessageExecuteResponse, res.Type)

			require.Equal(t, response.CodeOK, res.Code)
			require.Equal(t, requestID, res.RequestID)
			require.Equal(t, executionResult, res.Result)
		})

		var nodeWG sync.WaitGroup
		nodeWG.Add(1)

		// Start the node request asynchronously.
		go func() {
			defer nodeWG.Done()

			time.Sleep(subscriptionDiseminationPause)

			err = node.processExecute(ctx, receiver.ID(), payload)
			require.NoError(t, err)
		}()

		// Mock worker workflow.

		deadlineCtx, dcancel := context.WithTimeout(ctx, publishTimeout)
		defer dcancel()

		// Mock worker should wait for the roll call to be broadcast.
		msg, err := subscription.Next(deadlineCtx)
		require.NoError(t, err)

		from := msg.ReceivedFrom
		require.Equal(t, node.host.ID(), from)

		var received request.RollCall
		err = json.Unmarshal(msg.Data, &received)

		require.Equal(t, blockless.MessageRollCall, received.Type)
		require.Equal(t, functionID, received.FunctionID)

		requestID = received.RequestID
		require.NotEmpty(t, requestID)

		// Reply to the server that we can do the work.
		res := response.RollCall{
			Type:       blockless.MessageRollCallResponse,
			Code:       response.CodeAccepted,
			FunctionID: received.FunctionID,
			RequestID:  requestID,
		}

		rcPayload := serialize(t, res)

		// Mock worker should respond to an execution request.
		err = mockWorker.SendMessage(ctx, node.host.ID(), rcPayload)
		require.NoError(t, err)

		receiverWG.Wait()
		nodeWG.Wait()
	})
}
