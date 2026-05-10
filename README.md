# EndGit CLI
**Command-line tool for the Endstone plugin ecosystem**   
EndGit CLI allows developers to publish, search, and install Endstone plugins directly from the terminal.

[![Release](https://github.com/two-tech-dev/endgit-cli/actions/workflows/release.yml/badge.svg)](https://github.com/two-tech-dev/endgit-cli/actions/workflows/release.yml)   

## Installation

**Windows:**
```powershell
irm https://endgit.dev/installer.ps1 | iex
```

**Linux / macOS:**
```bash
curl -sSL https://endgit.dev/installer.sh | bash
```


## Usage

```bash
# Login with your EndGit account
endgit login

# Search plugins
endgit search <query>

# View plugin details
endgit info <plugin>

# Install a plugin (latest version)
endgit install <plugin>

# Install a specific version
endgit install <plugin>@1.2.0

# Install a dev build by commit
endgit install <plugin>@abc1234

# Initialize a new plugin project
endgit init

# Update EndGit to the latest version
endgit update
```

## Commands

| Command | Description |
|---------|-------------|
| `endgit login` | Authenticate with EndGit |
| `endgit logout` | Logout of EndGit |
| `endgit search <query>` | Search the plugin marketplace |
| `endgit info <plugin>` | Show detailed plugin information |
| `endgit install <plugin[@version]>` | Install a plugin to your server |
| `endgit init` | Initialize a new plugin project |
| `endgit publish` | Package and publish your plugin |
| `endgit update` | Update EndGit CLI to the latest version |

## Building

### Requirements

- Go 1.26+ (see `go.mod` for exact version)
- `GOPATH/bin` added to your system `PATH`

### Build from Source

```bash
git clone https://github.com/two-tech-dev/endgit-cli.git
cd endgit-cli

# Development build
go build -o endgit .

# Production build (with version injection)
go build -ldflags="-s -w -X github.com/two-tech-dev/endgit-cli/cmd.Version=v0.1.5" -o endgit .
```

### Run Locally

```bash
go run .
```

## License

This project is licensed under the [GNU Affero General Public License v3.0](LICENSE).
