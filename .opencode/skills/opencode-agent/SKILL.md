---
name: opencode-agent
description: Skill for creating and managing new opencode agents
---

# OpenCode Agent Creation Skill

This skill provides guidance and templates for creating new OpenCode agents. Agents are specialized AI assistants that can be used for various development tasks within the OpenCode environment.

## Agent Types

OpenCode supports two main agent modes:

### Primary Agents
- Used for main development workflows
- Typically invoked directly by users
- Have broader capabilities and permissions

### Subagent Agents
- Specialized for specific tasks
- Invoked by primary agents or through the Task tool
- More focused capabilities with restricted permissions

## Agent Configuration Structure

Each agent is defined in a Markdown file with metadata at the top:

```markdown
---
description: Brief description of agent purpose
mode: primary|subagent
model: provider/model-name
temperature: 0.3
tools:
  write: true
  edit: false
  bash: false
permission:
  edit: deny
  bash:
    "*": ask
    "git status": allow
---
System prompt content here...
```

## Metadata Fields

### description
A brief description of what the agent does

### mode
Either `primary` or `subagent` indicating the agent type

### model
The LLM model to use for this agent (e.g., `sber-qwen-coder-30b-dev/qwen3-coder-30b-dev`)

### temperature
Controls randomness (0.0-1.0):
- Low (0.1-0.2): For analysis and planning
- Medium (0.3-0.5): For balanced tasks
- High (0.6-1.0): For creativity and exploration

### tools
Permissions for tools that the agent can use:
- write: Can create new files
- edit: Can modify existing files
- bash: Can execute shell commands

### permission
Fine-grained control over specific tools:
- `deny`: Completely disallow access
- `ask`: Require user confirmation before executing
- `allow`: Allow unrestricted access

## Creating New Agents

### Step 1: Choose Agent Type
Determine if your agent should be primary or subagent based on its role.

### Step 2: Define Purpose
Create a clear description of what your agent will do.

### Step 3: Select Model
Choose an appropriate model based on the complexity of tasks.

### Step 4: Set Temperature
Adjust temperature according to the agent's purpose:
- Planning and analysis: 0.1-0.2
- Balanced development: 0.3-0.5
- Creative tasks: 0.6-0.8

### Step 5: Configure Tools
Set appropriate tool permissions based on security and functionality needs.

### Step 6: Write System Prompt
Create the system prompt that defines how the agent behaves and what it should accomplish.

## Example Agent Templates

### Basic Agent Template
```markdown
---
description: Brief description of agent purpose
mode: primary|subagent
model: sber-qwen-coder-30b-dev/qwen3-coder-30b-dev
temperature: 0.3
tools:
  write: true
  edit: true
  bash: true
permission:
  edit: deny
  bash:
    "*": ask
    "git status": allow
---
System prompt content here...
```

### Security Agent Template
```markdown
---
description: Performs security audits and identifies vulnerabilities
mode: subagent
tools:
  write: false
  edit: false
permission:
  write: deny
  edit: deny
---
You are a security expert. Focus on identifying potential security issues.
```

### Documentation Agent Template
```markdown
---
description: Writes and maintains project documentation
mode: subagent
tools:
  bash: false
permission:
  bash: deny
---
You are a technical writer. Create clear, comprehensive documentation.
```

## Best Practices

1. **Security First**: Always restrict unnecessary permissions
2. **Purpose Clarity**: Each agent should have a well-defined role
3. **Model Consistency**: Use the same model across related agents for consistent performance and behavior
4. **Temperature Tuning**: Adjust based on desired output characteristics
5. **Permission Granularity**: Use specific permission controls for sensitive operations
6. **Documentation**: Include clear descriptions and usage examples
7. **Testing**: Test agents with sample inputs before deploying
8. **Version Control**: Keep agent configurations under version control

## Integration Points

New agents can be integrated into the OpenCode system by placing them in the `.opencode/agents/` directory. The system will automatically discover and load agents from this location.

## Agent Lifecycle Management

1. **Creation**: Agents are defined as Markdown files with proper metadata
2. **Activation**: Agents are loaded when referenced in prompts or workflows
3. **Execution**: Agents run with their configured permissions and tools
4. **Monitoring**: Track agent performance and usage through logs
5. **Maintenance**: Update agent definitions as requirements evolve

## Advanced Features

### Tool Permissions
Agents can have granular control over tool access:
- `allow`: Full access to tool
- `deny`: No access to tool
- `ask`: User must confirm before tool execution

### Model Selection
Choose models based on:
- Task complexity and required intelligence
- Cost considerations
- Response speed requirements
- Availability of specific model capabilities

### Context Awareness
Agents can be designed to:
- Remember previous interactions within a session
- Understand project context
- Maintain state across multiple interactions
