package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

var (
	defaultCompressors = []string{"snappy", "zlib", "zstd"}
	defaultPingTimeout = time.Second * 2
)

func Connect(ctx context.Context, serverURL string) (*mongo.Client, error) {

	opts := options.Client().
		SetCompressors(defaultCompressors).
		ApplyURI(serverURL)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}

	pingctx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()

	err = client.Ping(pingctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return client, nil
}
