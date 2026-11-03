# Devcontainer Setup

This project includes a devcontainer configuration for consistent development environments using VS Code's Remote - Containers extension.

## Requirements

- Visual Studio Code
- Docker
- VS Code Remote - Containers extension

## Getting Started

1. Open the project in VS Code
2. When prompted, click "Reopen in Container" or use the command palette: `Dev Containers: Reopen in Container`
3. VS Code will build the container and install all dependencies

## Tooling

The devcontainer uses **mise** to manage development tools:

- **Go**: Latest version (configured in `.mise.toml`)
- **golangci-lint**: For linting Go code

The container adds mise's shims, mise itself, and Go's installed tool binaries to PATH for VS Code terminals, tasks, and extensions. Shims provide tool selection without requiring interactive shell activation.

After changing the devcontainer configuration, run `Dev Containers: Rebuild Container` to apply it.

## Working in the Container

Once the devcontainer is running, you can:

- Build: `go build`
- Run: `go run main.go`
- Test: `go test ./...`
- Lint: `golangci-lint run`

## Updating Tools

To update or add tools, edit `.mise.toml` and run `mise install` in the container terminal.

## Manual Setup (if not using devcontainer)

If you prefer to set up mise locally:

```bash
curl https://mise.run | sh
export PATH="$HOME/.local/share/mise/shims:$HOME/.local/bin:$HOME/go/bin:$PATH"
mise install
go version
```

Add the same PATH export to your shell startup file to persist it across sessions.
