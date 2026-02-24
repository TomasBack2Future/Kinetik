#!/bin/bash

# Build script for local Docker image
set -e

echo "Building kinetik-automation:latest Docker image..."

# Build the image with the latest tag
docker build -t kinetik-automation:latest .

echo "✓ Image built successfully!"
echo ""
echo "Image details:"
docker images | grep kinetik-automation | head -n 1
echo ""
echo "You can now run: docker-compose up -d"
