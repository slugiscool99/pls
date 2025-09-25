#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_PATH="$SCRIPT_DIR/pls"
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

if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: pls binary not found at $BINARY_PATH"
    exit 1
fi

if [ "$EUID" -ne 0 ]; then
    echo "This script requires root privileges. Running with sudo..."
    if sudo bash "$0" "$@"; then
        # After successful installation, source the RC file in user's shell
        RC_FILE=$(get_rc_file)
        if [ -n "$RC_FILE" ] && [ -f "$RC_FILE" ]; then
            echo "Reloading shell configuration..."
            # Source the RC file in a subshell to avoid affecting the current script
            (source "$RC_FILE")
        fi
        echo "Installation complete! You can now use 'pls' immediately."
        echo "Run 'pls help' for instructions."
    fi
    exit $?
fi

# Root installation part
echo "Copying pls to ${INSTALL_DIR}..."
if ! cp "$BINARY_PATH" "$INSTALL_DIR/pls"; then
    echo "Failed to copy pls to $INSTALL_DIR"
    exit 1
fi

chmod +x "$INSTALL_DIR/pls"

RC_FILE=$(get_rc_file)
if [ -n "$RC_FILE" ]; then
    if ! grep -q "$INSTALL_DIR" "$RC_FILE"; then
        echo "export PATH=\$PATH:$INSTALL_DIR" >> "$RC_FILE"
        echo "Added $INSTALL_DIR to PATH in $RC_FILE"
    fi
else
    echo "Couldn't detect shell. Please add $INSTALL_DIR to your PATH manually."
fi

echo "Installation completed successfully."