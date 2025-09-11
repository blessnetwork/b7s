package fstore

import (
	"net/http"
	"sync"

	"github.com/cavaliergopher/grab/v3"
	"github.com/hashicorp/go-metrics"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/blessnetwork/b7s/models/bls"
)

// FStore - function store - deals with all of the function-related actions - saving/reading them from backing storage,
// downloading them, unpacking them etc.
type FStore struct {
	log        zerolog.Logger
	store      bls.FunctionStore
	http       *http.Client
	downloader *grab.Client

	functionCount sync.Once

	workdir string
	tracer  trace.Tracer
	metrics *metrics.Metrics
}

// New creates a new function store.
func New(log zerolog.Logger, store bls.FunctionStore, workdir string) *FStore {

	// Create an HTTP client.
	cli := &http.Client{
		Timeout:   defaultTimeout,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	// Create a download client.
	downloader := grab.NewClient()
	downloader.UserAgent = defaultUserAgent
	downloader.HTTPClient = cli

	h := FStore{
		log:        log,
		store:      store,
		http:       cli,
		downloader: downloader,
		workdir:    workdir,
		tracer:     otel.Tracer(tracerName),
		metrics:    metrics.Default(),
	}

	return &h
}
