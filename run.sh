#!/bin/bash
# run.sh
set -e

# Default values
PORT=${1:-8080}
MAX_CONNECTIONS=${2:-1000000}

# Check if the server binary exists
if [ ! -f "./server" ]; then
    echo "Server binary not found. Running build script..."
    ./build.sh
fi

# Set system limits for high concurrent connections
ulimit -n 1048576
sysctl -w net.core.somaxconn=65535 >/dev/null 2>&1 || true
sysctl -w net.ipv4.tcp_max_syn_backlog=65535 >/dev/null 2>&1 || true

echo "Starting server on port $PORT with max connections: $MAX_CONNECTIONS"
./server -port $PORT -max-connections $MAX_CONNECTIONS
