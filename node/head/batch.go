package head

import (
	"context"
	"strings"

	"github.com/blessnetwork/b7s/models/request"
	"github.com/hashicorp/go-multierror"
	"github.com/libp2p/go-libp2p/core/peer"
)

type sendError struct {
	target peer.ID
}

func sendErr(p peer.ID) *sendError {
	err := &sendError{
		target: p,
	}

	return err
}

func (e *sendError) Error() string {
	return "send error"
}

type batchSendError struct {
	errors []*sendError
}

func newBatchSendError(errs ...error) *batchSendError {

	if len(errs) == 0 {
		return nil
	}

	outErr := &batchSendError{
		errors: make([]*sendError, len(errs)),
	}

	for i := range errs {
		se, ok := errs[i].(*sendError)
		if !ok {
			continue
		}

		outErr.errors[i] = se
	}

	return outErr
}

func (e *batchSendError) Error() string {
	strs := make([]string, len(e.errors))
	for i := range e.errors {
		strs[i] = e.errors[i].Error()
	}

	return strings.Join(strs, "\n")
}

func (e *batchSendError) Targets() []peer.ID {
	out := make([]peer.ID, len(e.errors))
	for i := range e.errors {
		out[i] = e.errors[i].target
	}

	return out
}

func (h *HeadNode) sendBatch(ctx context.Context, assignments map[peer.ID]*request.WorkOrderBatch) error {

	var eg multierror.Group
	for peer, w := range assignments {
		eg.Go(func() error {
			err := h.Send(ctx, peer, w)
			if err != nil {
				return sendErr(peer)
			}
			return nil
		})
	}

	err := eg.Wait()
	if err == nil || len(err.Errors) == 0 {
		// If everything succeeded => ok.
		return nil
	}

	return newBatchSendError(err.Errors...)
}
