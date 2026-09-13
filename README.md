# DBrain

**DBrain = Development Brain**

DBrain is an authority-aware development brain for long-running, AI-assisted engineering projects.

## Status

- Architecture: BC1–BC9 closed; final architecture gate passed.
- Brain/Knowledge research Lane A+B: centrally synthesized.
- PKG-M0: contract/schema freeze in progress.
- Implementation: not yet authorized beyond bounded scaffolding/contracts.
- Existing platform Notion→Git migration: not authorized.

## Core idea

DBrain is not a document warehouse and not merely an agent-memory/RAG system.

It provides an authority-aware way to resolve and package the knowledge an agent needs for a bounded task while preserving provenance, scope, lifecycle, qualification, and recovery boundaries.

```text
Development Workspace = 1..N repositories

Authority Map
    ↓
Canonical Documents + optional Claims
    ↓
Context / Brain Runtime
    ↓
Bounded Context Bundle
    ↓
Bootstrap → Qualification → Task Admission → Work
                          ↘ Recovery / Resume
```

## Development

This repository currently contains a Go module scaffold:

```bash
go build ./...
go test ./...
```

## PKG-M0

PKG-M0 freezes contracts and schemas before implementation.

See [docs/PKG-M0.md](docs/PKG-M0.md).
