# SSH Tunnel Manager

[中文](./README_zh.md) | English

A web-based GUI tool for managing SSH tunnels, built with Go.

## Features

- **Web UI** - Clean web interface for managing tunnels
- **Local & Remote Tunnels** - Support for both `-L` (local) and `-R` (remote) SSH tunnel types
- **RESTful API** - Full API for tunnel CRUD operations
- **SSH Authentication** - Support for both SSH key and password authentication
- **Auto-recovery** - Automatic retry mechanism when tunnels fail to start
- **Process Persistence** - Survives application restarts, restores tunnel states
- **Cross-platform** - Runs on Windows, macOS, and Linux

## Quick Start

### Build

```bash
go build -o ssh_tunnel_manager .
```

### Run

```bash
./ssh_tunnel_manager
```

The application will:
1. Start HTTP server on port `11108`
2. Open browser automatically to `http://localhost:11108`
3. Load existing tunnels from `config.json`

## Configuration

Tunnels are stored in `config.json`:

```json
{
  "tunnels": [
    {
      "name": "dev_db",
      "type": "local",
      "local_port": "33306",
      "remote_host": "localhost",
      "remote_port": "13306",
      "ssh_host": "155.248.183.71",
      "ssh_port": "22",
      "ssh_user": "ubuntu",
      "ssh_key": "/path/to/key.pem",
      "ssh_pass": ""
    }
  ]
}
```

### Fields

| Field | Description | Required |
|-------|-------------|----------|
| `name` | Unique tunnel identifier | Yes |
| `type` | Tunnel type: `local` or `remote` | Yes |
| `local_port` | Local port for the tunnel | Yes |
| `remote_host` | Remote host to tunnel to | Yes |
| `remote_port` | Remote port to tunnel to | Yes |
| `ssh_host` | SSH server hostname/IP | Yes |
| `ssh_port` | SSH server port (default: 22) | No |
| `ssh_user` | SSH username | Yes |
| `ssh_key` | Path to SSH private key | Yes* |
| `ssh_pass` | SSH password | Yes* |

*Either `ssh_key` or `ssh_pass` must be provided.

## API Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/tunnels` | List all tunnels |
| POST | `/api/tunnels` | Create a new tunnel |
| GET | `/api/tunnels/{id}` | Get tunnel details |
| PUT | `/api/tunnels/{id}` | Update tunnel |
| DELETE | `/api/tunnels/{id}` | Delete tunnel |
| POST | `/api/tunnels/{id}/start` | Start tunnel |
| POST | `/api/tunnels/{id}/stop` | Stop tunnel |
| GET | `/api/tunnels/{id}/status` | Get tunnel status |
| GET | `/api/tunnels/{id}/stats` | Get tunnel statistics |
| GET | `/api/ping` | Health check |

## Tunnel Types

### Local Tunnel (`-L`)

Forward traffic from local port to remote host via SSH server:

```
localhost:local_port -> SSH_SERVER -> remote_host:remote_port
```

### Remote Tunnel (`-R`)

Forward traffic from remote port to local machine:

```
SSH_SERVER:remote_port -> localhost:local_port
```

## Dependencies

- [gorilla/mux](https://github.com/gorilla/mux) - HTTP routing

## License

MIT
