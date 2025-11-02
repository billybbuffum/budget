# Bootstrap Claude Code Configuration for Any Project

You are tasked with generating a comprehensive Claude Code configuration including sub agents, slash commands, skills, and MCP recommendations.

**This command supports two modes:**
1. **Empty/New Project Mode**: User provides description (e.g., "I'm building an ovulation tracker app using React Native and Firebase")
2. **Existing Project Mode**: Analyze existing codebase to determine configuration

## Phase 0: Determine Project Mode and Gather Requirements

### 0.1 Check Project State

**Determine if project is empty or has existing code:**
```bash
# Count source files
find . -type f \( -name "*.go" -o -name "*.js" -o -name "*.ts" -o -name "*.tsx" -o -name "*.jsx" -o -name "*.py" -o -name "*.java" -o -name "*.rb" -o -name "*.php" -o -name "*.rs" -o -name "*.swift" -o -name "*.kt" -o -name "*.c" -o -name "*.cpp" -o -name "*.cs" \) | wc -l

# Check for package/dependency files
ls -la | grep -E "(package.json|go.mod|requirements.txt|Gemfile|composer.json|Cargo.toml|pom.xml|build.gradle|pubspec.yaml)"

# List project structure
ls -R | head -50
```

**Decision criteria:**
- **Empty project** (0-5 source files): Use Mode 1 (Interactive Requirements Gathering)
- **Existing project** (6+ source files): Use Mode 2 (Codebase Analysis)

---

## MODE 1: Empty/New Project - Interactive Requirements Gathering

**Use this mode when the user provided a description OR project has minimal/no code.**

### 1.1 Extract Information from User's Description

**If user provided description** (e.g., "I'm building an ovulation tracker app using React Native and Firebase"):

Extract:
- **Project type**: Mobile app, web app, API, CLI tool, library, etc.
- **Domain**: Ovulation tracking, e-commerce, fintech, healthcare, productivity, etc.
- **Tech stack mentioned**: React Native, Firebase, etc.
- **Key features implied**: Cycle tracking, predictions, notifications, etc.

### 1.2 Ask Clarifying Questions (Interactive)

**Ask user to complete the picture:**

**1. Technology Stack:**
```
I'll help you set up Claude Code for your {domain} {project_type}!

Let me ask a few questions to generate the best configuration:

**Technology Stack:**
- Primary language: {detected or ask: "What language? (JavaScript/TypeScript, Python, Go, Java, etc.)"}
- Framework/platform: {detected or ask: "What framework? (React Native, Next.js, Django, Spring Boot, etc.)"}
- Database: {ask: "What database? (Firebase/Firestore, PostgreSQL, MongoDB, SQLite, etc.)"}
- Additional tools: {ask: "Any other tools/services? (Auth0, Stripe, AWS, etc.)"}
```

**2. Domain and Features:**
```
**Domain Understanding:**
- Primary domain: {detected from description}
- Key features: {ask: "What are the core features? (e.g., cycle tracking, symptom logging, predictions, notifications)"}
- Key entities: {ask: "Main data entities? (e.g., User, Cycle, Symptom, Prediction)"}
- Business rules: {ask: "Any critical business logic? (e.g., ovulation prediction algorithm, data privacy requirements)"}
```

**3. Architecture and Testing:**
```
**Architecture:**
- Architecture preference: {ask: "Preferred architecture? (Clean Architecture, MVC, Feature-based, etc.) [or 'recommend based on tech stack']"}
- Testing approach: {ask: "Testing strategy? (Unit + Integration, E2E with Playwright, etc.) [or 'recommend']"}
```

**4. Development Workflow:**
```
**Workflow:**
- Team size: {ask: "Solo or team project?"}
- Deployment: {ask: "Deployment target? (App Store, Google Play, web hosting, etc.)"}
- CI/CD: {ask: "Want CI/CD setup? (GitHub Actions, GitLab CI, etc.)"}
```

### 1.3 Determine Configuration Based on Answers

**Based on responses, determine:**

**Tech Stack Profile:**
- Languages: [e.g., TypeScript, JavaScript]
- Frameworks: [e.g., React Native, Expo]
- Backend: [e.g., Firebase Functions]
- Database: [e.g., Firestore]
- Services: [e.g., Firebase Auth, Cloud Messaging]
- Testing: [e.g., Jest, React Native Testing Library]
- Deployment: [e.g., App Store, Google Play]

**Domain Profile:**
- Domain: Health tracking / Women's health
- Subdomain: Ovulation and fertility tracking
- Key Concepts: Cycle, ovulation, fertile window, symptoms, predictions
- Regulatory: HIPAA considerations, data privacy
- Algorithms: Ovulation prediction, pattern recognition

**Architecture Profile:**
- Pattern: Feature-based (common for React Native)
- Layers: UI components, business logic, data layer
- State management: Redux/Zustand/Context
- Navigation: React Navigation

