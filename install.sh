#!/bin/bash

set -e

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

VERSION="latest"
if [ "$1" != "" ]; then
    VERSION="$1"
fi

echo "Installing Orb VCS for $OS-$ARCH..."

# Download and install orb
ORB_URL="https://github.com/ayushsarode/orb/releases/download/$VERSION/orb-$OS-$ARCH"
if [ "$OS" = "windows" ]; then
    ORB_URL="${ORB_URL}.exe"
fi

echo "Downloading from: $ORB_URL"
curl -L "$ORB_URL" -o orb
chmod +x orb
sudo mv orb /usr/local/bin/

echo "Orb VCS installed successfully!"
echo "Try: orb --help"
echo "Get started with: orb init"