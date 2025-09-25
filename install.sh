#!/usr/bin/env bash

VERSION="0.0.15"
DOWNLOAD_URL="https://github.com/slugiscool99/pls/raw/refs/heads/main/pls"
INSTALL_DIR="/usr/local/bin"

detect_shell() {
    if [ -n "$ZSH_VERSION" ]; then
        echo "zsh"
    elif [ -n "$BASH_VERSION" ]; then
        echo "bash"
    else
        echo "unknown"
    fi
}

get_rc_file() {
    local shell=$(detect_shell)
    if [ "$shell" = "zsh" ]; then
        echo "$HOME/.zshrc"
    elif [ "$shell" = "bash" ]; then
        echo "$HOME/.bashrc"
    else
        echo ""
    fi
}

if [ "$EUID" -ne 0 ]; then
    echo "This script requires root privileges. Running with sudo..."
    exec sudo bash "$0" "$@"
    exit $?
fi

TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

echo "Downloading pls version ${VERSION}..."
if ! curl -sSL "$DOWNLOAD_URL" -o "pls"; then
    echo "Failed to download pls"
    exit 1
fi

echo "Installing pls to ${INSTALL_DIR}..."
if ! mv "pls" "$INSTALL_DIR"; then
    echo "Failed to move pls to $INSTALL_DIR"
    exit 1
fi

chmod +x "$INSTALL_DIR/pls"

cd - > /dev/null
rm -rf "$TMP_DIR"

RC_FILE=$(get_rc_file)
if [ -n "$RC_FILE" ]; then
    if ! grep -q "$INSTALL_DIR" "$RC_FILE"; then
        echo "export PATH=\$PATH:$INSTALL_DIR" >> "$RC_FILE"
        echo "Added $INSTALL_DIR to PATH in $RC_FILE"
    fi
else
    echo "Couldn't detect shell. Please add $INSTALL_DIR to your PATH manually."
fi

export PATH="$PATH:$INSTALL_DIR"

echo "Installation complete. Run 'pls help' for instructions. If running 'pls' does not work, restart your shell."