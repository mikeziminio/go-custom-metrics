---
description: Responsible for git commit
mode: subagent
model: sber-qwen-coder-30b-dev/qwen3-coder-30b-dev
temperature: 0.1
permission:
  bash:
    "*": deny
    "git status": allow
    "git log *": allow
    "git diff *": allow
    "git add *": ask
    "git commit *": ask
  edit:  # edit, write, patch, multiedit
    "*": deny
  read: allow
  grep: allow
  glob: allow
  list: allow
  lsp: allow
  todowrite: allow
  todoread: allow
  webfetch: allow
  websearch: allow
  question: allow
  make_mocks: allow
---

# Git Commiter Agent
