# Claude Code Feature Usage Guide

> Last Updated: October 31, 2025
>
> This guide defines when and how to use Sub Agents, Slash Commands, MCP Servers, and Skills in Claude Code to maximize development efficiency.

---

## Table of Contents

1. [Overview](#overview)
2. [Sub Agents](#sub-agents)
3. [Slash Commands](#slash-commands)
4. [MCP Servers](#mcp-servers)
5. [Skills](#skills)
6. [Decision Matrix](#decision-matrix)
7. [Best Practices](#best-practices)

---

## Overview

Claude Code provides four primary extensibility features, each serving distinct purposes:

| Feature | Purpose | Context | Invocation |
|---------|---------|---------|------------|
| **Sub Agents** | Specialized task execution with isolated context | Separate per agent | Automatic or manual |
| **Slash Commands** | Reusable prompts and workflows | Main conversation | Manual by user |
| **MCP Servers** | External tool and data integrations | Available to all | Automatic tool use |
| **Skills** | Domain knowledge and capabilities | Auto-loaded when relevant | Automatic by Claude |

---

## Sub Agents

### What Are Sub Agents?

Sub agents are specialized AI assistants with their own:
- **Custom system prompts** that guide behavior
- **Isolated context window** (prevents main conversation pollution)
- **Specific tool access** tailored to their purpose
- **Task-specific configuration**

### When to Use Sub Agents

✅ **USE SUB AGENTS FOR:**

1. **Complex, Multi-Step Tasks**
   - Code reviews requiring detailed analysis
   - Refactoring large codebases
   - Test generation and verification
   - Documentation generation

2. **Specialized Expertise**
   - Security auditing
   - Performance optimization
   - API design and review
   - Database schema design

3. **Long-Running Operations**
   - Migration scripts
   - Batch processing
   - Comprehensive testing suites
   - Large-scale refactoring

4. **Context Isolation Needs**
   - Tasks that would clutter main conversation
   - Parallel work streams
   - Independent analysis

### How to Create Sub Agents

Sub agents are defined in `.claude/agents/[name]/AGENT.md`:

```markdown
---
name: agent-name
description: Brief description of what this agent does
tools: [Read, Edit, Bash, Grep]  # Optional: restrict tools
model: claude-sonnet-4-5  # Optional: specify model
---

# Agent System Prompt

Detailed instructions for the agent...

## Responsibilities
- Specific task 1
- Specific task 2

## Guidelines
- Best practice 1
- Best practice 2
```

### Sub Agent Best Practices

1. **Clear Purpose**: Each agent should have a single, well-defined responsibility
2. **Appropriate Tools**: Only grant tools the agent actually needs
3. **Detailed Instructions**: Provide comprehensive guidance in the system prompt
4. **Return Format**: Specify what the agent should return to main conversation
5. **Error Handling**: Include instructions for handling common error scenarios

### Sub Agent Anti-Patterns

❌ **AVOID:**
- Creating agents for simple, one-line tasks
- Overlapping agent responsibilities
- Agents that need access to main conversation context
- Too many agents for a single workflow

---

## Slash Commands

### What Are Slash Commands?

Slash commands are reusable prompts stored as Markdown files that:
- Execute frequently-used workflows
- Standardize common operations
- Can accept arguments
- Support frontmatter configuration

### When to Use Slash Commands

✅ **USE SLASH COMMANDS FOR:**

1. **Repetitive Workflows**
   - Creating new features following a template
   - Running standard test suites
   - Generating boilerplate code
   - Code review checklists

2. **Standardized Processes**
   - PR creation with standard format
   - Release preparation
   - Bug report templates
   - Deployment procedures

3. **Team Consistency**
   - Enforcing coding standards
   - Documentation templates
   - Review processes
   - Commit message formats

4. **Quick Access to Complex Prompts**
   - Multi-step instructions
   - Prompts with specific formatting requirements
   - Domain-specific terminology

### How to Create Slash Commands

Commands are stored in:
- **Project**: `.claude/commands/[name].md` (shared with team)
- **User**: `~/.claude/commands/[name].md` (personal)

```markdown
---
description: Brief description shown in /help
argument-hint: <optional-arg-name>
allowed-tools: [Read, Edit, Bash]  # Optional
model: claude-sonnet-4-5  # Optional
---

# Command Instructions

Detailed prompt for Claude...

{{arg}} can be used to reference command arguments
```

### Slash Command Best Practices

1. **Clear Arguments**: Use descriptive argument hints
2. **Frontmatter**: Always include a description
3. **Documentation**: Explain what the command does and when to use it
4. **Tool Restrictions**: Limit tools if the command has a narrow scope
5. **Namespacing**: Use directories for organization (e.g., `test/unit.md` → `/test/unit`)

### Slash Command Anti-Patterns

❌ **AVOID:**
- Commands that require extensive context gathering
- One-off tasks that won't be repeated
- Commands that duplicate built-in functionality
- Overly generic commands

---

## MCP Servers

### What Are MCP Servers?

MCP (Model Context Protocol) servers provide:
- **External tool integrations** (APIs, databases, services)
- **Data source connections** (filesystems, cloud storage)
- **Real-time information** (monitoring, analytics)
- **Remote capabilities** (cloud-hosted or local)

### When to Use MCP Servers

✅ **USE MCP SERVERS FOR:**

1. **External Service Integration**
   - Database queries (PostgreSQL, MySQL, MongoDB)
   - Cloud platforms (AWS, GCP, Vercel)
   - Monitoring services (Sentry, DataDog)
   - Version control platforms (GitHub, GitLab)

2. **Real-Time Data Access**
   - Live deployment status
   - Error tracking and logs
   - Analytics and metrics
   - API endpoints

3. **Specialized Tools**
   - Design tools (Figma)
   - Communication platforms (Slack)
   - CI/CD pipelines
   - Testing platforms

4. **Data Sources**
   - File systems and storage
   - Databases
   - APIs
   - Documentation systems

### How to Configure MCP Servers

MCP servers are configured in `~/.claude/mcp.json`:

```json
{
  "mcpServers": {
    "server-name": {
      "transport": {
        "type": "http",
        "url": "https://api.example.com/mcp"
      }
    },
    "local-server": {
      "transport": {
        "type": "stdio",
        "command": "node",
        "args": ["./path/to/server.js"]
      }
    }
  }
}
```

### MCP Server Best Practices

1. **Remote First**: Prefer HTTP servers over stdio when available (easier maintenance)
2. **Security**: Use environment variables for credentials
3. **Testing**: Verify connectivity with `claude mcp list`
4. **Scope**: Configure servers at user or project scope as appropriate
5. **Documentation**: Document what each server provides and when to use it

### Popular MCP Servers

| Server | Use Case | Type |
|--------|----------|------|
| PostgreSQL | Database queries | Local |
| Sentry | Error monitoring | Remote |
| Figma | Design context | Remote |
| Vercel | Deployment management | Remote |
| AWS | Cloud resource management | Remote |
| GitHub | Repository operations | Remote |

### MCP Server Anti-Patterns

❌ **AVOID:**
- Running local servers when remote alternatives exist
- Connecting servers you don't actually need
- Storing credentials directly in config files
- Using MCP for simple file operations (use built-in tools instead)

---

## Skills

### What Are Skills?

Skills are modular capabilities that:
- **Package domain knowledge** and expertise
- **Auto-load when relevant** to user requests
- **Include executable code** (scripts, templates, tools)
- **Organize instructions** in SKILL.md files

### When to Use Skills

✅ **USE SKILLS FOR:**

1. **Domain Expertise**
   - Framework-specific knowledge (React, Vue, Django)
   - Language-specific patterns (Python, TypeScript)
   - Platform expertise (AWS, Kubernetes)
   - Design patterns and architectures

2. **Complex Workflows**
   - Testing strategies
   - Deployment procedures
   - Code generation patterns
   - Migration processes

3. **Reusable Capabilities**
   - Custom linting rules
   - Code generators
   - Analysis tools
   - Template systems

4. **Cross-Project Knowledge**
   - Company coding standards
   - Architecture patterns
   - Security best practices
   - Accessibility guidelines

### How to Create Skills

Skills are stored in `.claude/skills/[skill-name]/`:

```
.claude/skills/react-testing/
├── SKILL.md          # Main instructions
├── templates/        # Optional supporting files
├── scripts/          # Optional executable scripts
└── examples/         # Optional example code
```

**SKILL.md structure:**

```markdown
---
name: react-testing
description: Best practices for testing React components
triggers: [test, testing, react, component test]  # Optional
---

# React Component Testing Skill

This skill provides expertise in testing React components...

## Testing Patterns

1. **Unit Tests**: Test components in isolation...
2. **Integration Tests**: Test component interactions...
3. **E2E Tests**: Test full user flows...

## Tools and Libraries

- Jest for unit testing
- React Testing Library for component testing
- Cypress for E2E testing

## Examples

[Include examples here]
```

### Skills Best Practices

1. **Clear Triggers**: Define when the skill should be loaded
2. **Comprehensive Documentation**: Include examples and patterns
3. **Supporting Files**: Provide templates and scripts
4. **Organization**: Group related skills together
5. **Testing**: Verify skills load correctly with test prompts

### Skills Anti-Patterns

❌ **AVOID:**
- Skills that are too generic (too often loaded unnecessarily)
- Duplicating built-in Claude knowledge
- Skills that require external dependencies not documented
- Overlapping skill responsibilities

---

## Decision Matrix

### Which Feature Should I Use?

Use this decision tree to choose the right feature:

```
START: What are you trying to accomplish?

├─ Need to integrate external tool/API?
│  └─ → USE MCP SERVER
│
├─ Need to package domain knowledge that auto-loads?
│  └─ → USE SKILL
│
├─ Need to execute complex task with isolated context?
│  └─ → USE SUB AGENT
│
└─ Need to run a repeated workflow or standardized prompt?
   └─ → USE SLASH COMMAND
```

### Detailed Scenarios

| Scenario | Recommended Feature | Why |
|----------|-------------------|-----|
| Code review process | Sub Agent | Complex analysis, isolated context |
| Create new component | Slash Command | Repeatable workflow |
| Access production database | MCP Server | External tool integration |
| React best practices | Skill | Domain knowledge, auto-loads |
| Generate test suite | Sub Agent | Complex, multi-file task |
| PR template | Slash Command | Standardized process |
| Monitor Sentry errors | MCP Server | External service |
| TypeScript patterns | Skill | Framework knowledge |
| Security audit | Sub Agent | Specialized expertise |
| Deploy to staging | Slash Command | Standard workflow |

---

## Best Practices

### General Guidelines

1. **Start Simple**: Begin with slash commands before creating sub agents
2. **Avoid Duplication**: Don't create multiple features for the same purpose
3. **Documentation**: Always document what you create and why
4. **Team Alignment**: For project-level features, ensure team agreement
5. **Iterative Improvement**: Start minimal and expand based on actual needs

### Composition Patterns

Features can work together:

1. **Slash Command → Sub Agent**
   - Command triggers agent for complex work
   - Example: `/review` command invokes code-review agent

2. **Sub Agent + MCP Server**
   - Agent uses MCP tools for external access
   - Example: Deployment agent uses Vercel MCP

3. **Skill + Slash Command**
   - Skill provides knowledge, command applies it
   - Example: Testing skill + `/test` command

4. **MCP Server + Skill**
   - Skill knows how to use MCP-provided tools
   - Example: AWS skill + AWS MCP server

### Performance Considerations

- **Sub Agents**: Isolate context but add latency
- **Slash Commands**: Fastest, no overhead
- **MCP Servers**: Add network latency
- **Skills**: Minimal overhead, auto-loaded

### Maintenance

1. **Regular Review**: Audit features quarterly
2. **Remove Unused**: Delete features that aren't being used
3. **Update Documentation**: Keep guides current
4. **Version Control**: Track changes to custom features
5. **Share Knowledge**: Document decisions and patterns

---

## Quick Reference

### File Locations

```
Project Structure:
.claude/
├── agents/
│   └── [agent-name]/
│       └── AGENT.md
├── commands/
│   └── [command-name].md
└── skills/
    └── [skill-name]/
        └── SKILL.md

User Global:
~/.claude/
├── commands/
│   └── [command-name].md
└── mcp.json
```

### Commands

```bash
# MCP Management
claude mcp add [name] --scope user
claude mcp list
claude mcp remove [name]

# Help
/help                    # List all commands
```

### Tool Invocation

```bash
# In conversation:
/command-name [args]     # Run slash command
@agent-name             # Invoke sub agent (when supported)

# Skills load automatically based on context
```

---

## Additional Resources

- **Official Docs**: docs.claude.com/en/docs/claude-code/
- **Sub Agents Guide**: docs.claude.com/en/docs/claude-code/sub-agents
- **Slash Commands**: docs.claude.com/en/docs/claude-code/slash-commands
- **MCP Protocol**: docs.claude.com/en/docs/claude-code/mcp
- **Skills Guide**: docs.claude.com/en/docs/claude-code/skills
- **GitHub Skills Examples**: github.com/anthropics/skills
- **Community Resources**: claudelog.com

---

*This guide is a living document. Update it as you learn more about effective feature usage in your projects.*
