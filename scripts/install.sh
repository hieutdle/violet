#!/bin/bash

# Set the binary name and installation path
BINARY_NAME="violet"
BUILD_DIR=$(pwd)/scripts
PROJECT_ROOT=$(dirname "$BUILD_DIR")
INSTALL_PATH="/usr/local/bin"

# Function to install the binary
install_binary() {
    if [ -f "$PROJECT_ROOT/$BINARY_NAME" ]; then
        echo "Installing the binary to $INSTALL_PATH"
        sudo sudo mv "$PROJECT_ROOT/$BINARY_NAME" $INSTALL_PATH/$BINARY_NAME
        sudo chmod +x $INSTALL_PATH/$BINARY_NAME
        echo "Binary installed successfully!"
    else
        echo "Binary not found. Build failed!"
        exit 1
    fi
}

# Check if Go is installed
if command -v go &> /dev/null; then
    echo "Go is installed. Building the binary using build.sh..."

# Navigate to the project root and run build.sh
    cd $PROJECT_ROOT
    ./scripts/build.sh

# Install the binary
    install_binary
fi