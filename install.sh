#!/usr/bin/env bash
#
# Installs pls to /usr/local/bin.
#
# Works both from a checkout (./install.sh, uses the local ./pls binary) and
# piped from the web (curl -sSL .../install.sh | sudo bash, downloads it).

set -euo pipefail

INSTALL_DIR="/usr/local/bin"
BINARY_URL="https://github.com/slugiscool99/pls/raw/refs/heads/main/pls"

if [ "$(id -u)" -ne 0 ]; then
    echo "This script requires root privileges. Re-run with sudo:"
    echo "  curl -sSL https://github.com/slugiscool99/pls/raw/refs/heads/main/install.sh | sudo bash"
    exit 1
fi

if [ "$(uname -s)" != "Darwin" ] || [ "$(uname -m)" != "arm64" ]; then
    echo "The prebuilt pls binary only supports Apple Silicon macOS."
    echo "On other platforms, clone the repo and run 'go build'."
    exit 1
fi

# Use a local binary when run from a checkout; otherwise download one.
# (When piped into bash, BASH_SOURCE is empty, so there is no local binary.)
SCRIPT_DIR=""
if [ -n "${BASH_SOURCE[0]:-}" ] && [ -f "${BASH_SOURCE[0]}" ]; then
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

if [ -n "$SCRIPT_DIR" ] && [ -f "$SCRIPT_DIR/pls" ]; then
    BINARY_PATH="$SCRIPT_DIR/pls"
else
    echo "Downloading pls..."
    BINARY_PATH="$TMP_DIR/pls"
    if ! curl -fsSL "$BINARY_URL" -o "$BINARY_PATH"; then
        echo "Failed to download pls from $BINARY_URL"
        exit 1
    fi
fi

# A previous buggy install could leave the binary itself at /usr/local/bin.
if [ -e "$INSTALL_DIR" ] && [ ! -d "$INSTALL_DIR" ]; then
    echo "Found a file at $INSTALL_DIR from a broken earlier install. Repairing..."
    rm -f "$INSTALL_DIR"
fi

# /usr/local/bin does not exist on fresh Apple Silicon Macs.
mkdir -p "$INSTALL_DIR"

echo "Installing pls to $INSTALL_DIR/pls..."
install -m 755 "$BINARY_PATH" "$INSTALL_DIR/pls"

# Drop the quarantine flag so Gatekeeper doesn't block the downloaded binary.
xattr -d com.apple.quarantine "$INSTALL_DIR/pls" 2>/dev/null || true

echo "Installation complete. Run 'pls help' for instructions."
echo "If 'pls' isn't found, open a new terminal (or run 'rehash' in zsh)."
