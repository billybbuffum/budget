# Bootstrap Claude Code for Any Project

This guide explains how to use the `/bootstrap-claude-code` command to automatically generate Claude Code configuration for any project - **even brand new empty projects!**

## 🚀 New: Works for Empty Projects!

You can now use `/bootstrap-claude-code` **before writing any code**! Just tell Claude what you're building:

```
/bootstrap-claude-code I'm building an ovulation tracker app using React Native and Firebase
```

Claude will ask clarifying questions, then generate complete configuration tailored to your project - perfect for starting new projects with professional tooling from day one!

## What It Does

The `/bootstrap-claude-code` command works in two modes and automatically generates:

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

The `/bootstrap-claude-code` command works in **two modes**:

### Mode 1: Empty/New Project (With Description)

**Perfect for starting a brand new project!**

```bash
# Navigate to your empty project directory
cd ~/projects/my-new-app

# Run bootstrap with a description of what you're building
```

In Claude Code:
```
/bootstrap-claude-code I'm building an ovulation tracker app using React Native and Firebase
```

**What happens:**
1. Claude parses your description (project type, domain, tech stack)
2. Asks clarifying questions about:
   - Complete tech stack (database, auth, etc.)
   - Core features and entities
   - Architecture preference
   - Testing strategy
   - Deployment targets
3. Generates comprehensive configuration based on your answers
4. Optionally scaffolds initial project structure

**Examples:**
```
/bootstrap-claude-code I'm building an e-commerce platform with Next.js and PostgreSQL

/bootstrap-claude-code Building a fitness tracking mobile app with Flutter and Supabase

/bootstrap-claude-code Creating a task management API with Go and MongoDB

/bootstrap-claude-code I want to build a recipe sharing website using Django
```

### Mode 2: Existing Project (Analyze Codebase)

**For projects with existing code!**

```bash
# Navigate to your existing project
cd ~/projects/existing-project

# Run bootstrap without arguments
```

In Claude Code:
```
/bootstrap-claude-code
```

**What happens:**
1. Claude analyzes your codebase (languages, frameworks, architecture)
2. Identifies your business domain from code/docs
3. Generates appropriate sub agents (5-10 agents)
4. Creates relevant slash commands (6-12 commands)
5. Generates domain and tech-stack skills
6. Recommends MCPs for your tech stack
7. Creates comprehensive documentation

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

### For an Empty Project: React Native Ovulation Tracker

**Input:**
```
/bootstrap-claude-code I'm building an ovulation tracker app using React Native and Firebase
```

**Generated Configuration:**

**Sub Agents:**
- `code-reviewer` - React Native, TypeScript, and Expo best practices
- `test-generator` - Jest + React Native Testing Library patterns
- `security-auditor` - HIPAA compliance, health data encryption, Firebase security rules
- `refactoring-assistant` - Safe refactoring for React Native apps
- `documentation-generator` - README, API docs, user guides
- `ovulation-tracker-expert` - Cycle prediction algorithms, fertile window calculation, symptom validation, privacy compliance
- `ui-tester` - Interactive mobile UI testing
- `firebase-specialist` - Firestore optimization, Cloud Functions, auth patterns

**Slash Commands:**
- `/create-spec` - Create validated specification with domain expert review
- `/implement-spec` - Implement specification with quality gates
- `/new-feature` - Scaffold new feature (creates components, hooks, screens, types)
- `/review-pr` - Comprehensive PR review
- `/check-architecture` - Verify feature-based architecture
- `/test-ui` - Interactive UI testing on emulator/device
- `/run-tests` - Run Jest test suite
- `/build-ios` - Build iOS release
- `/build-android` - Build Android release
- `/deploy` - Deploy Firebase Functions and Firestore rules

**Skills:**
- `feature-based-architecture` - React Native feature folder patterns
- `ovulation-tracking` - Cycle prediction formulas, BBT ranges, HIPAA compliance, data validation rules
- `typescript-testing` - Jest patterns, React Testing Library, E2E testing
- `react-native-patterns` - Hooks, navigation, state management, performance
- `firestore-best-practices` - Security rules, data modeling, offline support

