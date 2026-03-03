---
name: go-mocks
description: Comprehensive mocks skill for Go projects, automatic mock generation
---

# Go Mocks Skill

## Purpose
This skill provides mocking functionality for Go projects using Mockery v3, enabling automatic generation of mock implementations for interfaces to streamline testing.

## Dependencies
- go
- mockery v3
- testify (for assertions)

## Mockery config file
Config file is in the root of the project: `.mockery.yml`

## Constraints

**NEVER** use mockery directly, instead use `task mockery` tool.
**NEVER** create new / edit any files with the exception of `.mockery.yml`.
**NEVER** edit *_mock.go files directly, instead use `task mockery` tool.

## Best Practices
- Generate mocks for interfaces, not concrete types
- Use meaningful mock names that match testing needs
- Verify generated mocks before committing
- Update mocks when interfaces change
- Always regenerate mocks after interface updates
