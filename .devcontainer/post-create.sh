#!/bin/bash
set -e

echo "Installing mise..."
curl https://mise.run | sh
export PATH="$HOME/.local/share/mise/shims:$HOME/.local/bin:$HOME/go/bin:$PATH"

echo "Initializing mise and installing Go..."
cd "$(dirname "$0")/.."
mise install

echo "Verifying Go installation..."
go version

echo "Installing Go development tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

echo "Devcontainer setup complete!"
