#!/bin/bash

echo "Building Atom interpreter..."

# Detect OS and set appropriate executable name
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    EXECUTABLE="atom.linux"
    echo "Building for Linux..."
elif [[ "$OSTYPE" == "darwin"* ]]; then
    EXECUTABLE="atom.macos"
    echo "Building for macOS..."
elif [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
    EXECUTABLE="atom.exe"
    echo "Building for Windows..."
else
    EXECUTABLE="atom.unknown"
    echo "Building for unknown OS, using default executable name..."
fi

# Build the Go application
go build -o "$EXECUTABLE" .

echo "Build complete! $EXECUTABLE created."