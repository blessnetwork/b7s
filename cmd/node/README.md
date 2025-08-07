
# Bless Node 

## Description

Bless b7s Node is a peer-to-peer networking daemon for the Bless network.

A Node in the Bless network can have one of two roles - it can be a Head Node or a Worker Node.

In short, Worker Nodes are nodes that will be doing the actual execution of work within the Bless P2P network.
Worker Nodes do this by relying on the Bless Runtime.
Bless Runtime needs to be available locally on the machine where the Node is run.

Head Nodes are nodes that coordinate work between a number of Worker Nodes.
When a Head Node receives an execution request to execute a piece of work (a Bless Function), it will start a process of finding a Worker Node most suited to do this work.
Head Node does not need to have access to Bless Runtime.

Head Nodes also serve a REST API that can be used to query or trigger certain actions.

## Usage

There are two main ways of specifying configuration options - the CLI flags and a config file.
CLI flags override any configuration options in the config file.

List of supported CLI flags is listed below.

```console
Usage of b7s-node:
  -r, --role string                    role this node will have in the Bless protocol (head or worker) (default "worker")
  -c, --concurrency uint               maximum number of requests node will process in parallel (default 10)
      --boot-nodes strings             list of addresses that this node will connect to on startup, in multiaddr format
      --workspace string               directory that the node can use for file storage
      --load-attributes                node should try to load its attribute data from IPFS
      --topics strings                 topics node should subscribe to
      --db string                      path to the database used for persisting peer and function data
  -l, --log-level string               log level to use (default "info")
  -a, --address string                 address that the b7s host will use (default "0.0.0.0")
  -p, --port uint                      port that the b7s host will use
      --private-key string             private key that the b7s host will use
      --dialback-address string        external address that the b7s host will advertise
      --dialback-port uint             external port that the b7s host will advertise
  -w, --websocket                      should the node use websocket protocol for communication
      --websocket-port uint            port to use for websocket connections
      --websocket-dialback-port uint   external port that the b7s host will advertise for websocket connections
      --no-dialback-peers              start without dialing back peers from previous runs
      --must-reach-boot-nodes          halt node if we fail to reach boot nodes on start
      --disable-connection-limits      disable libp2p connection limits (experimental)
      --connection-count uint          maximum number of connections the b7s host will aim to have
      --rest-api string                address where the head node REST API will listen on
      --runtime-path string            Bless Runtime location (used by the worker node)
      --runtime-cli string             runtime CLI name (used by the worker node)
      --cpu-percentage-limit float     amount of CPU time allowed for Bless Functions in the 0-1 range, 1 being unlimited
      --memory-limit int               memory limit (kB) for Bless Functions
      --enable-tracing                 emit tracing data
      --tracing-grpc-endpoint string   tracing exporter GRPC endpoint
      --tracing-http-endpoint string   tracing exporter HTTP endpoint
      --enable-metrics                 emit metrics
      --prometheus-address string      address where prometheus metrics will be served
      --config string                  path to a config file
```

Alternatively to the CLI flags, you can create a YAML file and specify the parameters there.
All of the CLI flags have a corresponding config option in the YAML file.
In the config file, parameters are grouped based on the functionality they impact.

Example configuration for a worker node:

```yaml
role: worker
concurrency: 10
workspace: /tmp/workspace
load-attributes: false

log:
  level: debug

connectivity:
  address: 127.0.0.1
  port: 9000
  private-key: /home/user/.b7s/path/to/priv/key.bin
  websocket: true


worker:
  runtime-path: /home/user/.local/Bless-runtime/bin
  cpu-percentage-limit: 0.8

telemetry:
  tracing:
    enable: true
    http:
      endpoint:
  metrics:
    enable: true
    prometeus-address: 0.0.0.0:8080

```

You can find a more complete reference in [example.yaml](/cmd/node/example.yaml).

### Configuration Option Details

You can find more information about `multiaddr` format for network addresses [here](https://github.com/multiformats/multiaddr) and [here](https://multiformats.io/multiaddr/).

Private key path relates to the private key created by the [keygen](/cmd/keygen/README.md) utility.
Using the same private key in multiple `node` runs will ensure the node has the same identity on the network.
If a private key is not specified the node will start with a randomly generated identity.

## Examples

### Starting a Worker Node

```console
$ ./node --db /tmp/db --log-level debug --port 9000 --role worker --runtime ~/.local/bin --workspace workspace --private-key ./keys/priv.bin
```

The created `node` will listen on all addresses on TCP port 9000.
Database used to persist Node data between runs will be created in the `/tmp/db` subdirectory.

Bless Runtime path is given as `/home/user/.local/bin`.
At startup, node will check if the Bless Runtime is actually found there, namely the [bls-runtime](https://Bless.network/docs/protocol/runtime).

Node Identity will be determined by the private key found in `priv.bin` file in the `keys` subdirectory.

Any transient files needed for node operation will be created in the `workspace` subdirectory.

### Starting a Head Node

```console
$ ./node --db /var/tmp/b7s/db --log-level debug --port 9002 -r head --workspace /var/tmp/b7s/workspace --private-key ~/keys/priv.bin --rest-api ':8080'
```

The created `node` will listen on all addresses on TCP port 9002.
Database used to persist Node peer and function data between runs will be created at `/var/tmp/b7s/db`.

Any transient files needed for node operation will be created in the `/var/tmp/b7s/workspace` directory.

Head Node REST API will be available on all addresses on port 8080.
