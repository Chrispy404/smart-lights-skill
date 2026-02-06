#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/hue-control"

echo "Building hue-control..."
go build -o hue-control .
chmod +x hue-control

cd "$SCRIPT_DIR/weather-lights"

echo "Building weather-lights..."
go build -o weather-lights .
chmod +x weather-lights

echo "Build complete:"
echo "  - $SCRIPT_DIR/hue-control/hue-control"
echo "  - $SCRIPT_DIR/weather-lights/weather-lights"
