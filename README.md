# portwatch

Lightweight CLI daemon that monitors open TCP ports and alerts on unexpected changes.

## Features

- Scans configurable port ranges at a set interval
- Detects newly opened and recently closed ports
- Filters out known/expected ports via an ignore list
- Persists state between runs so restarts don't produce false positives
- Alerts via stdout, desktop notifications, and/or webhooks

## Installation

```bash
go install github.com/yourname/portwatch/cmd/portwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/portwatch.git
cd portwatch
go build -o portwatch ./cmd/portwatch
```

## Usage

```
portwatch [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-config` | `portwatch.json` | Path to config file |
| `-webhook` | `` | Webhook URL to POST alerts to |
| `-desktop` | `false` | Enable desktop notifications |
| `-once` | `false` | Run a single scan and exit |

### Example

```bash
# Run with defaults (reads portwatch.json if present)
portwatch

# Run once and print results
portwatch -once

# Send alerts to a Slack-compatible webhook
portwatch -webhook https://hooks.slack.com/services/XXX/YYY/ZZZ

# Enable desktop notifications (macOS / Linux with libnotify)
portwatch -desktop
```

## Configuration

Copy the example config and edit as needed:

```bash
cp portwatch.example.json portwatch.json
```

```json
{
  "interval": 60,
  "port_range": [1, 65535],
  "ignore_ports": [22, 80, 443]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `interval` | int | Seconds between scans |
| `port_range` | [min, max] | Inclusive port range to scan |
| `ignore_ports` | []int | Ports to silently ignore |

Defaults are used for any missing fields.

## State

portwatch stores a snapshot of the last known open ports in `~/.portwatch/state.json`. This ensures that restarting the daemon does not re-alert on ports that were already open.

To reset state and re-baseline:

```bash
rm ~/.portwatch/state.json
```

## Alert Format

Alerts are written to stdout in a structured, human-readable format:

```
[ALERT] 2024-01-15T10:23:01Z  port 8080 is now OPEN
[WARN]  2024-01-15T10:23:01Z  port 3000 is now CLOSED
```

Webhook payloads are JSON:

```json
{"event": "opened", "port": 8080, "timestamp": "2024-01-15T10:23:01Z"}
```

## Development

```bash
# Run tests
go test ./...

# Run with race detector
go test -race ./...

# Lint
golangci-lint run
```

## License

MIT
