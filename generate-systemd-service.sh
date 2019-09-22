#!/bin/bash

BINARY_PATH="${BINARY_PATH:-/usr/local/bin/quickdoc}"

cat <<-EOF
[Unit]
Description=quickdoc - show local documentation quickly

[Service]
DynamicUser=true
ExecStart=$BINARY_PATH
ProtectSystem=strict
SystemCallFilter=@io-event @network-io @process @system-service
WorkingDirectory=/tmp
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF
