package head

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/codes"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/request"
	"github.com/blessnetwork/b7s/models/response"
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

func (h *HeadNode) ExecuteFunctionBatch(ctx context.Context, req request.ExecuteBatch) (*response.ExecuteBatch, error) {

	requestID := newRequestID()

	log := h.Log().With().
		Str("request", requestID).
		Str("function", req.Template.FunctionID).
		Int("size", len(req.Arguments)).Logger()

	log.Info().Msg("processing batch execution request via API")

	results, err := h.executeBatch(ctx, requestID, req)
	if err != nil {
		return nil, fmt.Errorf("could not execute batch request: %w", err)
	}

	log.Info().Any("results", results).Msg("received batch responses")

	// TODO: Add actual status code.
	res := req.Response(codes.OK, requestID).WithResults(results)

	return res, nil
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
