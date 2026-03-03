---
name: go-test-run
description: Comprehensive skill for running unit and integration tests in Go
---

# Go Testing Skill

## Overview
Comprehensive skill for running unit and integration tests in Go.

## Constraints

- **NEVER** run `go test` when it is required to run **ALL** tests.
  **ALWAYS** use `task test` instead.

- **ALWAYS** use `go test -race {path} -run {TestMethod}` when it is requested
  to run specific test.