### 1.4 Skip to Phase 2 (Generation)

With requirements gathered, proceed directly to **Phase 2** to generate configuration.

---

## MODE 2: Existing Project - Codebase Analysis

**Use this mode when project has substantial existing code (6+ source files).**

### 2.1 Analyze Codebase Structure

**Discover technology stack:**
```bash
# Find all file types to understand tech stack
find . -type f \( -name "*.go" -o -name "*.js" -o -name "*.ts" -o -name "*.tsx" -o -name "*.jsx" -o -name "*.py" -o -name "*.java" -o -name "*.rb" -o -name "*.php" -o -name "*.rs" -o -name "*.swift" -o -name "*.kt" \) | head -100

# Check for package/dependency files and read them
cat package.json 2>/dev/null || true
cat go.mod 2>/dev/null || true
cat requirements.txt 2>/dev/null || true
cat Cargo.toml 2>/dev/null || true
cat pom.xml 2>/dev/null || true

# Examine project structure
tree -L 3 -I 'node_modules|vendor|venv' || ls -R | head -150
```

**Identify frameworks and tools:**
- Read package.json, go.mod, requirements.txt, etc.
- Look for framework-specific files (next.config.js, django settings, etc.)
- Identify testing frameworks
- Identify build tools
- Identify database technologies

**Analyze architecture patterns:**
- Look for architecture indicators (clean architecture, MVC, microservices, monolith)
- Examine directory structure (domain/, application/, infrastructure/, features/, etc.)
- Check for design patterns in code

### 2.2 Analyze Domain and Business Logic

**Understand the business domain:**
- Read README.md, docs/, specifications
- Analyze main domain entities and models
- Identify core business rules and logic
- Understand the problem space (e-commerce, fintech, healthcare, SaaS, etc.)

**Search for domain-specific concepts:**
- Look for domain terminology in code comments
- Examine API endpoints and function names
- Review test files for business scenarios
- Check for domain documentation

### 2.3 Identify Current Development Workflows

**Examine existing practices:**
- Check for CI/CD configuration (.github/workflows, .gitlab-ci.yml, etc.)
- Review testing setup and patterns
- Look for code quality tools (linters, formatters)
- Identify deployment practices
- Check for existing documentation generation

---

## Phase 2: Generate Sub Agents

**Based on analysis (Mode 1 or Mode 2), create sub agents in `.claude/agents/` directory.**

### 2.1 Always Create These Core Agents:

**1. code-reviewer**
- Language-specific best practices
- Framework-specific patterns
- Architecture compliance checks
- Code quality verification
- Performance considerations

**Template customization:**
- For JavaScript/TypeScript: ESLint rules, React patterns, async/await patterns
- For Python: PEP 8, type hints, Django/Flask patterns
- For Go: Go idioms, error handling, goroutine safety
- For Java: SOLID principles, Spring patterns, exception handling
- For mobile: Platform-specific guidelines (iOS/Android), performance optimization

**2. test-generator**
- Framework-specific test patterns
- Unit test generation
- Integration test generation
- Test coverage strategies
- Mock/stub patterns

**Template customization:**
- For React/React Native: Jest + React Testing Library, component testing, hook testing
- For Python: pytest patterns, fixtures, parametrized tests
- For Go: table-driven tests, test suites, mock interfaces
- For Java: JUnit 5, Mockito, integration tests
- For mobile: UI testing, snapshot testing, E2E testing

**3. security-auditor**
- Language-specific vulnerabilities (SQL injection, XSS, etc.)
- Framework security best practices
- Authentication/authorization checks
- Data validation
- Secrets management
- Domain-specific security (PCI-DSS for payments, HIPAA for health data, etc.)

**4. refactoring-assistant**
- Safe refactoring strategies
- Architecture preservation
- Breaking change detection
- Migration assistance
- Code smell detection

**5. documentation-generator**
- API documentation (OpenAPI, JSDoc, GoDoc, JavaDoc, etc.)
- Architecture documentation
- README generation
- Code comments
- User guides (for user-facing apps)

### 2.2 Create Domain-Specific Agent:

**{domain}-expert** (e.g., ovulation-tracker-expert, ecommerce-expert, fintech-expert)

**Include domain-specific knowledge:**

**For ovulation tracking app:**
- Cycle length validation (21-35 days typical)
- Ovulation prediction algorithms (temperature method, calendar method, symptom-based)
- Fertile window calculation (5 days before + day of ovulation)
- Data privacy requirements (HIPAA, GDPR for health data)
- Symptom categorization (basal body temperature, cervical mucus, mood, etc.)
- Prediction accuracy and confidence levels
- Historical data analysis for pattern recognition

