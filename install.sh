#!/bin/sh
set -e

# CCSL installer script
# Usage: curl -sSfL https://raw.githubusercontent.com/ming-claude/ccsl/main/install.sh | sh

REPO="ming-claude/ccsl"
INSTALL_DIR="${CCSL_INSTALL_DIR:-$HOME/.local/bin}"

# Detect OS
OS="$(uname -s)"
case "$OS" in
    Darwin) OS="darwin" ;;
    Linux)  OS="linux" ;;
    *)
        echo "Error: unsupported OS: $OS" >&2
        exit 1
        ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)
        echo "Error: unsupported architecture: $ARCH" >&2
        exit 1
        ;;
esac

# Get latest version
VERSION="$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')"
if [ -z "$VERSION" ]; then
    echo "Error: failed to get latest version" >&2
    exit 1
fi

ARCHIVE="ccsl_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/v${VERSION}/${ARCHIVE}"

echo "Installing ccsl v${VERSION} (${OS}/${ARCH})..."

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download and extract
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -sSfL "$URL" -o "$TMPDIR/$ARCHIVE"
tar -xzf "$TMPDIR/$ARCHIVE" -C "$TMPDIR"
install -m 755 "$TMPDIR/ccsl" "$INSTALL_DIR/ccsl"

echo "Installed ccsl to ${INSTALL_DIR}/ccsl"

# Check if install dir is in PATH
case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
        echo ""
        echo "Note: ${INSTALL_DIR} is not in your PATH."
        echo "Add it with: export PATH=\"${INSTALL_DIR}:\$PATH\""
        ;;
esac

echo ""
echo "To enable ccsl in Claude Code, add to ~/.claude/settings.json:"
echo '  "statusLine": { "type": "command", "command": "ccsl" }'
