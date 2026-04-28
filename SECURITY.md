# AgileOS BPM - Security Documentation

## Authentication & Authorization

### Overview

AgileOS BPM implements enterprise-grade security using:
- **JWT (JSON Web Tokens)** for stateless authentication
- **bcrypt** for password hashing
- **Role-Based Access Control (RBAC)** for authorization
- **Middleware-based** protection for API endpoints

---

## User Roles

### Available Roles

| Role | Description | Permissions |
|------|-------------|-------------|
| **admin** | System Administrator | Full access to all features |
| **manager** | Department Manager | Create processes, approve tasks, view reports |
| **employee** | Regular Employee | Submit requests, view own tasks |
| **finance** | Finance Team | Approve financial workflows |
| **procurement** | Procurement Team | Handle procurement workflows |

### Role Hierarchy

```
admin (highest)
  ↓
manager
  ↓
finance / procurement
  ↓
employee (lowest)
```

---

## API Authentication

### Login

**Endpoint:** `POST /api/v1/auth/login`

**Request:**
```json
{
  "username": "admin",
  "password": "password123"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "user:admin",
    "username": "admin",
    "email": "admin@agileos.com",
    "role": "admin",
    "full_name": "System Administrator",
    "department": "IT"
  }
}
```

### Using Access Token

Include the access token in the `Authorization` header:

```http
GET /api/v1/auth/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### Token Expiry

- **Access Token**: 24 hours
- **Refresh Token**: 7 days

### Refresh Token

**Endpoint:** `POST /api/v1/auth/refresh`

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

---

## Protected Endpoints

### Public Endpoints (No Authentication)

- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/refresh` - Token refresh
- `GET /health` - Health check

### Protected Endpoints (Authentication Required)

#### All Authenticated Users

- `GET /api/v1/auth/profile` - Get current user profile
- `GET /api/v1/workflows` - List workflows
- `GET /api/v1/workflow/:id` - Get workflow details
- `GET /api/v1/tasks/pending/:assignedTo` - Get pending tasks
- `POST /api/v1/task/:id/complete` - Complete task

#### Manager & Admin Only

- `POST /api/v1/process/start` - Start new process

#### Admin Only

- `POST /api/v1/workflow` - Create workflow
- `GET /api/v1/users` - List all users

---

## Password Security

### Hashing

Passwords are hashed using **bcrypt** with cost factor 12:

```go
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
```

### Password Requirements

- Minimum 8 characters
- Recommended: Mix of uppercase, lowercase, numbers, and symbols

### Generating Password Hash

Use the utility script:

```bash
cd backend-go
go run scripts/hash-password.go "mySecurePassword123"
```

Output:
```
Password hashed successfully!
---
Plain text: mySecurePassword123
Hashed:     $2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIeWU7u3oi
---
Use this hash in your database or seed script.
```

---

## Creating Users

### Via API (Registration)

**Endpoint:** `POST /api/v1/auth/register`

```json
{
  "username": "newuser",
  "email": "newuser@agileos.com",
  "password": "securePassword123",
  "full_name": "New User",
  "department": "Sales",
  "role": "employee"
}
```

### Via Database (Manual)

1. Hash the password:
```bash
go run scripts/hash-password.go "password123"
```

2. Insert into SurrealDB:
```sql
CREATE user:newuser SET
    username = "newuser",
    email = "newuser@agileos.com",
    password_hash = "$2a$12$...",
    role = "employee",
    full_name = "New User",
    department = "Sales",
    is_active = true,
    created_at = time::now(),
    updated_at = time::now();
```

### Seed Default Users

Run the seed script:

```bash
# Open http://localhost:8002
# Copy and paste content from: backend-go/database/seed-users.surql
```

**Default Users:**
- Username: `admin` / Password: `password123` (Role: admin)
- Username: `manager` / Password: `password123` (Role: manager)
- Username: `employee` / Password: `password123` (Role: employee)
- Username: `finance` / Password: `password123` (Role: finance)
- Username: `procurement` / Password: `password123` (Role: procurement)

---

## Frontend Integration

### Login Example

```typescript
import { login, setTokens, setUser } from '@/lib/auth';

const handleLogin = async () => {
  try {
    const response = await login('admin', 'password123');
    // Tokens and user info are automatically stored
    router.push('/dashboard');
  } catch (error) {
    console.error('Login failed:', error);
  }
};
```

### Authenticated API Calls

```typescript
import { authenticatedFetch } from '@/lib/auth';

const fetchWorkflows = async () => {
  const response = await authenticatedFetch(
    `${process.env.NEXT_PUBLIC_API_URL}/api/v1/workflows`
  );
  const data = await response.json();
  return data;
};
```

### Role-Based UI

```typescript
import { hasRole, isAdmin } from '@/lib/auth';

// Show admin-only button
{isAdmin() && (
  <button onClick={createWorkflow}>Create Workflow</button>
)}

// Check multiple roles
{hasRole(['admin', 'manager']) && (
  <button onClick={startProcess}>Start Process</button>
)}
```

---

## Security Best Practices

### JWT Secret

**CRITICAL:** Change the JWT secret in production!

```bash
# Generate a secure random secret
openssl rand -base64 32

# Set in .env
JWT_SECRET=your-generated-secret-here
```

### HTTPS

Always use HTTPS in production:

```nginx
server {
    listen 443 ssl http2;
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    # ... rest of config
}
```

### CORS

Configure CORS properly in production:

```go
r.Use(func(c *gin.Context) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", "https://your-domain.com")
    c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    // ...
})
```

### Rate Limiting

Implement rate limiting for auth endpoints:

```nginx
limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=5r/m;

location /api/v1/auth/ {
    limit_req zone=auth_limit burst=10 nodelay;
    # ...
}
```

### Password Policy

Enforce strong passwords:
- Minimum 12 characters
- Mix of character types
- No common passwords
- Password expiry (optional)

### Session Management

- Implement token blacklist for logout
- Monitor suspicious login attempts
- Log all authentication events

---

## Testing Authentication

### Run Test Script

```powershell
cd backend-go
.\scripts\test-auth.ps1
```

### Manual Testing

```bash
# Login
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}'

# Get profile (use token from login response)
curl -X GET http://localhost:8081/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Access admin endpoint
curl -X GET http://localhost:8081/api/v1/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## Troubleshooting

### "Invalid or expired token"

- Token might be expired (24 hours)
- Use refresh token to get new access token
- Check JWT_SECRET matches between requests

### "Access denied"

- User doesn't have required role
- Check user role in database
- Verify middleware configuration

### "User not found"

- User doesn't exist in database
- Run seed script to create default users
- Check database connection

---

## Security Checklist

- [ ] Change JWT_SECRET in production
- [ ] Enable HTTPS
- [ ] Configure CORS properly
- [ ] Implement rate limiting
- [ ] Set up monitoring and alerts
- [ ] Regular security audits
- [ ] Keep dependencies updated
- [ ] Implement password policy
- [ ] Enable audit logging
- [ ] Set up backup and recovery

---

**Last Updated**: April 2026
**Version**: 1.0.0