**For e-commerce:**
- Inventory management (stock levels, reservations, backorders)
- Pricing logic (discounts, promotions, taxes)
- Cart abandonment
- Order workflow (pending → paid → shipped → delivered)
- Payment processing (idempotency, refunds)
- Product catalog (variants, attributes, categories)

**For fintech:**
- Transaction consistency (ACID properties)
- Double-entry bookkeeping
- Currency handling (avoid floating point)
- Regulatory compliance (KYC, AML)
- Account reconciliation
- Audit trails

**For SaaS:**
- Multi-tenancy (data isolation)
- Subscription management (trials, billing, cancellations)
- Feature flags
- Usage-based pricing
- Onboarding workflows

**For healthcare:**
- HIPAA compliance (PHI protection, access logs, encryption)
- Patient consent management
- Clinical data validation
- Interoperability standards (FHIR, HL7)
- Audit trails

### 2.3 Create Conditional Agents Based on Tech Stack:

**If API/backend project:**
- **api-designer** - RESTful/GraphQL API design, endpoint structure, versioning

**If frontend/mobile project:**
- **ui-tester** - Interactive UI testing with Playwright MCP (web) or Appium (mobile)
- **accessibility-auditor** - WCAG compliance, a11y checks, screen reader testing
- **ux-optimizer** - UX best practices, navigation flows, user feedback

**If database-heavy:**
- **database-optimizer** - Query optimization, schema design, indexing strategies

**If has deployment config:**
- **deployment-helper** - CI/CD, containerization, cloud deployment, app store submission

**If mobile app:**
- **platform-specialist** - iOS/Android platform-specific guidelines, performance, app store requirements

**If uses specific services:**
- **firebase-specialist** (if Firebase)
- **aws-specialist** (if AWS)
- **stripe-specialist** (if payments with Stripe)

---

## Phase 3: Generate Slash Commands

**Create slash commands in `.claude/commands/` directory.**

### 3.1 Always Create These Commands:

**1. /new-feature** `<feature-name>`
- Scaffold new feature following project architecture
- Create necessary files/folders
- Generate boilerplate code
- Update routing/navigation if applicable

**Customization:**
- React Native: Create feature folder with components, hooks, screens, types
- Go clean architecture: Create domain, application, infrastructure layers
- Django: Create app with models, views, serializers, URLs
- Next.js: Create page, API route, components

**2. /review-pr**
- Comprehensive PR review
- Architecture compliance
- Code quality checks
- Security validation
- Test coverage verification

**3. /run-tests**
- Execute test suite
- Report results
- Identify failing tests
- Show coverage report

**Customization:**
- React Native: `npm test` or `yarn test`
- Go: `go test ./...`
- Python: `pytest` or `python -m pytest`
- Java: `mvn test` or `gradle test`

**4. /check-architecture**
- Verify architecture compliance
- Check layer boundaries
- Validate dependency direction
- Detect architectural violations

### 3.2 Spec-Driven Development Commands:

**5. /create-spec** `<feature-description>`
- Create validated specification
- Domain expert review
- Security review
- Technical design
- Test plan

**6. /implement-spec** `<spec-url>`
- Implement validated specification
- Code review loop
- Testing loop
- Quality gates

### 3.3 Tech Stack Specific Commands:

**If has API (backend):**
- **/test-endpoint** `<path>` - Test API endpoints with curl/Postman
- **/new-endpoint** `<description>` - Add new API endpoint
- **/generate-api-docs** - Generate OpenAPI/Swagger documentation

**If has frontend/mobile UI:**
- **/test-ui** `<workflow>` - Interactive UI testing with Playwright/Appium
- **/check-accessibility** - Check WCAG compliance
- **/test-navigation** - Test navigation flows

**If has database:**
- **/migrate-db** - Run database migrations
- **/optimize-query** `<query>` - Analyze and optimize database queries
- **/seed-db** - Seed database with test data

**If has deployment:**
- **/deploy** `<environment>` - Deploy to environment (dev/staging/prod)
- **/check-deployment** - Verify deployment health
- **/rollback** - Rollback to previous version

**If mobile app:**
- **/build-ios** - Build iOS app
- **/build-android** - Build Android app
- **/submit-app-store** - Prepare app store submission
- **/test-on-device** - Test on physical device

**If uses CI/CD:**
- **/fix-ci** - Diagnose and fix CI/CD failures
- **/update-ci** - Update CI/CD configuration

---

## Phase 4: Generate Skills

**Create skills in `.claude/skills/` directory.**

### 4.1 Architecture Skill:

**{architecture-pattern}-architecture**

