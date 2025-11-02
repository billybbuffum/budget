# Bootstrap Claude Code Configuration for Any Project

You are tasked with analyzing the current project and generating a comprehensive Claude Code configuration including sub agents, slash commands, skills, and MCP recommendations.

## Phase 1: Deep Project Analysis

### 1.1 Analyze Codebase Structure

**Discover technology stack:**
```bash
# Find all file types to understand tech stack
find . -type f -name "*.go" -o -name "*.js" -o -name "*.ts" -o -name "*.py" -o -name "*.java" -o -name "*.rb" -o -name "*.php" -o -name "*.rs" -o -name "*.swift" -o -name "*.kt" | head -50

# Check for package/dependency files
ls -la | grep -E "(package.json|go.mod|requirements.txt|Gemfile|composer.json|Cargo.toml|pom.xml|build.gradle)"

# Examine project structure
ls -R | head -100
```

**Identify frameworks and tools:**
- Read package.json, go.mod, requirements.txt, etc.
- Look for framework-specific files (e.g., next.config.js, django settings, etc.)
- Identify testing frameworks
- Identify build tools
- Identify database technologies

**Analyze architecture patterns:**
- Look for architecture indicators (clean architecture, MVC, microservices, monolith, etc.)
- Examine directory structure (domain/, application/, infrastructure/, etc.)
- Check for design patterns in code

### 1.2 Analyze Domain and Business Logic

**Understand the business domain:**
- Read README.md, docs/, specifications
- Analyze main domain entities and models
- Identify core business rules and logic
- Understand the problem space (e.g., e-commerce, fintech, healthcare, SaaS, etc.)

**Search for domain-specific concepts:**
- Look for domain terminology in code comments
- Examine API endpoints and function names
- Review test files for business scenarios
- Check for domain documentation

### 1.3 Identify Current Development Workflows

**Examine existing practices:**
- Check for CI/CD configuration (.github/workflows, .gitlab-ci.yml, etc.)
- Review testing setup and patterns
- Look for code quality tools (linters, formatters)
- Identify deployment practices
- Check for existing documentation generation

## Phase 2: Generate Sub Agents

Based on the analysis, create sub agents in `.claude/agents/` directory.

### 2.1 Always Create These Core Agents:

**1. code-reviewer**
- Language-specific best practices
- Framework-specific patterns
- Architecture compliance checks
- Code quality verification
- Performance considerations

**2. test-generator**
- Framework-specific test patterns
- Unit test generation
- Integration test generation
- Test coverage strategies
- Mock/stub patterns

**3. security-auditor**
- Language-specific vulnerabilities (SQL injection, XSS, etc.)
- Framework security best practices
- Authentication/authorization checks
- Data validation
- Secrets management

**4. refactoring-assistant**
- Safe refactoring strategies
- Architecture preservation
- Breaking change detection
- Migration assistance

**5. documentation-generator**
- API documentation (OpenAPI, JSDoc, GoDoc, etc.)
- Architecture documentation
- README generation
- Code comments

### 2.2 Create Domain-Specific Agent:

**domain-expert**
- Business logic validation
- Domain-specific calculations
- Business rule compliance
- Domain terminology and concepts
- Use case validation

*Name it based on domain (e.g., `ecommerce-expert`, `fintech-expert`, `healthcare-expert`)*

### 2.3 Create Conditional Agents Based on Tech Stack:

**If API/backend project:**
- api-designer - RESTful API design, endpoint structure

**If frontend project:**
- ui-tester - Interactive UI testing with Playwright MCP
- accessibility-auditor - WCAG compliance, a11y checks

**If database-heavy:**
- database-optimizer - Query optimization, schema design

**If has deployment config:**
- deployment-helper - CI/CD, containerization, cloud deployment

## Phase 3: Generate Slash Commands

Create slash commands in `.claude/commands/` directory.

### 3.1 Always Create These Commands:

