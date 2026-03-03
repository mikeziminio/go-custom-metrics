---
description: Responsible for Go-specific implementation and compliance
mode: subagent
model: sber-qwen-coder-30b-dev/qwen3-coder-30b-dev
temperature: 0.1
permission:
  bash:
    "*": deny
  edit:  # edit, write, patch, multiedit
    "*": deny
    "asyncapi.yaml": allow
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
  skill:
    "asyncapi-modifier": allow
---

## Skills

- **MUST** use "asyncapi-rmq" skill when it is required to write / rewrite / validate asyncapi spec.
  **NEVER** write / rewrite / validate asyncapi spec directly, without "asyncapi-rmq" skill.
