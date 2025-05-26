## Prerequisites

- A machine running Docker
- A valid AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY for an S3-compatible storage provider. This is used for backing up your node's keys and configuration.
- A valid KEY_PATH and KEY_PASSWORD for your S3-compatible storage provider. This is used for backing up your node's keys and configuration.

## Running the Image

First, you'll need to pull the latest version of the b7s Docker image from the public registry:

```bash
docker pull ghcr.io/blocklessnetwork/b7s:v0.6.6.patch2
```

Run the image

To Run the Node in the Head Node Configuration

```bash
docker run -d --name b7s \
  -e AWS_ACCESS_KEY_ID=<YOUR_AWS_ACCESS_KEY_ID> \
  -e AWS_SECRET_ACCESS_KEY=<YOUR_AWS_SECRET_ACCESS_KEY> \
  -e KEY_PATH=<YOUR_S3_KEY_PATH> \
  -e KEY_PASSWORD=<YOUR_S3_KEY_PASSWORD> \
  -e NODE_ROLE=head \
  -e P2P_PORT=9527 \
  -e REST_API=8081 \
  -e DIALBACK_PORT=32342 \
  -e DIALBACK_ADDRESS=1.1.1.1 \
  -v /var/tmp/b7s/db:/var/tmp/b7s/db \
  -p 9527:9527 \
  ghcr.io/blocklessnetwork/b7s:v0.6.6.patch2
```
To Run the Node in the Worker Node Configuration

```bash
docker run -d --name b7s \
  -e AWS_ACCESS_KEY_ID=<YOUR_AWS_ACCESS_KEY_ID> \
  -e AWS_SECRET_ACCESS_KEY=<YOUR_AWS_SECRET_ACCESS_KEY> \
  -e KEY_PATH=<YOUR_S3_KEY_PATH> \
  -e KEY_PASSWORD=<YOUR_S3_KEY_PASSWORD> \
  -e NODE_ROLE=worker \
  -e P2P_PORT=9527 \
  -e BOOT_NODES=<HEAD_NODE_MULTIADDRESS> \
  -v /var/tmp/b7s/db:/var/tmp/b7s/db \
  -p 9527:9527 \
  ghcr.io/blocklessnetwork/b7s:v0.6.6.patch2
```

In order to pass in a private key to use with the worker:

```bash
docker run -d --name b7s \
  -e NODE_ROLE=worker \
  -e P2P_PORT=9527 \
  -e BOOT_NODES=<HEAD_NODE_MULTIADDRESS> \
  -e NODE_KEY_PATH=/app/worker/key.bin \
  -v /path/to/private/key.bin:/app/worker/key.bin \
  -v /var/tmp/b7s/db:/var/tmp/b7s/db \
```
