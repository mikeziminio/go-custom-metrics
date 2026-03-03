---
description: Commit all changes
agent: git-commiter
---

Check current git diff.
Based on these results, suggest one-line commit message.
Use a semicolon as a separator if necessary.

Print `git status` result to the user.

Run commands to commit changes:
```bash
git add {modified and new file list}
git commit -m "{commit message}"
```
