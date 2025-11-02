# Bootstrap Claude Code for Any Project

This guide explains how to use the `/bootstrap-claude-code` command to automatically generate Claude Code configuration for any project.

## What It Does

The `/bootstrap-claude-code` command analyzes a project and automatically generates:

✅ **Sub Agents** - Specialized agents for code review, testing, security, refactoring, documentation, and domain expertise
✅ **Slash Commands** - Common workflows like /new-feature, /review-pr, /create-spec, /implement-spec
✅ **Skills** - Auto-loading knowledge modules for architecture, domain, and testing
✅ **MCP Recommendations** - Suggested MCP servers for the tech stack
✅ **Documentation** - README, FEATURE_USAGE_GUIDE, and WORKFLOWS

## Making It Available Globally (User-Level)

To use this command across all your projects, move it to your user-level Claude Code configuration:

### Option 1: Copy to User Config (Recommended)

```bash
# Create user-level commands directory if it doesn't exist
mkdir -p ~/.claude/commands

# Copy the bootstrap command to user config
cp .claude/commands/bootstrap-claude-code.md ~/.claude/commands/

# Now you can use /bootstrap-claude-code in ANY project!
```

### Option 2: Symlink to User Config

```bash
# Create symbolic link in user config
mkdir -p ~/.claude/commands
ln -s /home/user/budget/.claude/commands/bootstrap-claude-code.md ~/.claude/commands/

# Updates to the source file will be reflected globally
```

## Using the Command

### Step 1: Navigate to Your Project

```bash
cd ~/projects/my-new-project
```

### Step 2: Run the Bootstrap Command

In Claude Code:
```
/bootstrap-claude-code
```

### Step 3: Review Generated Configuration

The command will:
1. Analyze your codebase (languages, frameworks, architecture)
2. Identify your business domain
3. Generate appropriate sub agents (5-10 agents)
4. Create relevant slash commands (6-12 commands)
5. Generate domain and tech-stack skills
6. Recommend MCPs for your tech stack
7. Create comprehensive documentation

### Step 4: Verify and Customize

After generation:
```bash
# Review generated configuration
ls -la .claude/agents/
ls -la .claude/commands/
ls -la .claude/skills/

# Read the documentation
cat .claude/README.md

# Test a command
/new-feature test-feature
```

## What Gets Generated

### For a Go + React E-commerce Project:

**Sub Agents:**
- `code-reviewer` - Go and React best practices
- `test-generator` - Go table-driven tests, React Testing Library
- `security-auditor` - SQL injection, XSS, payment security
- `refactoring-assistant` - Safe refactoring for Go and React
- `api-documenter` - OpenAPI documentation
- `ecommerce-expert` - Inventory, pricing, cart, checkout logic
- `ui-tester` - Playwright UI testing

**Slash Commands:**
- `/create-spec` - Create validated specification
- `/implement-spec` - Implement specification
- `/new-feature` - Scaffold new feature
- `/new-endpoint` - Add new API endpoint
- `/review-pr` - Comprehensive PR review
- `/check-architecture` - Verify clean architecture
- `/test-endpoint` - Test API endpoints
- `/test-ui` - Interactive UI testing
- `/run-tests` - Run test suite

**Skills:**
- `go-clean-architecture` - Go layer patterns
- `ecommerce` - E-commerce domain knowledge
- `go-testing` - Go testing patterns
- `react-patterns` - React best practices
- `postgres-best-practices` - Database optimization

**MCP Recommendations:**
- Playwright MCP - UI testing
- PostgreSQL MCP - Database operations
- Stripe MCP - Payment processing
- npm MCP - Package management

### For a Python Django SaaS Project:

**Sub Agents:**
- `code-reviewer` - Python and Django best practices
- `test-generator` - pytest patterns
- `security-auditor` - Django security, OWASP
- `database-optimizer` - Query optimization
- `saas-expert` - Multi-tenancy, subscriptions, billing

**Slash Commands:**
- `/create-spec`, `/implement-spec`, `/new-feature`, etc.
- `/migrate-db` - Run Django migrations
- `/test-api` - Test API endpoints

**Skills:**
- `django-patterns` - Django best practices
- `saas` - SaaS domain knowledge
- `python-testing` - pytest patterns
- `postgres-best-practices` - Database optimization

**MCP Recommendations:**
- PostgreSQL MCP
- Stripe MCP
- AWS MCP
- pip MCP

## Example Workflow

### Setting Up a New Project

