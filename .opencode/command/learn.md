---
description: Extract non-obvious learnings from session to agents and skills files to build codebase understanding
agent: plan
---

Analyze this session and extract non-obvious learnings that could be added to opencode agents and skill files.

$ARGUMENTS

What counts as a learning (non-obvious discoveries only):

- New knowledge about how agents and skills can perform tasks better
- Suggestions for limiting agents and skills
- Hidden relationships between files or modules
- Execution paths that differ from how code appears
- Non-obvious configuration, env vars, or flags
- Debugging breakthroughs when error messages were misleading
- API/tool quirks and workarounds
- Build/test commands not in README
- Architectural decisions and constraints
- Files that must change together

What NOT to include:

- Obvious facts from documentation
- Standard language/framework behavior
- Verbose explanations
- Session-specific details

Process:

- Review session for discoveries, errors that took multiple attempts, unexpected connections
- Determine scope - what agent / skill does each learning apply to?
- Read existing agent / skill files
- Suggest extending agent / skill files OR suggest create new agents / skills.
- Keep entries to 1-3 lines per insight
- Don't change agents / skill files. Just suggest changes in structured way: 
  The list of:
  - Summarized new knowledge
  - Agent / skill
