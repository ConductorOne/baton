# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Baton is a CLI tool for exploring and analyzing identity governance data (users, permissions, roles, groups, resources) stored in `.c1z` files. It includes a web-based explorer UI built with React/TypeScript, embedded into the Go binary.

## Build & Development Commands

```bash
# Build the full binary (builds frontend first, then Go)
make build

# Build only the frontend (React app)
make frontend

# Run linter (golangci-lint v2)
make lint

# Run all tests
go test -v ./...

# Run a single test
go test -v -run TestName ./path/to/package

# Update dependencies (updates, tidies, and vendors)
make update-deps

# Generate protobuf code
make protogen
```

## Architecture

### CLI Layer (`cmd/baton/`)
Each subcommand is a separate file using Cobra. Commands load a `.c1z` file via the `-f` flag (default: `sync.c1z`) and format output through a pluggable Manager interface supporting console and JSON formats.

Key commands: `resources`, `entitlements`, `grants`, `access`, `principals`, `diff`, `export`, `stats`, `explorer`.

### Core Packages (`pkg/`)
- **`pkg/storecache/`** — Thread-safe in-memory cache (sync.Map) for resources, resource types, entitlements, and grants. Used across commands to avoid repeated DB lookups. Returns "MISSING" placeholders for unfound data.
- **`pkg/output/`** — Output Manager interface with console and JSON implementations for marshaling protobuf messages.
- **`pkg/explorer/`** — Gin-based web server that serves the embedded React frontend and exposes REST handlers for querying c1z data. Auto-opens browser on startup.

### Frontend (`frontend/`)
React 18 + TypeScript app using Material-UI, ReactFlow for graph visualization, and React Router. Built artifacts are copied into `pkg/explorer/frontend/` and embedded into the Go binary via `embed.FS`.

### Protobuf (`proto/` → `pb/`)
Output message types defined in `proto/baton/v1/outputs.proto`. Generated with Buf into `pb/`. CI validates lint and breaking changes.

## Dependencies & Tooling

- **Go 1.25.2** with vendored dependencies (`vendor/`)
- **Nix flake** for dev environment (`.envrc` + `flake.nix`)
- **golangci-lint v2** with 26+ linters; line length limit of 200 chars
- **Buf** for protobuf generation and linting