**MCP Recommendations:**
- React Native Debugger MCP - Mobile debugging
- Firebase MCP - Firestore operations, auth, functions
- npm MCP - Package management
- Expo MCP (if using Expo) - Build and deployment

**Documentation:**
- README with quick start, tech stack overview, domain summary
- FEATURE_USAGE_GUIDE with React Native specific examples
- WORKFLOWS with mobile development workflows (spec creation, UI testing, app store submission)
- MCP setup guide

**Optional Project Scaffolding:**
```
src/
├── features/
│   ├── auth/
│   ├── cycle-tracking/
│   └── predictions/
├── shared/
│   ├── components/
│   ├── hooks/
│   └── utils/
├── navigation/
├── services/
└── App.tsx
```

### For an Existing Project: Go + React E-commerce Platform

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

## Example Workflows

### Workflow 1: Brand New Empty Project

**Starting from scratch with just an idea:**

```bash
# 1. Create project directory
mkdir my-ovulation-tracker
cd my-ovulation-tracker
git init

# 2. Bootstrap with description (no code needed!)
# In Claude Code:
/bootstrap-claude-code I'm building an ovulation tracker app using React Native and Firebase

# 3. Answer Claude's clarifying questions:
# - Database? → Firestore
# - Core features? → Cycle tracking, symptom logging, predictions, notifications
# - Main entities? → User, Cycle, Symptom, Prediction
# - Architecture? → Feature-based (recommended for React Native)
# - Testing? → Jest + React Native Testing Library + E2E

# 4. Wait for generation (2-5 minutes)
# Claude creates:
# - ovulation-tracker-expert agent with cycle prediction algorithms
# - code-reviewer for React Native + TypeScript
# - test-generator for Jest patterns
# - security-auditor with HIPAA compliance checks
# - ui-tester for mobile UI testing
# - /new-feature, /test-ui, /create-spec, etc.
# - ovulation-tracking skill with domain knowledge
# - Complete documentation

# 5. Optionally scaffold project structure
# Claude asks: "Would you like me to scaffold the initial project structure?"
# Say yes to get starter files and folder structure!

# 6. Review generated configuration
cat .claude/README.md
ls -la .claude/agents/
ls -la .claude/commands/

# 7. Start building immediately!
/create-spec "Add cycle tracking with period logging"
/implement-spec docs/spec-cycle-tracking.md
```

### Workflow 2: Existing Project

**Already have code, want to add Claude Code:**

```bash
# 1. Navigate to your existing project
cd ~/projects/my-ecommerce-api
# (Project already has Go code, PostgreSQL, etc.)

# 2. Bootstrap without description (analyzes existing code)
# In Claude Code:
/bootstrap-claude-code

# 3. Claude analyzes your codebase
# - Detects: Go, Chi router, PostgreSQL, Clean Architecture
# - Identifies domain: E-commerce (from models, endpoints, tests)
# - Discovers: Inventory, orders, payments, shipping

# 4. Wait for generation (2-5 minutes)
# Claude creates agents specific to YOUR codebase:
# - code-reviewer for Go + your specific patterns
# - ecommerce-expert with inventory/pricing logic
# - postgres-optimizer for your database
# - Slash commands matching your structure
# - Skills for your architecture pattern

# 5. Review and start using
cat .claude/README.md
/review-pr
/create-spec "Add abandoned cart recovery"
```

### Workflow 3: Quick Prototype to Production

**Evolve from idea to production:**

```bash
# Day 1: Start with idea
mkdir fitness-tracker
cd fitness-tracker
/bootstrap-claude-code Building a fitness app with Flutter and Supabase
# → Claude generates config, scaffolds structure
# → Start coding with full Claude Code support

# Week 2: Add features with specs
/create-spec "Add workout logging"
/implement-spec docs/spec-workout-logging.md
/test-ui "workout logging flow"

# Week 4: Code review and polish
/review-pr
# → Automated review finds issues
# → domain-expert validates fitness calculations
# → security-auditor checks data privacy

# Week 6: Deploy to production
/deploy production
# → Uses deployment-helper agent
# → Runs pre-deployment checks
# → Ships with confidence!
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
