# Legacy Docker Files

These files are kept for reference only. They are **not maintained**.

## What was here
- 6 docker-compose files (MySQL + PostgreSQL variants with overlay chaining)
- 5 docker helper scripts (build/run in .sh and .bat, cleanup)

## Why they were replaced
- Go version mismatch (1.21 vs 1.23 in go.mod)
- MySQL configs were dead code (project uses PostgreSQL)
- Required chaining 2-3 compose files with `-f` flags
- Network conflicts between overlay files
- Massive env var duplication

## Current setup
See `docker/docker-compose.dev.yml` and `docker/docker-compose.prod.yml` at the parent directory.
