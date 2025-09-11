package node

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-metrics"
	"github.com/hashicorp/go-multierror"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/models/bls"
)

type topicInfo struct {
	handle       *pubsub.Topic
	subscription *pubsub.Subscription
}

func (c *core) Subscribe(ctx context.Context, topic string) error {

	h, sub, err := c.host.Subscribe(topic)
	if err != nil {
		return fmt.Errorf("could not subscribe to topic: %w", err)
	}

	ti := topicInfo{
		handle:       h,
		subscription: sub,
	}

	c.topics.Set(topic, ti)

	c.metrics.IncrCounter(subscriptionsMetric, 1)

	return nil
}

// send serializes the message and sends it to the specified peer.
func (c *core) Send(ctx context.Context, to peer.ID, msg bls.Message) error {

	opts := new(messageSpanConfig).pipeline(DirectMessagePipeline).peer(to).spanOpts()
	ctx, span := c.tracer.Start(ctx, msgSendSpanName(spanMessageSend, msg.Type()), opts...)
	defer span.End()

	saveTraceContext(ctx, msg)

	// Serialize the message.
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not encode record: %w", err)
	}

	// Send message.
	err = c.host.SendMessage(ctx, to, payload)
	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	c.metrics.IncrCounterWithLabels(messagesSentMetric, 1, []metrics.Label{{Name: "type", Value: msg.Type()}})

	return nil
}

// sendToMany serializes the message and sends it to a number of peers. `requireAll` dictates how we treat partial errors.
func (c *core) SendToMany(ctx context.Context, peers []peer.ID, msg bls.Message, requireAll bool) error {

	opts := new(messageSpanConfig).pipeline(DirectMessagePipeline).peers(peers...).spanOpts()
	ctx, span := c.tracer.Start(ctx, msgSendSpanName(spanMessageSend, msg.Type()), opts...)
	defer span.End()

	saveTraceContext(ctx, msg)

	// Serialize the message.
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not encode record: %w", err)
	}

	var eg multierror.Group
	for i, peer := range peers {
		i := i
		peer := peer

		eg.Go(func() error {
			err := c.host.SendMessage(ctx, peer, payload)
			if err != nil {
				return fmt.Errorf("peer %v/%v send error (peer: %s): %w", i+1, len(peers), peer, err)
			}

			return nil
		})
	}

	// Might be off in the case of partial success.
	c.metrics.IncrCounterWithLabels(messagesSentMetric, float32(len(peers)), []metrics.Label{{Name: "type", Value: msg.Type()}})

	retErr := eg.Wait()
	if retErr == nil || len(retErr.Errors) == 0 {
		// If everything succeeded => ok.
		return nil
	}

	switch len(retErr.Errors) {
	case len(peers):
		// If everything failed => error.
		return fmt.Errorf("all sends failed: %w", retErr)

	default:
		// Some sends failed - do as requested by `requireAll`.
		if requireAll {
			return fmt.Errorf("some sends failed: %w", retErr)
		}

		c.log.Warn().Err(retErr.ErrorOrNil()).Msg("some sends failed, proceeding")

		return nil
	}
}

func (c *core) Publish(ctx context.Context, msg bls.Message) error {
	return c.PublishToTopic(ctx, bls.DefaultTopic, msg)
}

func (c *core) PublishToTopic(ctx context.Context, topic string, msg bls.Message) error {

	opts := new(messageSpanConfig).pipeline(PubSubPipeline(topic)).spanOpts()
	ctx, span := c.tracer.Start(ctx, msgSendSpanName(spanMessagePublish, msg.Type()), opts...)
	defer span.End()

	saveTraceContext(ctx, msg)

	// Serialize the message.
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not encode record: %w", err)
	}

	topicInfo, ok := c.topics.Get(topic)
	if !ok {
		err = c.JoinTopic(topic)
		if err != nil {
			return fmt.Errorf("could not join topic (topic: %s): %w", topic, err)
		}
	}

	// Publish message.
	err = c.host.Publish(ctx, topicInfo.handle, payload)
	if err != nil {
		return fmt.Errorf("could not publish message: %w", err)
	}

	c.metrics.IncrCounterWithLabels(messagesPublishedMetric, 1,
		[]metrics.Label{
			{Name: "type", Value: msg.Type()},
			{Name: "topic", Value: topic},
		})

	return nil
}

// wrapper around topic joining + housekeeping.
func (c *core) JoinTopic(topic string) error {

	th, err := c.host.JoinTopic(topic)
	if err != nil {
		return fmt.Errorf("could not join topic (topic: %s): %w", topic, err)
	}

	ti := topicInfo{
		handle:       th,
		subscription: nil, // NOTE: No subscription, joining topic only.
	}

	c.topics.Set(topic, ti)

	return nil
}

func (c *core) Connected(peer peer.ID) bool {
	connections := c.host.Network().ConnsToPeer(peer)
	return len(connections) > 0
}
