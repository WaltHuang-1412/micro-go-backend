# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## MCP Usage Requirements

**MANDATORY: Always use MCP analysis before any development task**
- All feature development must start with MCP analysis of existing code
- All debugging must begin with MCP examination of relevant components  
- Use MCP to understand code patterns before making changes
- Available MCP tools: `analyze_go_files`, `generate_tests`, `analyze_docker_logs`, `check_api_health`

## Common Development Commands

```bash
# Development with hot reload
docker compose up --build

# Production deployment
docker compose -f docker-compose.prod.yml up --build -d

# Reset database (development only)
docker compose down -v && docker compose up --build

# Generate Swagger documentation (MUST run after adding new endpoints)
swag init

# Run specific tests
go test ./handlers -v
go test ./models -v

# Create new migration
migrate create -ext sql -dir migrations migration_name

# Database migration (manual if needed)
docker run --rm -v $(pwd)/migrations:/migrations --network micro-go-backend_default migrate/migrate -path=/migrations -database="mysql://user:password@tcp(mysql:3306)/go_backend" up
```

## Architecture Overview

**Layered Architecture Pattern:**
- `config/` - Environment-based configuration management
- `handlers/` - HTTP controllers with dependency injection pattern  
- `models/` - Database operations and business logic
- `routes/` - Route registration and middleware attachment
- `middlewares/` - JWT authentication, CORS, and cross-cutting concerns
- `migrations/` - Sequential database schema versioning

**Key Technologies:**
- **Gin framework** for HTTP routing and middleware
- **MySQL 8.0** with prepared statements for security
- **JWT (HS256)** for stateless authentication with 72h expiration
- **bcrypt** for password hashing
- **Swagger/OpenAPI** for API documentation

## Database Architecture

**Schema Relationships:**
- Users (1) → Sections (many) → Tasks (many)
- All data isolated per user via `user_id` foreign keys
- Cascade deletion: Section deletion removes associated tasks
- Sort ordering support for drag-and-drop functionality

**Transaction Patterns:**
- Complex operations use explicit transactions
- Batch updates for drag-and-drop reordering
- Connection retry logic with exponential backoff

## Authentication System

**JWT Implementation:**
- Factory pattern: handlers accept `*sql.DB` dependencies
- Bearer token validation in Authorization header
- User ownership verification on all protected endpoints
- Claims include `user_id`, `username`, and expiration

**Security Measures:**
- User data isolation through user_id checks
- SQL injection prevention via prepared statements
- Password strength handled by bcrypt default cost
- CORS middleware with configurable origins

## API Structure

**Base Path:** `/api/v1`
**Route Patterns:**
- Public: `/register`, `/login`  
- Protected: `/profile`, `/plans/sections/*`, `/plans/tasks/*`
- Complex: `/plans/sections-with-tasks` for hierarchical operations

**Response Patterns:**
- Consistent JSON error responses with HTTP status codes
- Swagger documentation for all endpoints
- Input validation using Gin's ShouldBindJSON

## Configuration Management

**Environment Variables:**
- Centralized config struct with nested types (DB, Server, Swagger, Email)
- Development vs production configuration separation
- JWT secret management (avoid hardcoded defaults)
- Database DSN generation with proper escaping

**CRITICAL: External Service Configuration:**
- **Email services**: Leave SMTP_HOST empty in development to enable dev mode
- **Never use fake-but-real-looking credentials** (e.g., "your-email@gmail.com") 
- **Always validate configuration on startup** and fail fast with clear error messages
- **Example safe dev config**:
  ```env
  # Development - triggers dev mode
  SMTP_HOST=
  SMTP_USERNAME=
  
  # Production - real credentials
  SMTP_HOST=smtp.gmail.com
  SMTP_USERNAME=real@gmail.com
  ```

## Development Workflow

**Before Code Changes:**
1. Use MCP to analyze relevant code components
2. Review existing patterns in similar handlers/models
3. Check current test coverage and patterns
4. Verify configuration requirements

**After Analysis - File Planning:**
After MCP analysis, always provide a clear breakdown of files to be modified or created, organized by folder:

```
Files to Update:
├── config/
│   └── config.go (add email configuration)
├── handlers/  
│   ├── auth.go (add forgot password endpoints)
│   └── auth_test.go (update)
├── models/
│   ├── password_reset.go (new)
│   └── user.go (add email validation)
└── migrations/
    └── 003_password_resets.up.sql (new)
```

**Implementation Order:**
List implementation sequence with dependencies and rationale for each step.

**Code Patterns to Follow:**
- Handler factory functions that accept database connections
- User ownership validation in all protected endpoints  
- Proper error handling with contextual messages
- Transaction usage for multi-table operations
- Consistent logging and error response formats

**MANDATORY: After Adding New API Endpoints:**
1. **ALWAYS run `swag init`** to regenerate Swagger documentation
2. Verify new endpoints appear in Swagger UI at http://localhost:8088/swagger/index.html
3. Check that request/response schemas are correctly documented
4. Update API documentation comments with proper Swagger annotations

**CRITICAL: External Service Integration (Email, SMS, etc.):**
1. **ALWAYS implement development mode first** - never start with real external services
2. **Use empty config values to trigger dev mode** - avoid fake credentials that look real
3. **Add detailed error logging at each step** - use fmt.Printf for debugging, structured logging for production
4. **Test incrementally**: Models → Services → Handlers → Integration
5. **Provide dev tools** - create `/dev/` endpoints to assist with testing (remove in production)

## Testing Strategy

**Current Pattern:**
- HTTP testing using `gin.TestMode` and `net/http/httptest`
- Mock database connections for unit tests
- Test file naming: `*_test.go` alongside source files

## Docker Environment

**Development:**
- Hot reload with Air for rapid iteration
- Volume mounts for live code changes
- Database persistence with named volumes

**Production:**
- Multi-stage builds with compiled binaries
- No volume mounts for security
- Restart policies for service recovery
- Automated migrations on startup

## Security Considerations

**Authentication:**
- Never expose password hashes in responses
- Validate JWT tokens on all protected routes
- Implement user ownership checks consistently

**Database:**
- Use prepared statements exclusively
- Implement proper transaction rollback on errors
- Validate all foreign key relationships

**Configuration:**
- Store sensitive values in environment variables
- Never commit secrets to repository
- Use strong JWT secrets in production

## Error Handling & Debugging

**Development Error Handling:**
- **Add step-by-step logging** with fmt.Printf for complex operations
- **Log successful operations** to confirm flow progression
- **Include error context** - which step failed and with what values
- **Example pattern**:
  ```go
  fmt.Printf("✅ User found: ID=%d, Email=%s\n", user.ID, user.Email)
  passwordReset, error := models.CreatePasswordReset(database, user.ID)
  if error != nil {
      fmt.Printf("🚨 CreatePasswordReset error: %v\n", error)
      context.JSON(500, gin.H{"error": "Failed to create reset token"})
      return
  }
  fmt.Printf("✅ Token created: %s\n", passwordReset.Token)
  ```

**Production Error Handling:**
- Replace fmt.Printf with structured logging (slog)
- Remove debug endpoints and development-only code
- Ensure sensitive information is not logged