**1. /new-feature** `<feature-name>`
- Scaffold new feature following project architecture
- Create necessary files/folders
- Generate boilerplate code

**2. /review-pr**
- Comprehensive PR review
- Architecture compliance
- Code quality checks
- Security validation

**3. /run-tests**
- Execute test suite
- Report results
- Identify failing tests

**4. /check-architecture**
- Verify architecture compliance
- Check layer boundaries
- Validate dependency direction

### 3.2 Spec-Driven Development Commands:

**5. /create-spec** `<feature-description>`
- Create validated specification
- Domain expert review
- Security review
- Technical design

**6. /implement-spec** `<spec-url>`
- Implement validated specification
- Code review loop
- Testing loop
- Quality gates

### 3.3 Tech Stack Specific Commands:

**If has API:**
- /test-endpoint `<path>` - Test API endpoints
- /new-endpoint `<description>` - Add new API endpoint
- /generate-api-docs - Generate API documentation

**If has frontend:**
- /test-ui `<workflow>` - Interactive UI testing
- /check-accessibility - Check WCAG compliance

**If has database:**
- /migrate-db - Run database migrations
- /optimize-query `<query>` - Analyze and optimize database queries

**If has deployment:**
- /deploy `<environment>` - Deploy to environment
- /check-deployment - Verify deployment health

## Phase 4: Generate Skills

Create skills in `.claude/skills/` directory.

### 4.1 Architecture Skill:

**{architecture-pattern}-architecture** (e.g., clean-architecture, mvc, microservices)
- Layer patterns
- Dependency rules
- Design patterns
- File organization

### 4.2 Domain Skill:

**{domain-name}** (e.g., ecommerce, fintech, healthcare)
- Domain terminology
- Business rules
- Key formulas/calculations
- Use cases
- Regulatory requirements

### 4.3 Testing Skill:

**{language}-testing** (e.g., go-testing, javascript-testing)
- Testing framework patterns
- Mock/stub strategies
- Test organization
- Coverage strategies

### 4.4 Framework/Database Skills:

**If specific framework** (e.g., react-patterns, django-patterns):
- Framework best practices
- Common patterns
- Performance optimization

**If database** (e.g., postgres-best-practices, sqlite-best-practices):
- Query optimization
- Schema design
- Transaction handling
- Performance tuning

## Phase 5: Generate MCP Recommendations

Create `.claude/MCP_RECOMMENDATIONS.md` with:

### 5.1 Priority 1 MCPs (Always Recommend):

**1. Playwright MCP** - UI testing (if frontend)
**2. Filesystem MCP** - File operations
**3. Git MCP** - Git operations

### 5.2 Tech Stack Specific MCPs:

**If Node.js:**
- npm MCP - Package management

**If Python:**
- pip MCP - Package management
- jupyter MCP - Notebook integration

**If has database:**
- database MCP for specific database (postgres, mysql, sqlite)

**If has cloud deployment:**
- AWS/GCP/Azure MCP based on provider

**If has monitoring:**
- observability MCP (Datadog, New Relic, etc.)

## Phase 6: Generate Documentation

### 6.1 Create README.md

**Structure:**
```markdown
# Claude Code Configuration for {Project Name}

## Overview
{Project description and domain}

## Sub Agents
{List all generated sub agents with descriptions}

## Slash Commands
{List all generated commands with usage examples}

## Skills
{List all generated skills with trigger contexts}

## MCP Recommendations
{Link to MCP_RECOMMENDATIONS.md}

## Example Workflows
{Show common development workflows}

## Learning Path
{Suggested learning progression}
```

### 6.2 Create FEATURE_USAGE_GUIDE.md

**Include:**
- When to use sub agents vs slash commands vs skills
- Decision matrix
- Best practices
- Anti-patterns
- Examples

### 6.3 Create WORKFLOWS.md

**Include:**
- Spec-driven development workflow
- Feature development workflow
- Testing workflow
- Code review workflow
- Deployment workflow
- Project-specific workflows

