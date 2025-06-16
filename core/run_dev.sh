#!/bin/bash

echo "Compiling billionmail..."

# Using the alpine image to compile the Go application
docker exec p-g-alpine sh -c "cd /opt/core && sh ./go-build.sh"

echo "Copying billionmail to billionmail-core-billionmail-1 container..."

# Determine the architecture and set the binary name accordingly
ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
    BINARY="billionmail-amd64"
    echo "Detect x86_64 architecture, using amd64 binary"
elif [[ "$ARCH" == "arm64" || "$ARCH" == "aarch64" ]]; then
    BINARY="billionmail-arm64"
    echo "Detect arm64/aarch64 architecture, using arm64 binary"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

echo "Copying the compiled binary from p-g-alpine to billionmail-core-billionmail-1..."

# Copy the compiled binary to the billionmail-core-billionmail-1 container
docker cp $BINARY billionmail-core-billionmail-1:/opt/billionmail/core/billionmail


echo "Removing the public/dist/ directory from billionmail-core-billionmail-1 container..."

# Remove the public/dist/ directory from the billionmail-core-billionmail-1 container
docker exec billionmail-core-billionmail-1 sh -c "rm -rf /opt/billionmail/core/public/dist"


echo "Copying the public/ directory to billionmail-core-billionmail-1 container..."

# Copy the public directory to the billionmail-core-billionmail-1 container
docker cp public/. billionmail-core-billionmail-1:/opt/billionmail/core/public/


echo "Copying the manifest/ directory to billionmail-core-billionmail-1 container..."

# Copy the manifest/ directory to the billionmail-core-billionmail-1 container
docker cp manifest/. billionmail-core-billionmail-1:/opt/billionmail/core/manifest/


echo "Copying the template/ directory to billionmail-core-billionmail-1 container..."

# Copy the template/ directory to the billionmail-core-billionmail-1 container
docker cp template/. billionmail-core-billionmail-1:/opt/billionmail/core/template/


echo "Restarting billionmail-core-billionmail-1 container..."

# Restart the billionmail-core-billionmail-1 container to apply changes
docker restart billionmail-core-billionmail-1

echo "billionmail has been successfully compiled and deployed."