package head

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/codes"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/request"
	"github.com/blessnetwork/b7s/models/response"
	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
)

// ExecuteFunction can be used to start function execution. At the moment this is used by the API server to start execution on the head node.
func (h *HeadNode) ExecuteFunction(ctx context.Context, req execute.Request, subgroup string) (codes.Code, string, execute.ResultMap, execute.Cluster, error) {

	requestID := newRequestID()

	code, results, cluster, err := h.execute(ctx, requestID, request.Execute{Request: req})
	if err != nil {
		h.Log().Error().Str("request", requestID).Err(err).Msg("execution failed")
	}

	return code, requestID, results, cluster, nil
}

func (h *HeadNode) StartFunctionBatchExecution(ctx context.Context, req request.ExecuteBatch) (string, error) {

	requestID := newRequestID()

	log := h.Log().With().
		Str("request", requestID).
		Str("function", req.Template.FunctionID).
		Int("size", len(req.Arguments)).Logger()

	log.Info().Msg("processing batch execution request via API")

	// Persist batch and work items.
	err := h.saveBatch(requestID, req)
	if err != nil {
		return "", fmt.Errorf("could not save batch request: %w", err)
	}

	go func() {
		err := h.startBatchExecution(context.Background(), requestID, req)
		if err != nil {
			h.Log().Error().Err(err).Str("batch", requestID).Msg("could not execute batch")
		}
	}()

	return requestID, nil
}

// ExecutionResult fetches the execution result from the node cache.
func (h *HeadNode) ExecutionResult(id string) (execute.ResultMap, bool) {
	// TBD: Head node currently does not cache results.
	return nil, false
}

// PublishFunctionInstall publishes a function install message.
func (h *HeadNode) PublishFunctionInstall(ctx context.Context, uri string, cid string, subgroup string) error {

	var req request.InstallFunction
	if uri != "" {
		var err error
		req, err = createInstallMessageFromURI(uri)
		if err != nil {
			return fmt.Errorf("could not create install message from URI: %W", err)
		}
	} else {
		req = createInstallMessageFromCID(cid)
	}

	if subgroup == "" {
		subgroup = bls.DefaultTopic
	}

	h.Log().Debug().Str("subgroup", subgroup).Str("url", req.ManifestURL).Str("cid", req.CID).Msg("publishing function install message")

	err := h.PublishToTopic(ctx, subgroup, &req)
	if err != nil {
		return fmt.Errorf("could not publish message: %w", err)
	}

	return nil
}

func (h *HeadNode) GetBatchResults(ctx context.Context, id string) (*response.ExecuteBatch, error) {

	batch, err := h.cfg.BatchStore.GetBatch(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve batch result: %w", err)
	}

	// We will need to group work items according to the group they belong to.
	chunks, err := h.cfg.BatchStore.FindChunks(ctx, batch.ID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve chunks for batch: %w", err)
	}

	// Find all work items belonging to this batch.
	items, err := h.cfg.BatchStore.FindWorkItems(ctx, batch.ID, "")
	if err != nil {
		return nil, fmt.Errorf("could not retrieve work items for batch: %w", err)
	}

	lookup := make(map[string]*batchstore.ChunkRecord)
	for _, chunk := range chunks {
		lookup[chunk.ID] = chunk
	}

	oc := make(map[string]response.NodeChunkResults)
	for _, item := range items {

		chunk, ok := lookup[item.ChunkID]
		if !ok {
			h.Log().Warn().Str("batch", batch.ID).Str("work_item", item.ID).Str("chunk", item.ChunkID).
				Msg("chunk not found for work item")
			continue
		}

		_, ok = oc[item.ChunkID]
		if !ok {

			id, err := peer.Decode(chunk.Worker)
			if err != nil {
				return nil, fmt.Errorf("invalid peer ID found (id: %s): %w", chunk.Worker, err)
			}

			oc[item.ChunkID] = response.NodeChunkResults{
				Peer:    id,
				Results: make(map[execute.RequestHash]*response.BatchFunctionResult),
			}
		}

		hash := execute.ExecutionID(batch.CID, batch.Method, item.Arguments)
		oc[item.ChunkID].Results[hash] = &response.BatchFunctionResult{
			NodeResult: execute.NodeResult{
				Result: execute.Result{Result: execute.RuntimeOutput{
					Stdout: item.Output,
				}},
			},
			FunctionInvocation: execute.FunctionInvocation(batch.CID, batch.Method),
			Arguments:          item.Arguments,
		}
	}

	out := &response.ExecuteBatch{
		RequestID: id,
		Code:      codes.OK, // TODO: Be more precise in this, not all executions are "OK".
		Chunks:    oc,
	}

	return out, nil
}

// createInstallMessageFromURI creates a MsgInstallFunction from the given URI.
// CID is calculated as a SHA-256 hash of the URI.
func createInstallMessageFromURI(uri string) (request.InstallFunction, error) {

	cid, err := deriveCIDFromURI(uri)
	if err != nil {
		return request.InstallFunction{}, fmt.Errorf("could not determine cid: %w", err)
	}

	msg := request.InstallFunction{
		ManifestURL: uri,
		CID:         cid,
	}

	return msg, nil
}

// createInstallMessageFromCID creates the MsgInstallFunction from the given CID.
func createInstallMessageFromCID(cid string) request.InstallFunction {

	req := request.InstallFunction{
		ManifestURL: manifestURLFromCID(cid),
		CID:         cid,
	}

	return req
}

func deriveCIDFromURI(uri string) (string, error) {

	h := sha256.New()
	_, err := h.Write([]byte(uri))
	if err != nil {
		return "", fmt.Errorf("could not calculate hash: %w", err)
	}
	cid := fmt.Sprintf("%x", h.Sum(nil))

	return cid, nil
}

func manifestURLFromCID(cid string) string {
	return fmt.Sprintf("https://%s.ipfs.w3s.link/manifest.json", cid)
}