## Phase 7: Execution Plan

### 7.1 Create Todo List

Before starting, create comprehensive todo list:
```
1. Analyze project structure and tech stack
2. Identify domain and business logic
3. Create code-reviewer agent
4. Create test-generator agent
5. Create security-auditor agent
6. Create refactoring-assistant agent
7. Create documentation-generator agent
8. Create domain-expert agent
9. Create [tech-stack specific agents]
10. Create /new-feature command
11. Create /review-pr command
12. Create /run-tests command
13. Create /check-architecture command
14. Create /create-spec command
15. Create /implement-spec command
16. Create [tech-stack specific commands]
17. Create architecture skill
18. Create domain skill
19. Create testing skill
20. Create [framework/database skills]
21. Create MCP_RECOMMENDATIONS.md
22. Create README.md
23. Create FEATURE_USAGE_GUIDE.md
24. Create WORKFLOWS.md
25. Commit and push changes
```

### 7.2 Execute Systematically

- Work through each item sequentially
- Mark items as in_progress when starting
- Mark items as completed when done
- Create comprehensive, high-quality artifacts
- Use the budget app `.claude/` directory as a reference template

### 7.3 Quality Standards

**For each sub agent:**
- 500-2000 lines of comprehensive instructions
- Specific to tech stack and domain
- Include checklists and validation criteria
- Provide examples
- Define clear success criteria

**For each slash command:**
- 300-700 lines of detailed workflow
- Explicit agent invocation instructions
- Quality gates and review loops
- Clear phase-by-phase execution
- Example usage

**For each skill:**
- 400-1000 lines of domain/technical knowledge
- Auto-loading triggers defined
- Key concepts documented
- Common patterns included
- Best practices and anti-patterns

**For documentation:**
- Comprehensive and professional
- Include examples and workflows
- Learning path for new developers
- Quick reference guides

## Phase 8: Validation and Commit

### 8.1 Validate Generated Configuration

- Ensure all sub agents have AGENT.md files
- Verify all slash commands work
- Check skills have proper triggers
- Validate MCP recommendations are appropriate
- Review documentation for completeness

### 8.2 Commit Changes

```bash
git add .claude/
git commit -m "Bootstrap Claude Code configuration for {project name}

Generated comprehensive Claude Code setup including:
- {N} sub agents (code-reviewer, test-generator, domain-expert, etc.)
- {N} slash commands (spec-driven development, testing, etc.)
- {N} skills (architecture, domain, testing, etc.)
- MCP recommendations
- Complete documentation

Tech stack: {languages, frameworks, tools}
Domain: {domain description}
Architecture: {architecture pattern}
"
git push
```

## Important Guidelines

**Be Thorough:**
- Don't rush through analysis
- Create comprehensive, production-quality artifacts
- Use the budget app `.claude/` directory as quality benchmark

**Be Specific:**
- Tailor everything to the actual project
- Use real domain terminology
- Reference actual tech stack
- Include project-specific examples

**Be Practical:**
- Focus on workflows developers will actually use
- Create tools that save time
- Automate repetitive tasks
- Provide clear value

**Be Complete:**
- Don't skip any phases
- Generate all recommended artifacts
- Write comprehensive documentation
- Test that commands work

## Success Criteria

You have successfully bootstrapped Claude Code configuration when:

✅ All core sub agents created (5+ agents)
✅ Domain-specific agent created with business logic
✅ Tech-stack specific agents created
✅ Essential slash commands created (6+ commands)
✅ Spec-driven development workflow implemented
✅ Architecture, domain, and testing skills created
✅ MCP recommendations documented
✅ Complete documentation suite created (README, FEATURE_USAGE_GUIDE, WORKFLOWS)
✅ All changes committed and pushed
✅ Configuration is immediately usable by development team

Now analyze this project and generate a comprehensive Claude Code configuration!
