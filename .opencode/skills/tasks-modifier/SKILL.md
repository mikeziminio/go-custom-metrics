---
name: tasks-modifier
description: Skill for modifying tasks.md file following best practices
---

# Tasks Modifier Skill

This skill enables working with project tasks using the https://tasks.md/ format, specifically modifying tasks in the `.opencode/tasks.md` file.

## What I do
- Modify existing tasks in the tasks.md file
- Follow best practices for task management
- Use proper formatting and structure as per tasks.md standards
- Ensure consistency when updating task items
- Handle task status updates and modifications properly

## Best Practices
1. **Task Formatting**: Follow the tasks.md standard with proper markdown list syntax
2. **Status Indicators**: Use correct task status markers ([ ], [x], [!] for pending, completed, in_progress)
3. **Priority Levels**: Include priority indicators when needed (high, medium, low)
4. **Consistent Structure**: Maintain consistent formatting across all tasks
5. **Clear Descriptions**: Provide meaningful task descriptions and notes
6. **Atomic Updates**: Make single, focused changes to tasks

## File Path
All tasks are stored in the hardcoded file path: `.opencode/tasks.md`

## Tasks.md Format Specification
Tasks are stored using the standard tasks.md format:
- Each task is a markdown list item
- Tasks can have priority indicators (high, medium, low)
- Tasks can have status indicators (pending, in_progress, completed, cancelled)
- Tasks can include brief descriptions and notes

## Example Usage
```
- [ ] High priority task with description
- [x] Completed task
- [ ] Medium priority task with notes
```

## Modification Commands
- Update task status
- Modify task descriptions
- Add or remove task notes
- Change task priorities
- Move tasks between statuses