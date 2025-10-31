# Claude Code Configuration

This directory contains Claude Code customizations for the Budget application, including sub agents, slash commands, skills, and documentation.

## 📁 Directory Structure

```
.claude/
├── README.md                    # This file
├── FEATURE_USAGE_GUIDE.md      # When to use each Claude Code feature
├── MCP_RECOMMENDATIONS.md       # Recommended MCP servers to install
├── agents/                      # Sub agents for specialized tasks
│   ├── code-reviewer/
│   ├── test-generator/
│   ├── api-documenter/
│   ├── refactoring-assistant/
│   ├── security-auditor/
│   └── budget-domain-expert/
├── commands/                    # Slash commands for common workflows
│   ├── new-feature.md
│   ├── new-endpoint.md
│   ├── review-pr.md
│   ├── check-architecture.md
│   ├── test-endpoint.md
│   ├── run-tests.md
│   └── generate-docs.md
└── skills/                      # Skills that auto-load when relevant
    ├── go-clean-architecture/
    ├── zero-based-budgeting/
    ├── go-testing/
    └── sqlite-best-practices/
```

## 🤖 Sub Agents

Sub agents are specialized AI assistants for complex tasks with isolated context.

### Available Agents

1. **code-reviewer** - Reviews Go code for quality and architecture compliance
   - Clean architecture verification
   - Go best practices
   - Budget domain logic
   - Security checks

2. **test-generator** - Generates comprehensive unit and integration tests
   - Service tests with mocks
   - Repository tests with test database
   - HTTP handler tests
   - Table-driven test patterns

3. **api-documenter** - Creates API documentation
   - OpenAPI specifications
   - Endpoint documentation
   - Request/response examples
   - Usage workflows

4. **refactoring-assistant** - Helps refactor code safely
   - Identifies code smells
   - Proposes improvements
   - Maintains clean architecture
   - Improves testability

5. **security-auditor** - Performs security audits
   - SQL injection detection
   - Input validation checks
   - Sensitive data handling
   - Financial data integrity

6. **budget-domain-expert** - Zero-based budgeting specialist
   - Verifies budget calculations
   - Validates business rules
   - Ensures correct rollover behavior
   - Credit card logic validation

### Using Sub Agents

In conversation with Claude, reference agents like:
- "Invoke the code-reviewer agent to review my changes"
- "Use test-generator to create tests for AccountService"
- "Have the security-auditor check for vulnerabilities"

## ⚡ Slash Commands

Slash commands are reusable workflows for common tasks.

### Available Commands

- **/new-feature** `<feature-name>` - Scaffold new feature following clean architecture
- **/new-endpoint** `<description>` - Add new API endpoint
- **/review-pr** - Review PR with budget app specific checks
- **/check-architecture** - Verify clean architecture compliance
- **/test-endpoint** `<path>` - Test API endpoint with curl
- **/run-tests** - Run all tests and report results
- **/generate-docs** - Generate comprehensive API documentation

### Using Slash Commands

Type `/` in Claude Code to see available commands, then:
```
/new-feature user-authentication
/test-endpoint /api/accounts
/review-pr
```

## 🎯 Skills

Skills are modular capabilities that Claude automatically loads when relevant.

### Available Skills

1. **go-clean-architecture** - Go clean architecture patterns
   - Layer separation
   - Dependency injection
   - Repository pattern
   - Service pattern

2. **zero-based-budgeting** - Zero-based budgeting domain knowledge
   - Ready to Assign calculation
   - Category Available with rollover
   - Credit card budgeting
   - Allocation rules

3. **go-testing** - Go testing best practices
   - Table-driven tests
   - Mocking strategies
   - Integration testing
   - Test organization

4. **sqlite-best-practices** - SQLite optimization and patterns
   - Query optimization
   - Transaction handling
   - Schema design
   - Performance tuning

### Skill Triggers

Skills auto-load based on keywords in your conversation:
- Mention "clean architecture" → loads go-clean-architecture skill
- Mention "budget" or "allocation" → loads zero-based-budgeting skill
- Mention "test" or "testing" → loads go-testing skill
- Mention "sqlite" or "database" → loads sqlite-best-practices skill

## 📚 Documentation

### FEATURE_USAGE_GUIDE.md

Comprehensive guide explaining:
- When to use sub agents vs slash commands vs skills vs MCPs
- Decision matrix for choosing the right tool
- Best practices and patterns
- Common anti-patterns to avoid

### MCP_RECOMMENDATIONS.md

