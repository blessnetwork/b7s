#!/bin/bash
CONFIG_PATH=/app/keys

if [ -n "$KEY_PATH" ]; then
  s3_uri_base="s3://${KEY_PATH}"
  aws_args="--endpoint-url ${S3_HOST}"
  if [ -n "$KEY_PASSWORD" ]; then
    file_suffix=".gpg"
  else
    file_suffix=""
  fi
fi

restore_key () {
  existing=$(aws $aws_args s3 ls "${s3_uri_base}/$1" | head -n 1)
  if [[ -z $existing ]]; then
    echo "$1 backup not found"
  else
    echo "Restoring $1 to $CONFIG_PATH"
    aws $aws_args s3 cp "${s3_uri_base}/$1" $CONFIG_PATH/$1$file_suffix
    if [ -n "$KEY_PASSWORD" ]; then
      echo "Decrypting"
      gpg --decrypt --batch --passphrase "$KEY_PASSWORD" $CONFIG_PATH/$1$file_suffix > $CONFIG_PATH/$1
      rm $CONFIG_PATH/$1$file_suffix
    fi
  fi
}

backup_key () {
  existing=$(aws $aws_args s3 ls "${s3_uri_base}/$1" | head -n 1)
  if [[ -z $existing ]]; then
    echo "Backing up $1"
    if [ -n "$KEY_PASSWORD" ]; then
      echo "Encrypting backup..."
      gpg --symmetric --batch --passphrase "$KEY_PASSWORD" $CONFIG_PATH/$1
    fi
    aws $aws_args s3 cp $CONFIG_PATH/$1$file_suffix "${s3_uri_base}/$1"
    [ -n "$KEY_PASSWORD" ] && rm $CONFIG_PATH/$1.gpg
  fi
}

# Restore keys
if [ -n "$KEY_PATH" ]; then
  for f in $(aws $aws_args s3 ls "${s3_uri_base}/" | awk '{print $4}'); do
    cd /app/keys
    restore_key "$f"
  done

  if [ -n "$KEY_PASSWORD" ]; then
    if [ -f "/app/keys/wallet.key" ]; then
      echo $KEY_PASSWORD | blsd keys import node /app/keys/wallet.key --keyring-backend=test  --home=/app/.blockless-chain
    fi
  fi
fi

if [ -f "$CONFIG_PATH/identity" ]; then
  echo "Node restored from backup"
else 
  echo "Generating New Node Identity"
  cd /app/keys
  ../b7s-keyforge
  # Backup keys
  if [ -n "$KEY_PATH" ]; then
    # backup the on chain node identity
    for f in $(ls /app/keys/*); do
      backup_key $(basename "$f")
    done
  fi
fi

# run  template against the config file
# load env var array as a data source
# /app/gomplate -d 'boot_nodes=env:///BOOT_NODES?type=text/csv' -f /app/docker-config.yaml -o /app/docker-config-env.yaml

# run the node
cd /app


dialback_args=""
if [ -n "$DIALBACK_ADDRESS" ] && [ -n "$DIALBACK_PORT" ]; then
  dialback_args="--dialback-address $DIALBACK_ADDRESS --dialback-port $DIALBACK_PORT"
fi

bootnode_args=""
if [ -n "$BOOT_NODES" ]; then
  bootnode_args="--boot-nodes $BOOT_NODES"
fi

wss_args=""
if [ -n "$WSS_PORT" ]; then
  # not setting WSS dialback port
  wss_args="--websocket --websocket-port $WSS_PORT"
fi

metrics_args=""
if [ -n "$METRICS_PORT" ]; then
  metrics_args="--enable-metrics --prometheus-address :$METRICS_PORT"
fi


if [ "$NODE_ROLE" = "head" ]; then
  ./b7s --db /var/tmp/b7s/db --log-level debug --port $P2P_PORT --role head --workspace $WORKSPACE_ROOT --private-key $NODE_KEY_PATH --rest-api :$REST_API $dialback_args $bootnode_args $wss_args $metrics_args
else
  ./b7s --db ./db --log-level debug --port $P2P_PORT --role worker --runtime-path /app/runtime --runtime-cli bls-runtime --workspace $WORKSPACE_ROOT --private-key $NODE_KEY_PATH $dialback_args $bootnode_args $wss_args $metrics_args
fi
