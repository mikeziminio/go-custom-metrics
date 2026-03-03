---
name: opencode-skill
description: Skill for creating and managing OpenCode skills
---

# OpenCode Skill Creation Guide

⚠️ **IMPORTANT**: Every skill file MUST start with YAML frontmatter containing `name` and `description` fields. NEVER FORGET THIS REQUIREMENT!

This guide provides comprehensive instructions for creating and managing OpenCode skills. Skills are reusable instruction sets that can be loaded on-demand by agents.

## File Location
Create one folder per skill name and put a `SKILL.md` inside it:
`.opencode/skills/<name>/SKILL.md`

## Frontmatter Requirements

Each `SKILL.md` must start with YAML frontmatter containing these required fields:

### Required Fields
- `name` (required): Unique identifier for the skill
- `description` (required): Brief description of what the skill does

## Naming Rules

Skill names must:
- Be 1–64 characters
- Be lowercase alphanumeric with single hyphen separators
- Not start or end with "-"
- Not contain consecutive "--"
- Match the directory name that contains `SKILL.md`

Equivalent regex: `^[a-z0-9]+(-[a-z0-9]+)*$`

## Writing Effective Skill Descriptions

Descriptions must be 1-1024 characters and clearly explain what the skill does. Keep them specific enough for agents to choose correctly.

## Example Skill Structure

```
---
name: git-release
description: Create consistent releases and changelogs
---

## What I do
- Draft release notes from merged PRs
- Propose a version bump
- Provide a copy-pasteable `gh release create` command

## When to use me
Use this when you are preparing a tagged release.
Ask clarifying questions if the target versioning scheme is unclear.
```

## Skill Permissions

Control which skills agents can access using pattern-based permissions in `opencode.json`:

```
{
  "permission": {
    "skill": {
      "*": "allow",
      "pr-review": "allow",
      "internal-*": "deny",
      "experimental-*": "ask"
    }
  }
}
```

## Best Practices

1. **Clear Descriptions**: Write specific, actionable descriptions
2. **Consistent Formatting**: Use consistent Markdown formatting
3. **Appropriate Permissions**: Set proper access controls for security
4. **Meaningful Names**: Use descriptive names that reflect skill purpose
5. **Comprehensive Documentation**: Explain when and how to use the skill
6. **Version Control**: Keep skill definitions under version control
7. **Testing**: Test skills with different agent scenarios

## Troubleshooting

If a skill does not show up:
1. Verify SKILL.md is spelled in all caps
2. Check that frontmatter includes `name` and `description`
3. Ensure skill names are unique across all locations
4. Check permissions—skills with `deny` are hidden from agents
