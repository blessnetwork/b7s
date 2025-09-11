package node

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/go-metrics"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/models/bls"
)

// Run will start the main loop for the node.
func (c *core) Run(ctx context.Context, process func(context.Context, peer.ID, string, []byte) error) error {

	err := c.host.InitPubSub(ctx)
	if err != nil {
		return fmt.Errorf("coould not initialize pubsub: %w", err)
	}

	topics := c.cfg.Topics
	for _, topic := range topics {
		err = c.Subscribe(ctx, topic)
		if err != nil {
			return fmt.Errorf("could not subscribe to topic (topic: %s): %w", topic, err)
		}
	}

	err = c.host.ConnectToKnownPeers(ctx)
	if err != nil {
		return fmt.Errorf("could not connect to known peers: %w", err)
	}

	// Set the handler for direct messages.
	c.listenDirectMessages(ctx, process)

	// Discover peers.
	// NOTE: Potentially signal any error here so that we abort the node
	// run loop if anything failed.
	for _, topic := range topics {
		go func(topic string) {

			// TODO: Check DHT initialization, now that we're working with multiple topics, may not need to repeat ALL work per topic.
			err := c.host.DiscoverPeers(ctx, topic)
			if err != nil {
				c.Log().Error().Err(err).Msg("could not discover peers")
			}
		}(topic)
	}

	c.log.Info().Uint("concurrency", c.cfg.Concurrency).Msg("starting node main loop")

	var (
		workers sync.WaitGroup
		wg      sync.WaitGroup
		sema    chan struct{} = make(chan struct{}, c.cfg.Concurrency)
	)

	// Process topic messages - spin up a goroutine for each topic that will feed the main processing loop below.
	// No need for locking since we're still single threaded here and these (subscribed) topics will not be touched by other code.
	for _, topicName := range c.topics.Keys() {

		topic, _ := c.topics.Get(topicName)

		workers.Add(1)

		go func(name string, subscription *pubsub.Subscription) {
			defer workers.Done()

			// Message processing loops.
			for {

				// Retrieve next message.
				msg, err := subscription.Next(ctx)
				if err != nil {
					// NOTE: Cancelling the context will lead us here.
					c.log.Error().Err(err).Msg("could not receive message")
					break
				}

				// Skip messages we published.
				if msg.GetFrom() == c.host.ID() {
					continue
				}

				c.log.Trace().
					Str("topic", name).
					Stringer("peer", msg.ReceivedFrom).
					Stringer("origin", msg.GetFrom()).
					Hex("id", []byte(msg.ID)).Msg("received message")

				// Try to get a slot for processing the request.
				sema <- struct{}{}
				wg.Add(1)

				go func(msg *pubsub.Message) {
					// Free up slot after we're done.
					defer wg.Done()
					defer func() { <-sema }()

					c.metrics.IncrCounterWithLabels(topicMessagesMetric, 1, []metrics.Label{{Name: "topic", Value: name}})

					err = c.processMessage(ctx, msg.GetFrom(), msg.GetData(), PubSubPipeline(name), process)
					if err != nil {
						c.log.Error().Err(err).Str("id", msg.ID).Str("peer", msg.ReceivedFrom.String()).Msg("could not process message")
						return
					}

				}(msg)
			}
		}(topicName, topic.subscription)
	}

	workers.Wait()

	c.log.Debug().Msg("waiting for messages being processed")
	wg.Wait()

	// Start the health signal emitter in a separate goroutine.
	go c.emitHealthPing(ctx, c.cfg.HealthInterval)

	return nil
}

// listenDirectMessages will process messages sent directly to the peer (as opposed to published messages).
func (c *core) listenDirectMessages(ctx context.Context, process func(context.Context, peer.ID, string, []byte) error) {
	c.host.SetStreamHandler(bls.ProtocolID, func(stream network.Stream) {
		defer stream.Close()

		from := stream.Conn().RemotePeer()

		c.metrics.IncrCounter(directMessagesMetric, 1)

		buf := bufio.NewReader(stream)
		msg, err := buf.ReadBytes('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			stream.Reset()
			c.log.Error().Err(err).Msg("error receiving direct message")
			return
		}

		c.log.Trace().Stringer("peer", from).Msg("received direct message")

		err = c.processMessage(ctx, from, msg, DirectMessagePipeline, process)
		if err != nil {
			c.log.Error().Err(err).Str("peer", from.String()).Msg("could not process direct message")
			return
		}
	})
}
