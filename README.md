# Package Manager (PM)

A Go-based package manager for creating, uploading, and downloading packages via SSH.

## Build

```bash
go build -o bin/pm ./cmd/pm
```

## Usage

### Create Package

Create `packet.json`:
```json
{
  "name": "my-package",
  "ver": "1.0.0",
  "targets": [
    "src/*.go",
    {
      "path": "docs/*",
      "exclude": ["*.tmp", "*.log"]
    }
  ]
}
```

Create `ssh-config.json`:
```json
{
  "host": "localhost",
  "port": 22,
  "username": "user",
  "key_path": "~/.ssh/id_rsa",
  "remote_dir": "/var/packages"
}
```

Upload package:
```bash
./bin/pm create packet.json -c ssh-config.json
```

### Install Packages

Create `packages.json`:
```json
{
  "packages": [
    {"name": "my-package", "ver": ">=1.0.0"},
    {"name": "other-package"}
  ]
}
```

Download packages:
```bash
./bin/pm update packages.json -c ssh-config.json
```

## Commands

- `pm create <packet.json>` - Create and upload package
- `pm update <packages.json>` - Download and install packages
- `pm version` - Show version

## File Patterns

- `*.go` - All Go files
- `**/*.go` - All Go files recursively
- `src/**/*` - All files in src directory

## Tests

```bash
# Requires SSH connection
# runs both unit and e2e tests
./scripts/test.sh
```
