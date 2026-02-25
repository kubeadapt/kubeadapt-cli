#!/bin/bash
set -e

REPO="kubeadapt/replace-me"
BINARY_NAME="replace-me"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux|darwin) ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Get latest release tag from GitHub API
LATEST=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
    echo "Error: Could not determine latest release. Check that the repository has published releases."
    exit 1
fi

VERSION="${LATEST#v}"
FILENAME="${BINARY_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST/$FILENAME"
CHECKSUMS_URL="https://github.com/$REPO/releases/download/$LATEST/checksums.txt"

echo "Downloading $BINARY_NAME $LATEST for $OS/$ARCH..."
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

curl -sL "$URL" -o "$TMP_DIR/$FILENAME"
curl -sL "$CHECKSUMS_URL" -o "$TMP_DIR/checksums.txt"

echo "Verifying checksum..."
cd "$TMP_DIR"
grep "$FILENAME" checksums.txt | sha256sum -c -
cd - > /dev/null

echo "Extracting..."
tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"

echo "Installing to $INSTALL_DIR..."
sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo "$BINARY_NAME $LATEST installed successfully!"
echo "Run '$BINARY_NAME version' to verify."
