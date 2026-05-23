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
else
    echo "Ambiente Linux padrão detectado."
    DEST="/usr/local/bin/lan-notify"
    echo "Isso requer privilégios de root para gravar em $DEST"
    sudo mv lan-notify "$DEST"
    sudo chmod +x "$DEST"
fi

echo "Sucesso! 'lan-notify' instalado em $DEST."
echo "Você já pode executar o comando 'lan-notify' de qualquer lugar."
