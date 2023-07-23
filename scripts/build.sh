#!/bin/bash

TARGET_DIR="./bin"
TARGET="$TARGET_DIR/xtund"

# Create the target directory if it does not exist
mkdir -p $TARGET_DIR

# Remove the target binary file if it exists
if [ -f "$TARGET" ]; then
    rm $TARGET
fi

echo "Building..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o $TARGET ./cmd/main.go

if [ $? -eq 0 ]; then
    echo "Build status: OK, binary at $TARGET"
else
    echo "Build failed."
fi
