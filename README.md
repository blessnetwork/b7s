[![Coverage](https://img.shields.io/badge/Coverage-64.5%25-yellow)](https://img.shields.io/badge/Coverage-64.5%25-yellow)
[![Go Report Card](https://goreportcard.com/badge/github.com/blessnetwork/b7s)](https://goreportcard.com/report/github.com/blessnetwork/b7s) 
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/blessnetwork/b7s/blob/main/LICENSE.md) 
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/blessnetwork/b7s)](https://img.shields.io/github/v/release/blessnetwork/b7s)


# b7s daemon

b7s is a peer-to-peer networking daemon for the bless network.
It is supported on Windows, Linux, and MacOS platforms for both x64 and arm64 architectures.

## Installation

You can install b7s using either curl or wget:

```bash
# using curl
sudo sh -c "curl https://raw.githubusercontent.com/blessnetwork/b7s/main/download.sh | bash"

# using wget
sudo sh -c "wget https://raw.githubusercontent.com/blessnetwork/b7s/main/download.sh -v -O download.sh; chmod +x download.sh; ./download.sh; rm -rf download.sh"
```

You can also use Docker to install b7s. See the [Docker documentation](/docker/README.md) for more information.

### Worker Node Installation

In order to successfully run a worker node you must also install [BLS runtime](https://github.com/blessnetwork/bls-runtime).

## Usage

For a more detailed overview of the node options and their meaning see [b7s-node Readme](/cmd/node/README.md#usage).
You can see an example YAML config file [here](/cmd/node/example.yaml).

### General Flags

| Flag                      | Short Form | Default Value           | Description                                                                             |
| ------------------------- | ---------- | ----------------------- | --------------------------------------------------------------------------------------- |
| config                    | N/A        | N/A                     | Config file to load.                                                                    |
| log-level                 | -l         | "info"                  | Level of logging to use.                                                                |
| db                        | N/A        | "db"                    | Path to database used for persisting peer and function data.                            |
| role                      | -r         | "worker"                | Role this node will have in the Bless protocol (head or worker).                    |
| workspace                 | N/A        | "./workspace"           | Directory that the node will use for file storage.                                      |
| concurrency               | -c         | 10                      | Maximum number of requests the node will process in parallel.                           |
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
| runtime-path              | N/A        | N/A                     | Local path to the Bless Runtime.                                                          |
| runtime-cli               | N/A        | "bls-runtime"           | Name of the Bless Runtime executable, as found in the runtime-path.                       |
| cpu-percentage-limit      | N/A        | 1.0                     | Amount of CPU time allowed for Bless Functions in the 0-1 range, 1 being unlimited (100%) |
| memory-limit              | N/A        | N/A                     | Memory limit for Bless Functions, in kB.                                                  |

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



## Private Key and Node Identity

A Node is known by its Peer ID.
Peer ID is an identifier that uniquely identifies a node in the network.
Peer ID is derived from the nodes private key.

To generate a private key for the node, use [keyforge](/cmd/keyforge/README.md).

## Dependencies

b7s depends on the following repositories:

- [BLS runtime](https://github.com/blessnetwork/bls-runtime)

## Architecture

```mermaid
flowchart TD
    %% External Entity
    ExternalClients["External Clients"]:::external

    %% API Layer
    subgraph "API Layer"
        API["API Endpoints"]:::api
    end

    %% Consensus Module
    subgraph "Consensus Module"
        PBFT["PBFT Module"]:::consensus
        Raft["Raft Module"]:::consensus
    end

    %% Core Node & Networking
    subgraph "Core Node & Networking"
        Head["Head Node"]:::node
        Worker["Worker Node"]:::node
        Host["Host (Discovery/Publisher)"]:::node
    end

    %% Executor/Worker
    subgraph "Executor/Worker"
        Executor["Executor Service"]:::executor
    end

    %% Storage
    subgraph "Storage"
        FStore["File Store"]:::storage
        Store["State Store"]:::storage
    end

    %% Cryptography
    subgraph "Cryptography"
        Crypto["Crypto Services"]:::crypto
    end

    %% Command-line Interfaces
    subgraph "Command-line Interfaces"
        CLI["CLI Tools"]:::cli
    end

    %% Connections
    ExternalClients -->|"calls"| API
    API -->|"processes"| Head
    Head -->|"delegates to"| Worker
    Head -->|"syncs with"| PBFT
    Head -->|"syncs with"| Raft
    Head -->|"delegates tasks"| Executor
    Head -->|"persists to"| FStore
    Head -->|"persists to"| Store
    API -->|"secureComm"| Crypto
    PBFT -->|"validateWith"| Crypto
    Raft -->|"validateWith"| Crypto
    Worker -->|"discoveryVia"| Host
    Host <-->|"p2pComm"| Head
    PBFT -->|"logsTo"| FStore
    Raft -->|"logsTo"| Store
    CLI -->|"commands"| Head

    %% Click Events
    click API "https://github.com/blessnetwork/b7s/tree/main/api/"
    click PBFT "https://github.com/blessnetwork/b7s/tree/main/consensus/pbft/"
    click Raft "https://github.com/blessnetwork/b7s/tree/main/consensus/raft/"
    click Head "https://github.com/blessnetwork/b7s/tree/main/node/"
    click Worker "https://github.com/blessnetwork/b7s/tree/main/node/"
    click Host "https://github.com/blessnetwork/b7s/tree/main/host/"
    click Executor "https://github.com/blessnetwork/b7s/tree/main/executor/"
    click FStore "https://github.com/blessnetwork/b7s/tree/main/fstore/"
    click Store "https://github.com/blessnetwork/b7s/tree/main/store/"
    click Crypto "https://github.com/blessnetwork/b7s/tree/main/crypto/"
    click CLI "https://github.com/blessnetwork/b7s/tree/main/cmd/"

    %% Styles
    classDef external fill:#fefefe,stroke:#333,stroke-width:2px;
    classDef api fill:#cce5ff,stroke:#1a75ff,stroke-width:2px;
    classDef consensus fill:#d1e7dd,stroke:#0f5132,stroke-width:2px;
    classDef node fill:#fff3cd,stroke:#ffc107,stroke-width:2px;
    classDef executor fill:#fde2e2,stroke:#d9534f,stroke-width:2px;
    classDef storage fill:#e2e3e5,stroke:#6c757d,stroke-width:2px;
    classDef crypto fill:#e2f0d9,stroke:#28a745,stroke-width:2px;
    classDef cli fill:#f8d7da,stroke:#dc3545,stroke-width:2px;
```

## Contributing

See src/README for information on contributing to the b7s project.
