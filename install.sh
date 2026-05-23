#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2026 Lucas José de Lima Silva
set -e

echo "Compilando lan-notify..."
go build -o lan-notify main.go

# $PREFIX é uma variável de ambiente nativa e exclusiva do Termux
if [ -n "$PREFIX" ] && [ -d "$PREFIX/bin" ]; then
    echo "Ambiente Termux detectado."
    DEST="$PREFIX/bin/lan-notify"
    mv lan-notify "$DEST"
    chmod +x "$DEST"
    
    # Configura Autocompletar no Termux
    CURRENT_SHELL=$(basename "$SHELL")
    if [ "$CURRENT_SHELL" = "bash" ]; then
        mkdir -p "$PREFIX/etc/bash_completion.d"
        $DEST completion bash > "$PREFIX/etc/bash_completion.d/lan-notify"
        echo "Autocompletar configurado para Bash!"
    elif [ "$CURRENT_SHELL" = "zsh" ]; then
        mkdir -p "$PREFIX/share/zsh/site-functions"
        $DEST completion zsh > "$PREFIX/share/zsh/site-functions/_lan-notify"
        echo "Autocompletar configurado para Zsh!"
    elif [ "$CURRENT_SHELL" = "fish" ]; then
        mkdir -p "$PREFIX/share/fish/vendor_completions.d"
        $DEST completion fish > "$PREFIX/share/fish/vendor_completions.d/lan-notify.fish"
        echo "Autocompletar configurado para Fish!"
    fi
else
    echo "Ambiente Linux padrão detectado."
    DEST="/usr/local/bin/lan-notify"
    echo "Isso requer privilégios de root para gravar em $DEST"
    sudo mv lan-notify "$DEST"
    sudo chmod +x "$DEST"
    
    # Configura Autocompletar no Linux
    CURRENT_SHELL=$(basename "$SHELL")
    if [ "$CURRENT_SHELL" = "bash" ]; then
        sudo mkdir -p /etc/bash_completion.d
        $DEST completion bash | sudo tee /etc/bash_completion.d/lan-notify > /dev/null
        echo "Autocompletar configurado para Bash!"
    elif [ "$CURRENT_SHELL" = "zsh" ]; then
        sudo mkdir -p /usr/local/share/zsh/site-functions
        $DEST completion zsh | sudo tee /usr/local/share/zsh/site-functions/_lan-notify > /dev/null
        echo "Autocompletar configurado para Zsh!"
    elif [ "$CURRENT_SHELL" = "fish" ]; then
        sudo mkdir -p /etc/fish/completions
        $DEST completion fish | sudo tee /etc/fish/completions/lan-notify.fish > /dev/null
        echo "Autocompletar configurado para Fish!"
    fi
fi

echo "Sucesso! 'lan-notify' instalado em $DEST."
echo "Você já pode executar o comando 'lan-notify' de qualquer lugar."
echo "Nota: Reinicie o seu terminal para o autocompletar fazer efeito."
