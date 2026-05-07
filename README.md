# EndGit CLI

**Command-line tool for the Endstone plugin ecosystem**

EndGit CLI allows developers to publish, search, and install Endstone plugins directly from the terminal.

## Installation

**Windows:**
```powershell
irm https://endgit.dev/installer.ps1 | iex
```

**Linux:**
```bash
curl -sSL https://endgit.dev/installer.sh | bash
```


## Usage

```bash
# Login with your EndGit account
endgit login

# Publish a plugin
endgit publish

# Search plugins
endgit search <query>

# Install a plugin
endgit install <plugin-name>
```

## Commands

| Command | Description |
|---------|-------------|
| `endgit login` | Authenticate with EndGit |
| `endgit logout` | Logout of EndGit |
| `endgit publish` | Package and publish your plugin |
| `endgit search` | Search the plugin marketplace |
| `endgit install` | Install a plugin to your server |

## License

This project is licensed under the [GNU Affero General Public License v3.0](LICENSE).
