---
name: go-lint
description: Comprehensive linting skill for Go projects with preparation and execution capabilities that integrates with Taskfile.yml
---

# Go Linting Skill

## Overview
Comprehensive linting skill for Go projects with preparation and execution capabilities that integrates with Taskfile.yml.

## Description
This skill provides linting functionality for Go projects, including environment preparation and lint execution. It integrates seamlessly with the Taskfile.yml in the project root and uses the latest stable version of golangci-lint.

## Roles
- senior go backend engineer
- senior devops engineer

## Dependencies
- golangci-lint v1.64.8 (latest stable)
- task

## Integration Points

### Taskfile.yml Integration
Add the following to your root Taskfile.yml:

```yaml
# Linting tasks
lint-prepare:
  desc: Prepare environment for linting
  cmds:
    - echo "Preparing linting environment..."
    - go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
    - echo "Linting environment prepared successfully"

lint:
  desc: Execute linting checks
  cmds:
    - echo "Running linters..."
    - golangci-lint run
    - echo "Linting completed successfully"
```

### Configuration
This skill references `.golangci.yml` for linting configuration. The configuration file should be placed in the project root.
