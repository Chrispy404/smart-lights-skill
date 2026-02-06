#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/hue-control"

echo "Building hue-control..."
go build -o hue-control .
chmod +x hue-control

echo "Build complete: $SCRIPT_DIR/hue-control/hue-control"
