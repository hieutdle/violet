#!/bin/bash

# Set default values
OUTPUT="violet"
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

# Function to display the usage
usage() {
    echo "Usage: $0 [-o output_name] [-s target_os] [-a target_arch]"
    echo "  -o output_name: Name of the output file (default: violet)"
    echo "  -s target_os: Target OS (default: $GOOS)"
    echo "  -a target_arch: Target architecture (default: $GOARCH)"
}

# Parse the command line arguments
while getopts "o:s:a:" opt; do
    case $opt in
        o) OUTPUT=$OPTARG ;;
        s) GOOS=$OPTARG ;;
        a) GOARCH=$OPTARG ;;
        \?) echo "Invalid option: -$OPTARG" >&2
            usage
            exit 1 ;;
        :) echo "Option -$OPTARG requires an argument." >&2
            usage
            exit 1 ;;
    esac
done

# Set GOOS and GOARCH environment variables
export GOOS=$GOOS
export GOARCH=$GOARCH

echo "Building the project for OS: $GOOS and Architecture: $GOARCH"
go build -buildvcs=false -o "$OUTPUT"

if [ $? -eq 0 ]; then
    echo "Build successful! Output file: $OUTPUT"
else
    echo "Build failed!"
    exit 1
fi
