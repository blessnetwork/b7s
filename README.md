[![Coverage](https://img.shields.io/badge/Coverage-64.5%25-yellow)](https://img.shields.io/badge/Coverage-64.5%25-yellow)
[![Go Report Card](https://goreportcard.com/badge/github.com/blocklessnetwork/b7s)](https://goreportcard.com/report/github.com/blocklessnetwork/b7s) 
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/blocklessnetwork/b7s/blob/main/LICENSE.md) 
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/blocklessnetwork/b7s)](https://img.shields.io/github/v/release/blocklessnetwork/b7s)


# b7s daemon

b7s is a peer-to-peer networking daemon for the blockless network. It is supported on Windows, Linux, and MacOS platforms for both x64 and arm64 architectures.

## Installation

You can install b7s using either curl or wget:

```bash
# using curl
sudo sh -c "curl https://raw.githubusercontent.com/blessnetwork/b7s/main/download.sh | bash"

# using wget
sudo sh -c "wget https://raw.githubusercontent.com/blessnetwork/b7s/main/download.sh -v -O download.sh; chmod +x download.sh; ./download.sh; rm -rf download.sh"
```

You can also use Docker to install b7s. See the [Docker documentation](/docker/README.md) for more information.

## Usage

For a more detailed overview of the node options and their meaning see [b7s-node Readme](/cmd/node/README.md#usage).
You can see an example YAML config file [here](/cmd/node/example.yaml).

### General Flags

| Flag                      | Short Form | Default Value           | Description                                                                             |
| ------------------------- | ---------- | ----------------------- | --------------------------------------------------------------------------------------- |
| config                    | N/A        | N/A                     | Config file to load.                                                                    |
| log-level                 | -l         | "info"                  | Level of logging to use.                                                                |
| db                        | N/A        | "db"                    | Path to database used for persisting peer and function data.                            |
| role                      | -r         | "worker"                | Role this node will have in the Blockless protocol (head or worker).                    |
| workspace                 | N/A        | "./workspace"           | Directory that the node will use for file storage.                                      |
| concurrency               | -c         | node.DefaultConcurrency | Maximum number of requests the node will process in parallel.                           |
| load-attributes           | N/A        | false                   | Load attributes from the environment.                                                   |
| topics                    | N/A        | N/A                     | Topics that the node should subscribe to.                                                |

### Connectivity

| Flag                      | Short Form | Default Value           | Description                                                                             |
| ------------------------- | ---------- | ----------------------- | --------------------------------------------------------------------------------------- |
| address                   | -a         | "0.0.0.0"               | Address that the libp2p host will use.                                                  |
| port                      | -p         | 0                       | Port that the libp2p host will use.                                                     |
| private-key               | N/A        | N/A                     | Private key that the libp2p host will use.                                              |
| boot-nodes                | N/A        | N/A                     | List of addresses that this node will connect to on startup, in multiaddr format.       |
| websocket                 | -w         | false                   | Use websocket protocol for communication, besides TCP.                                  |
| websocket-port            | N/A        | 0                       | Port that the libp2p host will use for websocket connections.                           |
| dialback-address          | N/A        | N/A                     | Advertised dialback address of the Node.                                                |
| dialback-port             | N/A        | N/A                     | Advertised dialback port of the Node.                                                   |
| websocket-dialback-port   | N/A        | 0                       | Advertised dialback port for Websocket connections.                                     |
| no-dialback-peers         | N/A        | false                   | Avoid dialing back peers known from past runs                                           |
| must-reach-boot-nodes     | N/A        | false                   | Halt if the Node cannot reach boot nodes on startup.                                    |
| disable-connection-limits | N/A        | false                   | Try to maintain as many connections as possible.                                        |
| connection-count          | N/A        | N/A                     | Number of connections that the node will aim to have.                                   |

### Worker Node

| Flag                      | Short Form | Default Value           | Description                                                                                   |
| ------------------------- | ---------- | ----------------------- | --------------------------------------------------------------------------------------------- |
| runtime-path              | N/A        | N/A                     | Local path to the Blockless Runtime.                                                          |
| runtime-cli               | N/A        | "bls-runtime"           | Name of the Blockless Runtime executable, as found in the runtime-path.                       |
| cpu-percentage-limit      | N/A        | 1.0                     | Amount of CPU time allowed for Blockless Functions in the 0-1 range, 1 being unlimited (100%) |
| memory-limit              | N/A        | N/A                     | Memory limit for Blockless Functions, in kB.                                                  |

### Head Node

| Flag                      | Short Form | Default Value           | Description                                                                             |
| ------------------------- | ---------- | ----------------------- | --------------------------------------------------------------------------------------- |
| rest-api                  | N/A        | N/A                     | Address where the head node will serve the REST API                                     |

### Telemetry

| Flag                      | Short Form | Default Value           | Description                                                                             |
| ------------------------- | ---------- | ----------------------- | --------------------------------------------------------------------------------------- |
| enable-tracing            | N/A        | false                   | Emit tracing data.                                                                      |
| tracing-grpc-endpoint     | N/A        | N/A                     | GRPC endpoint where node should send tracing data.                                      |
| tracing-http-endpoint     | N/A        | N/A                     | HTTP endpoint where node should send tracing data.                                      |
| enable-metrics            | N/A        | false                   | Enable metrics.                                                                         |
| prometheus-address        | N/A        | N/A                     | Address where node should serve metrics (for head node this is the REST API address)    |

## Dependencies

b7s depends on the following repositories:

- blessnetwork/runtime
- blessnetwork/orchestration-chain

## Contributing

See src/README for information on contributing to the b7s project.