```bash
# 1. Start new project
mkdir my-saas-app
cd my-saas-app
git init

# 2. Add some initial code (or clone existing repo)
# ... set up your project structure ...

# 3. Bootstrap Claude Code configuration
# In Claude Code:
/bootstrap-claude-code

# 4. Wait for analysis and generation (2-5 minutes)
# Claude will:
# - Analyze your codebase
# - Generate 5-10 sub agents
# - Create 6-12 slash commands
# - Generate 3-5 skills
# - Create documentation

# 5. Review generated configuration
cat .claude/README.md
cat .claude/FEATURE_USAGE_GUIDE.md

# 6. Start using it!
/create-spec "Add user authentication"
/implement-spec docs/spec-user-authentication.md
```

## Customization After Generation

The generated configuration is a starting point. Customize it:

### Add Project-Specific Agents

```bash
# Create custom agent for your unique needs
mkdir -p .claude/agents/my-custom-agent
vim .claude/agents/my-custom-agent/AGENT.md
```

### Add Custom Commands

```bash
# Create workflow command for your team
vim .claude/commands/my-workflow.md
```

### Extend Skills

```bash
# Add more domain knowledge
vim .claude/skills/my-domain/SKILL.md
```

## Quality Expectations

The bootstrap command generates **production-quality** configuration:

- **Sub Agents**: 500-2000 lines each with comprehensive checklists
- **Slash Commands**: 300-700 lines with explicit workflows
- **Skills**: 400-1000 lines with domain knowledge
- **Documentation**: Complete guides with examples

Quality is comparable to the budget app `.claude/` directory (this project).

## When to Use

✅ **Starting a new project** - Get Claude Code setup immediately
✅ **Existing project without Claude Code** - Retrofit configuration
✅ **Team onboarding** - Standardize Claude Code usage across team
✅ **Project templates** - Create reusable templates for similar projects

## When NOT to Use

❌ **Project already has comprehensive Claude Code setup** - Manual customization better
❌ **Very unique/experimental project** - May need custom agents not generated
❌ **Quick script or prototype** - Overkill for simple projects

## Troubleshooting

### "Command not found"

Make sure you've moved the command to user config:
```bash
cp .claude/commands/bootstrap-claude-code.md ~/.claude/commands/
```

### "Generated agents not specific enough"

The command works best with:
- Clear project structure
- Documented domain (README, docs/)
- Standard tech stack

Add more documentation to your project, then re-run.

### "Want to regenerate configuration"

Delete `.claude/` directory and re-run:
```bash
rm -rf .claude/
/bootstrap-claude-code
```

## Advanced: Creating Project Templates

Use bootstrap to create reusable templates:

```bash
# 1. Create template project
mkdir template-go-api
cd template-go-api

# 2. Add basic structure
mkdir -p cmd/api internal/domain internal/application internal/infrastructure
touch go.mod README.md

# 3. Bootstrap Claude Code
/bootstrap-claude-code

# 4. Commit template
git add .
git commit -m "Go API template with Claude Code"

# 5. Use template for new projects
cp -r template-go-api my-new-api
cd my-new-api
# Start coding with Claude Code already configured!
```

## Comparison: Manual vs Bootstrap

### Manual Setup (Old Way)
- ⏱️ 2-4 hours to create all agents, commands, skills
- 🧠 Requires deep knowledge of Claude Code features
- 📝 Risk of inconsistent quality
- 🔄 Must repeat for each project

### Bootstrap (New Way)
- ⏱️ 2-5 minutes for comprehensive setup
- 🤖 Automatic analysis and generation
- 📊 Consistent, production-quality output
- 🚀 Reusable across unlimited projects

## Real-World Example: Budget App

This project's `.claude/` directory was manually created over several hours. It includes:
- 7 sub agents (2000+ lines total)
- 10 slash commands (5000+ lines total)
- 4 skills (3000+ lines total)
- Comprehensive documentation

The `/bootstrap-claude-code` command can generate similar quality configuration **automatically** in 2-5 minutes.

## Feedback and Iteration

After using the bootstrap command:

1. **What worked well?** - Keep those agents/commands
2. **What's missing?** - Add custom agents/commands
3. **What's not useful?** - Delete or modify
4. **What needs refinement?** - Edit generated files

The generated configuration is a **starting point**, not final.

## Next Steps

1. **Move command to user config**: `cp .claude/commands/bootstrap-claude-code.md ~/.claude/commands/`
2. **Try it on a project**: Navigate to any project and run `/bootstrap-claude-code`
3. **Review generated config**: Check `.claude/` directory
4. **Start developing**: Use `/create-spec` and `/implement-spec`
5. **Customize as needed**: Add project-specific agents/commands

---

**The goal: Spend less time configuring tools, more time building features.**

With `/bootstrap-claude-code`, every project gets professional-grade Claude Code configuration in minutes, not hours.
