#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2026 Lucas José de Lima Silva
set -e

echo "Compiling lan-notify..."
go build -o lan-notify main.go

# $PREFIX is a native environment variable exclusive to Termux
if [ -n "$PREFIX" ] && [ -d "$PREFIX/bin" ]; then
    echo "Termux environment detected."
    DEST="$PREFIX/bin/lan-notify"
    mv lan-notify "$DEST"
    chmod +x "$DEST"
    
    # Configure Autocomplete on Termux
    CURRENT_SHELL=$(basename "$SHELL")
    if [ "$CURRENT_SHELL" = "bash" ]; then
        mkdir -p "$PREFIX/etc/bash_completion.d"
        $DEST completion bash > "$PREFIX/etc/bash_completion.d/lan-notify"
        echo "Autocomplete configured for Bash!"
    elif [ "$CURRENT_SHELL" = "zsh" ]; then
        mkdir -p "$PREFIX/share/zsh/site-functions"
        $DEST completion zsh > "$PREFIX/share/zsh/site-functions/_lan-notify"
        echo "Autocomplete configured for Zsh!"
    elif [ "$CURRENT_SHELL" = "fish" ]; then
        mkdir -p "$PREFIX/share/fish/vendor_completions.d"
        $DEST completion fish > "$PREFIX/share/fish/vendor_completions.d/lan-notify.fish"
        echo "Autocomplete configured for Fish!"
    fi
else
    echo "Standard Linux environment detected."
    DEST="/usr/local/bin/lan-notify"
    echo "This requires root privileges to write to $DEST"
    sudo mv lan-notify "$DEST"
    sudo chmod +x "$DEST"
    
    # Configure Autocomplete on Linux
    CURRENT_SHELL=$(basename "$SHELL")
    if [ "$CURRENT_SHELL" = "bash" ]; then
        sudo mkdir -p /etc/bash_completion.d
        $DEST completion bash | sudo tee /etc/bash_completion.d/lan-notify > /dev/null
        echo "Autocomplete configured for Bash!"
    elif [ "$CURRENT_SHELL" = "zsh" ]; then
        sudo mkdir -p /usr/local/share/zsh/site-functions
        $DEST completion zsh | sudo tee /usr/local/share/zsh/site-functions/_lan-notify > /dev/null
        echo "Autocomplete configured for Zsh!"
    elif [ "$CURRENT_SHELL" = "fish" ]; then
        sudo mkdir -p /etc/fish/completions
        $DEST completion fish | sudo tee /etc/fish/completions/lan-notify.fish > /dev/null
        echo "Autocomplete configured for Fish!"
    fi
fi

echo "Success! 'lan-notify' installed at $DEST."
echo "You can now run the 'lan-notify' command from anywhere."
echo "Note: Restart your terminal for autocomplete to take effect."
