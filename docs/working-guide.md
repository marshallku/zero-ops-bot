# Working Guide

Guidelines for developing this project with Claude.

## Commits

Use `~/dev/save.sh` for quick commits:

```bash
~/dev/save.sh "commit message"
```

This script stages all changes, commits, and pushes in one command.

**When to commit:**
- After completing each small task/todo item
- After adding a new feature or command
- After updating documentation
- Before switching to a different task

**Commit message style:**
- Start with verb: Add, Fix, Update, Remove, Refactor
- Keep it concise but descriptive
- Examples:
  - `Add /infra command`
  - `Fix nil pointer in message handler`
  - `Update environment variable docs`

## Documentation

### Required Documentation

1. **CLAUDE.md** (project root)
   - Quick reference for the project
   - Key decisions and architecture overview
   - Environment variables table
   - Common commands

2. **docs/architecture.md**
   - Design decisions with rationale
   - Package structure explanation
   - Data flow diagrams
   - Security considerations

3. **README.md**
   - User-facing setup instructions
   - Feature list
   - Quick start guide

### When to Update Docs

- Adding new features → Update README.md feature list
- Changing architecture → Update docs/architecture.md
- Adding environment variables → Update CLAUDE.md table
- Making design decisions → Document rationale in architecture.md

## Profiles

Developer profiles are stored in `~/.claude/profile/`. These define coding patterns and preferences.

### Relevant Profiles for Go

| File | When to Read |
|------|--------------|
| `120-architecture.md` | Project structure, layered architecture |
| `130-common-patterns.md` | Error handling patterns, early returns |
| `220-error-handling.md` | Error handling philosophy |
| `910-anti-patterns.md` | Things to avoid |

### Key Patterns to Follow

1. **Early Returns**
   ```go
   // Good
   if err != nil {
       return nil, err
   }
   // continue with happy path

   // Bad
   if err == nil {
       // nested logic
   }
   ```

2. **Error Wrapping**
   ```go
   return nil, fmt.Errorf("create session: %w", err)
   ```

3. **Layered Architecture**
   ```
   handlers/    → Request handling, validation
   services/    → Business logic, external calls
   config/      → Environment, configuration
   ```

4. **File Naming**
   - Use snake_case for files: `check_health.go`
   - One main concept per file
   - Keep files focused and small

## Adding New Commands

1. Create command file in `internal/commands/`:
   ```go
   // internal/commands/my_command.go
   var MyCommand = &Command{
       Definition: &discordgo.ApplicationCommand{
           Name:        "my-command",
           Description: "Description here",
       },
       Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate, n8n *services.N8nClient) {
           // Implementation
       },
   }
   ```

2. Register in `internal/commands/commands.go`:
   ```go
   func init() {
       Register(CheckHealth)
       Register(Infra)
       Register(MyCommand)  // Add this
   }
   ```

3. Update README.md feature list

4. Commit:
   ```bash
   ~/dev/save.sh "Add /my-command command"
   ```

## Workflow Summary

```
1. Understand task
2. Create todo list (if multiple steps)
3. Implement
4. Verify build: go build ./cmd/bot
5. Update docs if needed
6. Commit: ~/dev/save.sh "message"
7. Mark todo complete
8. Repeat for next task
```
