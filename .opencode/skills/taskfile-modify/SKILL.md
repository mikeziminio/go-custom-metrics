---
name: taskfile-modify
description: Skill for modifying Taskfile.yml file following best practices
---

# Taskfile Modify Skill

This skill helps modify Taskfiles following best practices from [Taskfile.dev](https://taskfile.dev/).

## Best Practices Implemented

1. **Logical Grouping**: Organizes tasks into meaningful groups (Build, Run, Test, Lint, DB, etc.)
2. **Consistent Naming**: Uses consistent naming conventions for tasks
3. **Proper Descriptions**: All tasks have clear, descriptive help text
4. **Command Quoting**: Ensures all commands are properly quoted for reliability
5. **Default Task**: Includes a default task that shows available tasks
6. **Improved Readability**: Formats the Taskfile for better maintainability

## Key Features

- Groups related tasks together (Build, Run, Test, Lint, DB)
- Uses consistent naming patterns (e.g., `build-*`, `test-*`, `lint-*`)
- Adds meaningful descriptions to all tasks
- Properly quotes all shell commands
- Includes a default task showing available tasks
- Optimizes task dependencies and execution flow

## Usage

This skill can be used to:
1. Refactor existing Taskfiles to follow best practices
2. Generate new Taskfiles with proper structure and naming conventions
3. Improve readability and maintainability of existing task definitions