**Examples:**
- **clean-architecture** (for Go, Java, C#)
- **mvc-architecture** (for Rails, Django)
- **feature-based-architecture** (for React Native, Flutter)
- **microservices-architecture** (for distributed systems)
- **jamstack-architecture** (for Next.js, Gatsby)

**Include:**
- Layer patterns and separation
- Dependency rules
- Design patterns
- File organization
- Component communication

### 4.2 Domain Skill:

**{domain-name}**

**Examples for ovulation tracker:**
- **ovulation-tracking** or **fertility-tracking**

**Content:**
```markdown
# Ovulation Tracking Domain Knowledge

## Triggers
- Discussing menstrual cycles
- Implementing period/ovulation prediction
- Validating cycle data
- Designing symptom tracking
- Implementing fertility algorithms

## Core Concepts

### Menstrual Cycle
- **Cycle length**: Typically 21-35 days (28 days average)
- **Phases**:
  - Menstrual phase (days 1-5): Period
  - Follicular phase (days 1-13): Egg development
  - Ovulation (day 14): Egg release
  - Luteal phase (days 15-28): Prepare for pregnancy or period

### Ovulation Detection Methods

**1. Calendar Method:**
- Ovulation typically occurs 14 days before next period
- Fertile window: 5 days before + day of ovulation
- Requires regular cycles for accuracy

**2. Basal Body Temperature (BBT):**
- Temperature drops before ovulation, rises after
- Requires daily morning temperature measurement
- Increase of 0.5-1°F indicates ovulation occurred

**3. Cervical Mucus Method:**
- Fertile: Clear, stretchy, egg-white consistency
- Non-fertile: Thick, sticky, or absent
- Peak fertility at "egg-white" mucus

**4. Symptom-Thermal Method:**
- Combines BBT + cervical mucus + other symptoms
- Most accurate non-medical method

**5. Ovulation Predictor Kits (OPK):**
- Detects LH surge (24-36 hours before ovulation)
- Most accurate home method

### Key Formulas

**Ovulation Prediction (Calendar Method):**
```
next_ovulation_date = next_period_start_date - 14 days
fertile_window_start = next_ovulation_date - 5 days
fertile_window_end = next_ovulation_date
```

**Average Cycle Length:**
```
average_cycle_length = sum(last_6_cycle_lengths) / 6
```

**Next Period Prediction:**
```
next_period_date = last_period_start + average_cycle_length
```

**Prediction Confidence:**
- Regular cycles (±2 days variation): High confidence
- Irregular cycles (>3 days variation): Low confidence
- Need 3+ cycles for initial predictions
- Need 6+ cycles for accurate predictions

### Data Validation Rules

**Cycle Length:**
- Minimum: 21 days
- Maximum: 35 days
- Flag if <21 or >35 (possible irregularity)
- Flag if varies >7 days between cycles

**Period Duration:**
- Typical: 3-7 days
- Flag if <2 or >8 days

**BBT Temperature:**
- Normal range: 97.0-99.0°F (36.1-37.2°C)
- Pre-ovulation: 97.0-97.5°F
- Post-ovulation: 97.6-98.6°F

### Privacy and Compliance

**HIPAA Considerations:**
- Menstrual data is Protected Health Information (PHI)
- Requires encryption at rest and in transit
- Access logs required
- User consent for data collection
- Right to data export and deletion

**GDPR Compliance:**
- Health data is "special category"
- Explicit consent required
- Purpose limitation
- Data minimization
- Right to erasure

### Prediction Algorithms

**Simple Calendar Method:**
```python
def predict_ovulation(cycles: List[Cycle]) -> Prediction:
    if len(cycles) < 3:
        return Prediction(confidence="low", message="Need more data")

    avg_length = sum(c.length for c in cycles[-6:]) / min(len(cycles), 6)
    std_dev = calculate_std_dev(cycles[-6:])

    last_period_start = cycles[-1].start_date
    next_period = last_period_start + timedelta(days=avg_length)
    ovulation_date = next_period - timedelta(days=14)

    confidence = "high" if std_dev <= 2 else "medium" if std_dev <= 4 else "low"

    return Prediction(
        ovulation_date=ovulation_date,
        fertile_window=(ovulation_date - timedelta(days=5), ovulation_date),
        confidence=confidence
    )
```

### Common Mistakes to Avoid

❌ Using floating-point for cycle day calculations
✅ Use integer days and date arithmetic

❌ Assuming 28-day cycles for everyone
✅ Calculate individual average cycle length

❌ Predicting ovulation without enough data
✅ Require minimum 3 cycles, recommend 6+

❌ Storing sensitive health data without encryption
✅ Encrypt at rest, use HTTPS, implement access controls

❌ Hard-coding ovulation day as "day 14"
✅ Calculate based on next period - 14 days (works for irregular cycles)

### Testing Checklist

- [ ] Validate cycle length boundaries (21-35 days)
- [ ] Handle irregular cycles gracefully
- [ ] Test prediction accuracy with real cycle data
- [ ] Verify privacy controls (encryption, access logs)
- [ ] Test edge cases (first cycle, very irregular cycles)
- [ ] Validate symptom data (BBT ranges, mucus types)
- [ ] Test prediction confidence levels
- [ ] Verify data export functionality
```

**Other domain examples:**

- **ecommerce** - Inventory, pricing, cart, checkout
- **fintech** - Transactions, accounts, compliance
- **saas** - Multi-tenancy, subscriptions, billing
- **healthcare** - HIPAA, clinical data, patient management

### 4.3 Testing Skill:

**{language}-testing**

**Examples:**
- **javascript-testing** (Jest, React Testing Library, Playwright)
- **python-testing** (pytest, unittest, mocking)
- **go-testing** (table-driven tests, test suites)
- **java-testing** (JUnit, Mockito, integration tests)

### 4.4 Framework/Database Skills:

**Framework skills:**
- **react-native-patterns** (for React Native apps)
- **nextjs-patterns** (for Next.js)
- **django-patterns** (for Django)
- **spring-boot-patterns** (for Spring Boot)

**Database skills:**
- **firestore-best-practices** (for Firebase/Firestore)
- **postgres-best-practices** (for PostgreSQL)
- **mongodb-patterns** (for MongoDB)
- **sqlite-best-practices** (for SQLite)

---

## Phase 5: Generate MCP Recommendations

**Create `.claude/MCP_RECOMMENDATIONS.md` with prioritized recommendations.**

### 5.1 Priority 1 MCPs (Essential for Tech Stack):

**For React Native + Firebase project (ovulation tracker example):**

**1. Playwright MCP (if also has web version)**
- Interactive browser testing
- Installation: `npm install -g @playwright/test @modelcontextprotocol/server-playwright`

**2. Firebase MCP**
- Firestore operations
- Authentication testing
- Cloud Functions deployment

**3. React Native Debugger MCP (if available)**
- Mobile debugging
- Component inspection

**4. npm MCP**
- Package management
- Dependency updates

### 5.2 Tech Stack Specific MCPs:

**If Node.js/JavaScript:**
- npm MCP - Package management
- ESLint MCP - Linting

**If Python:**
- pip MCP - Package management
- pytest MCP - Testing

**If Go:**
- Go modules MCP - Dependency management

**If mobile (iOS/Android):**
- Xcode MCP (iOS)
- Android Studio MCP (Android)
- Appium MCP (mobile testing)

**If database:**
- PostgreSQL MCP
- MongoDB MCP
- Firebase MCP

**If cloud:**
- AWS MCP (for AWS)
- Google Cloud MCP (for GCP)
- Azure MCP (for Azure)

**If payments:**
- Stripe MCP (for Stripe)

**If auth:**
- Auth0 MCP (for Auth0)

### 5.3 Domain Specific MCPs:

**For health/medical:**
- FHIR MCP (health data interoperability)
- HIPAA compliance MCP (if available)

**For e-commerce:**
- Shopify MCP
- Stripe MCP
- Inventory management MCP

---

## Phase 6: Generate Documentation

### 6.1 Create .claude/README.md

**Structure:**
```markdown
# Claude Code Configuration for {Project Name}

{Brief description of project and domain}

> This configuration was automatically generated by /bootstrap-claude-code

## Quick Start

### Spec-Driven Development (Recommended)
\`\`\`
# Create validated specification
/create-spec "Add {example feature}"

# Implement specification
/implement-spec docs/spec-{example-feature}.md
\`\`\`

### Direct Development
\`\`\`
/new-feature {example-feature}
/test-ui "{example workflow}"
/review-pr
\`\`\`

## Sub Agents

{List all generated agents with descriptions}

1. **code-reviewer** - {Language} and {framework} best practices
2. **test-generator** - {Testing framework} patterns
3. **security-auditor** - {Domain-specific security concerns}
4. **{domain}-expert** - {Domain} logic validation
5. **ui-tester** - Interactive UI testing
... (list all agents)

## Slash Commands

{List all generated commands}

**Spec-Driven Development:**
- /create-spec - Create validated specification
- /implement-spec - Implement specification

**Development:**
- /new-feature - Scaffold new feature
- /test-ui - Test UI workflows
- /run-tests - Run test suite
... (list all commands)

## Skills

{List all generated skills}

1. **{architecture}-architecture** - Architecture patterns
2. **{domain}** - Domain knowledge
3. **{language}-testing** - Testing patterns
... (list all skills)

## MCP Recommendations

See [MCP_RECOMMENDATIONS.md](MCP_RECOMMENDATIONS.md) for detailed MCP setup.

Priority 1:
- {MCP 1}
- {MCP 2}
- {MCP 3}

## Example Workflows

{Include 3-4 common workflows}

## Tech Stack

- **Language**: {language}
- **Framework**: {framework}
- **Database**: {database}
- **Platform**: {platform}
- **Testing**: {testing framework}
- **Deployment**: {deployment target}

## Domain

- **Primary domain**: {domain}
- **Key entities**: {entities}
- **Core features**: {features}

## Architecture

- **Pattern**: {architecture pattern}
- **Layers**: {layers/structure}

## Learning Path

### Week 1: Basics
- Read FEATURE_USAGE_GUIDE.md
- Try /create-spec and /implement-spec
- Create first feature

### Week 2: Advanced
- Use sub agents for code review
- Generate comprehensive tests
- Set up UI testing

### Week 3: Mastery
- Create custom workflows
- Optimize with domain expert
- Full spec-driven development
```

### 6.2 Create .claude/FEATURE_USAGE_GUIDE.md

**Include:**
- When to use sub agents vs slash commands vs skills
- Decision matrix specific to this project
- Best practices for this tech stack
- Anti-patterns to avoid
- Examples from similar projects

### 6.3 Create .claude/WORKFLOWS.md

**Include project-specific workflows:**

**For ovulation tracker app example:**

```markdown
# Development Workflows

## Spec-Driven Development

### Create New Feature with Spec

\`\`\`
/create-spec "Add BBT (Basal Body Temperature) tracking"

→ Claude gathers requirements:
  - What temperature unit? (Fahrenheit/Celsius)
  - Where to track? (New screen or add to existing?)
  - Charting requirements?
  - Integration with ovulation prediction?

→ Domain expert validates:
  - BBT ranges (97.0-99.0°F)
  - Pre/post ovulation temperatures
  - Data validation rules
  - Prediction algorithm updates

→ Security review:
  - Health data encryption
  - HIPAA compliance
  - Data export requirements

→ Generates docs/spec-bbt-tracking.md
\`\`\`

### Implement Spec

\`\`\`
/implement-spec docs/spec-bbt-tracking.md

→ Creates:
  - BBTScreen component
  - Temperature input component
  - Chart component for visualization
  - useBBT hook for data management
  - Firestore schema updates
  - BBT prediction logic
  - Tests for all components

→ Code review loop ensures quality
→ UI testing verifies user workflows
\`\`\`

## Feature Development

### Add New Symptom Type

\`\`\`
/new-feature mood-tracking

→ Scaffolds:
  - features/mood-tracking/
    ├── screens/MoodTrackingScreen.tsx
    ├── components/MoodPicker.tsx
    ├── hooks/useMoodTracking.ts
    ├── types/mood.types.ts
    └── __tests__/

→ Updates navigation
→ Adds Firestore schema
\`\`\`

## Testing Workflows

### Test Cycle Logging Flow

\`\`\`
/test-ui "logging a new cycle and viewing predictions"

→ ui-tester agent:
  1. Opens app on emulator/device
  2. Navigates to cycle logging screen
  3. Enters period start date
  4. Submits data
  5. Verifies prediction updates
  6. Checks fertile window display
  7. Generates automated test if successful
\`\`\`

### Run Full Test Suite

\`\`\`
/run-tests

→ Executes:
  - Jest unit tests
  - Component tests
  - Integration tests
  - Shows coverage report
  - Identifies failures
\`\`\`

## Validation Workflows

### Validate Ovulation Logic

\`\`\`
"Invoke ovulation-tracker-expert to review the prediction algorithm"

→ Domain expert checks:
  - Calendar method implementation
  - Fertile window calculation
  - Prediction confidence levels
  - Edge cases (irregular cycles, insufficient data)
  - Data validation rules
\`\`\`

### Security Audit

\`\`\`
"Invoke security-auditor to review health data handling"

→ Security audit:
  - Firestore security rules
  - Data encryption
  - HIPAA compliance
  - Access controls
  - Data export/deletion
\`\`\`

## Deployment Workflows

### Build for App Stores

\`\`\`
/build-ios
→ Builds iOS release
→ Generates .ipa file
→ Runs pre-submission checks

/build-android
→ Builds Android release
→ Generates .aab file
→ Runs pre-submission checks
\`\`\`

### Deploy Backend

\`\`\`
/deploy production

→ Deploys Firebase Functions
→ Updates Firestore rules
→ Runs smoke tests
→ Verifies deployment health
\`\`\`

## Common Patterns

### Add New Tracking Feature

1. `/create-spec "Add {feature} tracking"`
2. Review spec, validate with domain expert
3. `/implement-spec docs/spec-{feature}.md`
4. `/test-ui "{feature} logging workflow"`
5. `/review-pr`

### Fix Prediction Bug

1. "Invoke ovulation-tracker-expert to analyze prediction algorithm"
2. Fix identified issues
3. `/run-tests` to verify fix
4. `/test-ui "prediction accuracy"` to verify in app
5. `/review-pr`

### Improve UI/UX

1. `/test-ui "{workflow}"` to identify issues
2. Make improvements
3. `/check-accessibility` to verify a11y
4. `/test-ui "{workflow}"` to re-test
5. `/review-pr`
```

(Adapt for other project types)

---

## Phase 7: Optional - Scaffold Initial Project Structure

**If project is empty and user wants scaffolding:**

Ask user:
```
Would you like me to scaffold the initial project structure?
This will create:
- Directory structure following {architecture pattern}
- Basic configuration files ({package.json, tsconfig.json, etc.})
- Starter files (App.tsx, navigation setup, etc.)
- Testing setup
- CI/CD configuration
- README.md

[Yes/No]
```

If yes, create appropriate structure based on tech stack:

**React Native example:**
```
project/
├── src/
│   ├── features/           # Feature-based architecture
│   │   ├── auth/
│   │   ├── cycle-tracking/
│   │   └── predictions/
│   ├── shared/             # Shared components, hooks, utils
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── utils/
│   │   └── types/
│   ├── navigation/
│   ├── services/           # API, Firebase, etc.
│   └── App.tsx
├── __tests__/
├── .github/workflows/      # CI/CD
├── package.json
├── tsconfig.json
├── jest.config.js
├── .eslintrc.js
└── README.md
```

**Go clean architecture example:**
```
project/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/             # Entities and interfaces
│   ├── application/        # Use cases and services
│   └── infrastructure/     # Database, HTTP, etc.
├── pkg/                    # Public libraries
├── migrations/
├── tests/
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Phase 8: Execution Plan and Quality Standards

### 8.1 Create Comprehensive Todo List

Before starting generation, create todo list:

**For empty project mode:**
```
1. Detect project is empty
2. Parse user's description
3. Ask clarifying questions (tech stack, domain, features, architecture)
4. Determine tech stack profile
5. Determine domain profile
6. Determine architecture profile
7. Create code-reviewer agent for {language}/{framework}
8. Create test-generator agent for {testing framework}
9. Create security-auditor agent with {domain} focus
10. Create refactoring-assistant agent
11. Create documentation-generator agent
12. Create {domain}-expert agent with domain knowledge
13. Create [conditional agents based on tech stack]
14. Create /create-spec command
15. Create /implement-spec command
16. Create /new-feature command
17. Create /review-pr command
18. Create /run-tests command for {test framework}
19. Create /test-ui command for {platform}
20. Create [tech-stack specific commands]
21. Create {architecture}-architecture skill
22. Create {domain} skill with domain knowledge
23. Create {language}-testing skill
24. Create [framework/database skills]
25. Create MCP_RECOMMENDATIONS.md with {tech stack} MCPs
26. Create README.md with project overview
27. Create FEATURE_USAGE_GUIDE.md
28. Create WORKFLOWS.md with {project type} workflows
29. [Optional] Scaffold initial project structure
30. Validate all generated files
31. Commit and push changes
```

**For existing project mode:**
```
1. Analyze codebase structure
2. Identify tech stack from files
3. Analyze domain from code/docs
4. Identify architecture pattern
5. [Continue with steps 7-31 from above]
```

### 8.2 Quality Standards

**For each sub agent (500-2000 lines):**

Must include:
- Clear role and responsibilities
- Comprehensive checklists (20-50 items)
- Language/framework specific guidance
- Domain-specific considerations (if applicable)
- Code examples
- Common mistakes to avoid
- Success criteria

**For domain expert agent specifically:**
- Domain terminology glossary
- Key business rules and formulas
- Data validation rules
- Regulatory/compliance requirements
- Common domain patterns
- Anti-patterns and mistakes
- Testing checklist for domain logic

**For each slash command (300-700 lines):**

Must include:
- Clear purpose and when to use
- Phase-by-phase execution plan
- Explicit agent invocation instructions
- Quality gates and review loops
- Example usage
- Expected outcomes
- Troubleshooting section

**For each skill (400-1000 lines):**

Must include:
- Clear auto-loading triggers
- Comprehensive knowledge base
- Code patterns and examples
- Best practices
- Common mistakes
- Testing considerations
- References and resources

**For documentation:**

Must include:
- Clear quick start guide
- Complete feature listing
- Example workflows (3-5)
- Tech stack summary
- Domain summary
- Learning path
- Links to detailed guides

### 8.3 Execute Systematically

- Work through todo list sequentially
- Mark items in_progress when starting
- Mark items completed when done
- Create comprehensive, high-quality artifacts
- Use this project's `.claude/` as quality benchmark
- Test that everything works together

### 8.4 Validation

Before committing:

**Validate structure:**
```bash
# Check all required directories exist
ls .claude/agents/
ls .claude/commands/
ls .claude/skills/

# Check core files exist
ls .claude/README.md
ls .claude/FEATURE_USAGE_GUIDE.md
ls .claude/MCP_RECOMMENDATIONS.md
ls .claude/WORKFLOWS.md

# Check agent count (should have 5-10)
ls .claude/agents/ | wc -l

# Check command count (should have 6-12)
ls .claude/commands/ | wc -l

# Check skill count (should have 3-5)
ls .claude/skills/ | wc -l
```

**Validate content quality:**
- Each agent file is 500+ lines
- Each command file is 300+ lines
- Each skill file is 400+ lines
- Documentation is comprehensive
- All files use markdown formatting
- No placeholder text like "TODO" or "{fill this in}"

### 8.5 Commit and Push

```bash
git add .claude/

# If scaffolded project structure
git add src/ package.json # ... etc

git commit -m "Bootstrap Claude Code configuration for {project name}

Generated comprehensive Claude Code setup for {project type} project.

Project Details:
- Domain: {domain description}
- Tech Stack: {languages, frameworks, databases}
- Platform: {web/mobile/desktop/API/etc}
- Architecture: {architecture pattern}

Generated Configuration:
- {N} sub agents: {list key agents}
- {N} slash commands: {list key commands}
- {N} skills: {list skills}
- MCP recommendations for {tech stack}
- Complete documentation suite

Agents customized for:
- {Language} best practices
- {Framework} patterns
- {Domain} business logic
- {Platform} specific requirements

Ready to use with:
  /create-spec \"Add {example feature}\"
  /implement-spec docs/spec-{example}.md

Configuration generated via /bootstrap-claude-code
"

git push
```

---

## Success Criteria

You have successfully bootstrapped Claude Code configuration when:

### For Empty Project Mode:
✅ Gathered requirements interactively (tech stack, domain, architecture)
✅ Created domain-specific expert agent with comprehensive domain knowledge
✅ All core agents created and customized for tech stack (5+ agents)
✅ Tech-stack specific agents created (API designer, UI tester, etc.)
✅ Essential + tech-stack specific slash commands created (6+ commands)
✅ Architecture, domain, and testing skills created with proper triggers
✅ MCP recommendations tailored to tech stack
✅ Complete documentation suite created (README, FEATURE_USAGE_GUIDE, WORKFLOWS)
✅ All domain knowledge accurately captured (formulas, rules, compliance)
✅ Optional: Initial project structure scaffolded
✅ All changes committed and pushed
✅ Configuration immediately usable for development

### For Existing Project Mode:
✅ Codebase thoroughly analyzed (tech stack, domain, architecture identified)
✅ All core agents created and customized for discovered tech stack (5+ agents)
✅ Domain expert accurately reflects actual business domain
✅ Tech-stack specific agents match existing tools/frameworks
✅ Slash commands align with existing development workflows
✅ Skills match actual architecture and domain
✅ MCP recommendations fit existing tech stack
✅ Complete documentation created
✅ All changes committed and pushed
✅ Configuration ready to enhance existing development

---

## Examples of Quality Domain Experts

### Ovulation Tracker Domain Expert

See detailed example above with:
- Menstrual cycle phases
- Ovulation detection methods
- Prediction formulas
- Data validation rules
- HIPAA/privacy requirements
- Common mistakes to avoid

### E-commerce Domain Expert (Brief Example)

```markdown
## Core Concepts

### Inventory Management
- Stock levels must never go negative
- Reserved inventory (in carts) vs available
- Backorder handling

### Pricing
```
Formula: final_price = base_price × (1 - discount_percentage) + tax
```

### Order States
- PENDING → PAID → PROCESSING → SHIPPED → DELIVERED
- Or PENDING → CANCELLED

### Payment Processing
- Idempotency keys prevent duplicate charges
- Refunds must maintain audit trail
- Currency stored as integers (cents)

## Validation Rules
- SKU must be unique
- Price must be positive integer (cents)
- Inventory cannot be negative
- Order total must match items + tax + shipping
```

---

## Important Reminders

**Be Thorough:**
- Don't rush analysis or generation
- Create production-quality artifacts
- Match quality of this project's `.claude/` directory

**Be Specific:**
- Tailor everything to actual project
- Use real domain terminology
- Reference actual tech stack
- Include project-specific examples

**Be Accurate (Especially for Domain Knowledge):**
- Research domain concepts if needed
- Include accurate formulas and rules
- Document regulatory requirements
- Cite compliance standards (HIPAA, PCI-DSS, GDPR, etc.)

**Be Interactive (Empty Project Mode):**
- Ask clarifying questions
- Gather complete requirements
- Confirm understanding before generating
- Offer recommendations based on best practices

**Be Practical:**
- Focus on workflows developers will actually use
- Create tools that save time
- Automate repetitive tasks
- Provide clear value

Now analyze this project and generate comprehensive Claude Code configuration!

---

## Usage Examples

### Empty Project
```
/bootstrap-claude-code I'm building an ovulation tracker app using React Native and Firebase
```

### Existing Project
```
/bootstrap-claude-code
```

(Command will auto-detect and use appropriate mode)