Recommended MCP servers to install:
- **Priority 1**: GitHub, Sentry, PostgreSQL
- **Priority 2**: Filesystem, Docker, Git
- **Priority 3**: Slack, AWS, Vercel, Web Search

Includes installation instructions and use cases.

## 🚀 Getting Started

### 1. Review the Guides

Start by reading:
1. `FEATURE_USAGE_GUIDE.md` - Understand when to use each feature
2. `MCP_RECOMMENDATIONS.md` - Install recommended MCPs

### 2. Try Slash Commands

Most immediate value comes from slash commands:
```
/check-architecture
/test-endpoint /api/accounts
/review-pr
```

### 3. Invoke Sub Agents

For complex tasks, use sub agents:
```
"Invoke the code-reviewer agent to review the AllocationService"
"Use test-generator to create a comprehensive test suite"
```

### 4. Let Skills Auto-Load

Skills automatically activate based on context. Just work naturally and they'll help when relevant.

## 💡 Example Workflows

### Creating a New Feature

```
/new-feature transaction-import

→ Claude scaffolds:
  - Domain entity
  - Repository interface
  - Service implementation
  - HTTP handler
  - Router registration
  - Database schema

Then:
"Invoke test-generator to create tests"
"Invoke code-reviewer to verify implementation"
```

### Reviewing Code

```
/review-pr

→ Claude:
  1. Checks architecture compliance
  2. Reviews Go best practices
  3. Validates budget logic
  4. Runs security checks
  5. Provides detailed feedback
```

### Testing an API

```
/test-endpoint /api/accounts

→ Claude:
  1. Tests GET requests
  2. Tests POST with valid data
  3. Tests error cases
  4. Validates responses
  5. Reports results
```

## 🎓 Learning Path

### Week 1: Basics
- Read FEATURE_USAGE_GUIDE.md
- Try 2-3 slash commands
- Create one new feature with /new-feature

### Week 2: Advanced
- Invoke code-reviewer agent
- Generate tests with test-generator
- Review architecture with /check-architecture

### Week 3: Mastery
- Install recommended MCPs
- Create custom workflows
- Combine agents + commands + skills

## 🔧 Customization

### Adding Your Own Slash Commands

Create `.claude/commands/your-command.md`:
```markdown
---
description: What your command does
argument-hint: <optional-arg>
---

# Your Command

Instructions for Claude...
```

### Creating Custom Skills

Create `.claude/skills/your-skill/SKILL.md`:
```markdown
---
name: your-skill
description: What your skill provides
triggers: [keyword1, keyword2]
---

# Your Skill

Knowledge and patterns...
```

### Adding Sub Agents

Create `.claude/agents/your-agent/AGENT.md`:
```markdown
---
name: your-agent
description: What your agent does
tools: [Read, Write, Bash]
---

# Your Agent

Specialized instructions...
```

## 📖 Best Practices

### Use Sub Agents For:
- Complex, multi-step tasks
- Specialized analysis (security, testing)
- Tasks requiring isolated context

### Use Slash Commands For:
- Repetitive workflows
- Standard procedures
- Quick access to complex prompts

### Use Skills For:
- Domain knowledge
- Framework patterns
- Cross-cutting concerns

### Use MCPs For:
- External service integration
- Real-time data access
- Production monitoring

## 🐛 Troubleshooting

### Slash Command Not Found
- Check file exists in `.claude/commands/`
- Verify frontmatter is valid
- Restart Claude Code if needed

### Agent Not Working
- Check AGENT.md exists in agent directory
- Verify tools are correctly specified
- Try invoking explicitly

### Skill Not Loading
- Check trigger keywords are relevant
- Verify SKILL.md frontmatter
- Skills load automatically, be patient

## 🤝 Contributing

When adding new agents, commands, or skills:
1. Follow existing patterns
2. Document thoroughly
3. Test before committing
4. Update this README

## 📝 Maintenance

- **Monthly**: Review and update agents/commands
- **Quarterly**: Add new skills based on patterns
- **Yearly**: Major review and cleanup

## 🔗 Resources

- [Claude Code Documentation](https://docs.claude.com/en/docs/claude-code/)
- [Sub Agents Guide](https://docs.claude.com/en/docs/claude-code/sub-agents)
- [Slash Commands Guide](https://docs.claude.com/en/docs/claude-code/slash-commands)
- [Skills Guide](https://docs.claude.com/en/docs/claude-code/skills)
- [MCP Servers](https://docs.claude.com/en/docs/claude-code/mcp)

---

**Created:** October 31, 2025
**Last Updated:** October 31, 2025
**Version:** 1.0.0

*This configuration was created to fully unlock agentic coding capabilities for the Budget application.*
