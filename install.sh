#!/bin/sh
set -e

REPO="halilbulentorhon/pjf"
INSTALL_DIR="/usr/local/bin"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  darwin) ;;
  *) echo "Unsupported OS: $OS (macOS only for now)"; exit 1 ;;
esac

LATEST=$(curl -sSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)

if [ -z "$LATEST" ]; then
  echo "Failed to fetch latest release"
  exit 1
fi

URL="https://github.com/$REPO/releases/download/$LATEST/pjf_${OS}_${ARCH}.tar.gz"

echo "Installing pjf $LATEST ($OS/$ARCH)..."

TMP=$(mktemp -d)
curl -sSL "$URL" | tar xz -C "$TMP"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/pjf" "$INSTALL_DIR/pjf"
else
  sudo mv "$TMP/pjf" "$INSTALL_DIR/pjf"
fi

rm -rf "$TMP"

echo "pjf $LATEST installed to $INSTALL_DIR/pjf"
echo "Run 'pjf' to get started."
