# MCP Server Recommendations for Budget Application

> **Last Updated:** October 31, 2025
>
> This document recommends Model Context Protocol (MCP) servers to enhance the development experience for the Budget application.

---

## Table of Contents

1. [What are MCP Servers?](#what-are-mcp-servers)
2. [Recommended MCPs](#recommended-mcps)
3. [Priority Levels](#priority-levels)
4. [Installation Guide](#installation-guide)
5. [Configuration Examples](#configuration-examples)

---

## What are MCP Servers?

MCP (Model Context Protocol) servers extend Claude Code's capabilities by providing access to external tools, services, and data sources. They run locally or remotely and expose their functionality through a standardized protocol.

**Benefits:**
- Access external services (databases, cloud platforms, monitoring)
- Query live data in real-time
- Automate workflows across tools
- Integrate with development infrastructure

---

## Recommended MCPs

### Priority 1: Essential for Development

#### 1. **GitHub MCP Server**

**Why You Need It:**
- Enhanced PR management and review
- Issue tracking and project management
- Repository operations and insights
- Code search across repositories

**Use Cases for Budget App:**
- Create and manage PRs directly from Claude
- Review code changes with context
- Track issues and feature requests
- Search codebase across branches
- View CI/CD status

**Installation:**
```bash
claude mcp add github --scope user
```

**Configuration:** `~/.claude/mcp.json`
```json
{
  "mcpServers": {
    "github": {
      "transport": {
        "type": "http",
        "url": "https://api.github.com/mcp"
      },
      "env": {
        "GITHUB_TOKEN": "your-personal-access-token"
      }
    }
  }
}
```

**Features:**
- List and filter issues/PRs
- Comment on PRs
- Review code changes
- Check CI status
- Search code

---

#### 2. **PostgreSQL MCP Server** (Future Migration)

**Why You Need It:**
- Currently using SQLite, but may need PostgreSQL for production
- Test queries against production-like database
- Migration planning and testing
- Performance testing with realistic data

**Use Cases for Budget App:**
- Plan migration from SQLite to PostgreSQL
- Test queries in both databases
- Validate schema changes
- Performance comparisons
- Production database access (read-only)

**Installation:**
```bash
npm install -g @modelcontextprotocol/server-postgres
claude mcp add postgres --scope user
```

**Configuration:**
```json
{
  "mcpServers": {
    "postgres": {
      "transport": {
        "type": "stdio",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-postgres", "postgresql://localhost/budget"]
      }
    }
  }
}
```

**Features:**
- Run queries directly
- Inspect schema
- Analyze query performance
- View table statistics

---

#### 3. **Sentry MCP Server** (Error Monitoring)

**Why You Need It:**
- Production error monitoring
- Real-time error tracking
- Debug production issues
- Track error trends

**Use Cases for Budget App:**
- Monitor production errors in real-time
- Investigate error reports
- Track error resolution
- Analyze error patterns
- Debug user-reported issues

**Installation:**
```bash
claude mcp add sentry --scope user
```

**Configuration:**
```json
{
  "mcpServers": {
    "sentry": {
      "transport": {
        "type": "http",
        "url": "https://sentry.io/api/0/mcp"
      },
      "env": {
        "SENTRY_AUTH_TOKEN": "your-auth-token",
        "SENTRY_ORG": "your-org",
        "SENTRY_PROJECT": "budget"
      }
    }
  }
}
```

**Features:**
- View recent errors
- Search error logs
- View error details and stack traces
- Track error frequency
- Resolve issues

---

### Priority 2: Highly Recommended

#### 4. **Filesystem MCP Server**

**Why You Need It:**
- Enhanced file operations
- Search across large codebases
- File system navigation
- Pattern matching

**Use Cases for Budget App:**
- Search for patterns across all files
- Find all usages of functions/types
- Navigate project structure
- Bulk file operations

**Installation:**
```bash
npm install -g @modelcontextprotocol/server-filesystem
claude mcp add filesystem --scope project
```

**Configuration:**
```json
{
  "mcpServers": {
    "filesystem": {
      "transport": {
        "type": "stdio",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/budget"]
      }
    }
  }
}
```

---

#### 5. **Docker MCP Server**

**Why You Need It:**
- Container management
- View logs and status
- Debug container issues
- Manage deployments

**Use Cases for Budget App:**
- Start/stop application containers
- View container logs
- Inspect container status
- Debug Docker issues
- Manage Docker Compose services

**Installation:**
```bash
npm install -g @modelcontextprotocol/server-docker
claude mcp add docker --scope user
```

**Features:**
- List containers
- View container logs
- Inspect container details
- Execute commands in containers
- Manage networks and volumes

---

#### 6. **Git MCP Server**

**Why You Need It:**
- Enhanced Git operations
- Repository insights
- Branch management
- Commit history analysis

**Use Cases for Budget App:**
- View detailed commit history
- Analyze code changes over time
- Branch comparison
- Git statistics and insights
- Advanced Git operations

**Installation:**
```bash
npm install -g @modelcontextprotocol/server-git
claude mcp add git --scope project
```

---

### Priority 3: Nice to Have

#### 7. **Slack MCP Server** (Team Communication)

**Why You Need It:**
- Send notifications from Claude
- Post deployment updates
- Share code snippets
- Team collaboration

**Use Cases for Budget App:**
- Post deployment notifications
- Share bug reports
- Notify team of releases
- Request code reviews

**Installation:**
```bash
claude mcp add slack --scope user
```

**Configuration:**
```json
{
  "mcpServers": {
    "slack": {
      "transport": {
        "type": "http",
        "url": "https://slack.com/api/mcp"
      },
      "env": {
        "SLACK_TOKEN": "xoxb-your-token"
      }
    }
  }
}
```

---

#### 8. **AWS MCP Server** (Cloud Deployment)

**Why You Need It:**
- Deploy to AWS
- Manage cloud resources
- Monitor services
- Access cloud databases

**Use Cases for Budget App:**
- Deploy to AWS ECS/Lambda
- Manage RDS databases
- Monitor CloudWatch logs
- Manage S3 backups

**Installation:**
```bash
claude mcp add aws --scope user
```

**Configuration:**
```json
{
  "mcpServers": {
    "aws": {
      "transport": {
        "type": "http",
        "url": "https://aws.amazon.com/mcp"
      },
      "env": {
        "AWS_ACCESS_KEY_ID": "your-key",
        "AWS_SECRET_ACCESS_KEY": "your-secret",
        "AWS_REGION": "us-east-1"
      }
    }
  }
}
```

---

#### 9. **Vercel MCP Server** (Deployment Platform)

**Why You Need It:**
- Deploy frontend/fullstack apps
- Manage deployments
- View deployment logs
- Environment configuration

**Use Cases for Budget App:**
- Deploy to Vercel
- View deployment status
- Check build logs
- Manage environment variables
- Preview deployments

**Installation:**
```bash
claude mcp add vercel --scope user
```

---

#### 10. **Web Search MCP** (Research & Documentation)

**Why You Need It:**
- Search for solutions
- Find documentation
- Research best practices
- Stay updated on technologies

**Use Cases for Budget App:**
- Search for Go patterns
- Find SQLite optimization tips
- Research zero-based budgeting
- Look up API design patterns
- Find error solutions

**Installation:**
```bash
npm install -g @modelcontextprotocol/server-brave-search
claude mcp add web-search --scope user
```

**Configuration:**
```json
{
  "mcpServers": {
    "web-search": {
      "transport": {
        "type": "stdio",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-brave-search"]
      },
      "env": {
        "BRAVE_API_KEY": "your-brave-api-key"
      }
    }
  }
}
```

---

## Priority Levels

### 🔴 Priority 1: Install Now
- **GitHub**: Essential for code management
- **PostgreSQL**: Important for future production use
- **Sentry**: Critical for production monitoring

### 🟡 Priority 2: Install Soon
- **Filesystem**: Better file operations
- **Docker**: Container management
- **Git**: Enhanced version control

### 🔵 Priority 3: Optional
- **Slack**: Team communication
- **AWS**: If deploying to AWS
- **Vercel**: If using Vercel
- **Web Search**: Research assistance

---

## Installation Guide

### Basic Installation Steps

1. **Install MCP Server** (if needed):
   ```bash
   npm install -g @modelcontextprotocol/server-name
   ```

2. **Add to Claude Code**:
   ```bash
   claude mcp add server-name --scope user
   ```

3. **Configure** (edit `~/.claude/mcp.json`):
   ```json
   {
     "mcpServers": {
       "server-name": {
         "transport": { ... },
         "env": { ... }
       }
     }
   }
   ```

4. **Verify**:
   ```bash
   claude mcp list
   ```

### Remote vs Local MCP Servers

**Remote Servers (HTTP):**
- Easier to set up (just add URL)
- No local installation needed
- Managed by vendor
- Require authentication

**Local Servers (stdio):**
- Run on your machine
- More control
- No network latency
- Require npm installation

**Recommendation:** Prefer remote servers when available (less maintenance).

---

## Configuration Examples

### Complete MCP Configuration

`~/.claude/mcp.json`:

```json
{
  "mcpServers": {
    "github": {
      "transport": {
        "type": "http",
        "url": "https://api.github.com/mcp"
      },
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "sentry": {
      "transport": {
        "type": "http",
        "url": "https://sentry.io/api/0/mcp"
      },
      "env": {
        "SENTRY_AUTH_TOKEN": "${SENTRY_AUTH_TOKEN}",
        "SENTRY_ORG": "your-org",
        "SENTRY_PROJECT": "budget"
      }
    },
    "postgres": {
      "transport": {
        "type": "stdio",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-postgres", "${DATABASE_URL}"]
      }
    },
    "filesystem": {
      "transport": {
        "type": "stdio",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/budget"]
      }
    },
    "docker": {
      "transport": {
        "type": "stdio",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-docker"]
      }
    }
  }
}
```

### Environment Variables

Create `.env` file (never commit!):
```bash
GITHUB_TOKEN=ghp_your_token_here
SENTRY_AUTH_TOKEN=your_sentry_token
DATABASE_URL=postgresql://localhost/budget
BRAVE_API_KEY=your_brave_key
```

---

## Security Best Practices

### ✅ DO:
- Use environment variables for tokens
- Keep tokens in `.env` (add to `.gitignore`)
- Use read-only tokens when possible
- Rotate tokens regularly
- Use separate tokens for different environments

### ❌ DON'T:
- Commit tokens to Git
- Share tokens in documentation
- Use admin tokens when read-only is sufficient
- Store tokens in plain text configs

---

## Testing MCP Servers

After installation, test each MCP:

```bash
# List configured servers
claude mcp list

# Test in Claude Code
# Just ask Claude to use the MCP:
# "Check GitHub for open PRs"
# "Query the database for total accounts"
# "Show me recent Sentry errors"
```

---

## Budget App Specific Workflows

### Workflow 1: Deploy and Monitor

```
With GitHub + Sentry + Vercel MCPs:

1. "Create a PR for my changes"
2. "Deploy to Vercel"
3. "Monitor Sentry for errors after deployment"
4. "If errors found, create GitHub issue with details"
```

### Workflow 2: Database Operations

```
With PostgreSQL MCP:

1. "Check production database schema"
2. "Compare with local SQLite schema"
3. "Generate migration plan"
4. "Test migration queries"
```

### Workflow 3: Code Review

```
With GitHub + Filesystem MCPs:

1. "List open PRs"
2. "Review PR #123 with focus on budget logic"
3. "Check test coverage for changed files"
4. "Add review comments"
```

---

## Maintenance

### Regular Tasks

- **Weekly**: Review MCP usage and performance
- **Monthly**: Check for MCP updates
- **Quarterly**: Audit token permissions
- **Yearly**: Rotate authentication tokens

### Troubleshooting

**MCP not responding:**
```bash
# Check MCP status
claude mcp list

# Restart MCP
claude mcp remove problem-mcp
claude mcp add problem-mcp --scope user
```

**Connection errors:**
- Check network connectivity
- Verify authentication tokens
- Check MCP server status
- Review logs

---

## Getting Started

### Quick Start (Priority 1 Only)

1. **Install GitHub MCP:**
   ```bash
   claude mcp add github --scope user
   ```

2. **Configure with your token:**
   Edit `~/.claude/mcp.json` and add GitHub token

3. **Test:**
   Ask Claude: "List my GitHub repositories"

4. **Repeat for Sentry and PostgreSQL** when ready

---

## Resources

- **Official MCP Documentation**: https://docs.claude.com/en/docs/claude-code/mcp
- **MCP Community**: https://www.claudemcp.com/
- **MCP Server List**: https://github.com/modelcontextprotocol/servers
- **Budget App MCPs**: https://mcpcat.io/

---

## Next Steps

1. ✅ Install Priority 1 MCPs (GitHub, Sentry)
2. ✅ Configure authentication tokens
3. ✅ Test each MCP with simple queries
4. ✅ Install Priority 2 MCPs as needed
5. ✅ Create custom workflows using MCPs
6. ✅ Document your MCP usage patterns

---

*This document is a living guide. Update it as you install MCPs and discover new use cases for the Budget application.*
