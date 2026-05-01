# AgileOS API Documentation (Swagger/OpenAPI)

## 📚 Overview

AgileOS Backend sekarang dilengkapi dengan **Swagger UI** - dokumentasi API interaktif yang memungkinkan developer untuk:
- Melihat semua endpoint yang tersedia
- Memahami request/response schema
- **Test API langsung dari browser** dengan JWT authentication
- Generate API client untuk TypeScript/JavaScript

## 🚀 Akses Swagger UI

### Local Development
```
http://localhost:8080/swagger/index.html
```

### Production (Azure)
```
https://your-azure-app.azurewebsites.net/swagger/index.html
```

## 🔐 Cara Menggunakan Swagger UI dengan Authentication

### Step 1: Login untuk mendapatkan JWT Token
1. Buka Swagger UI
2. Scroll ke section **Auth**
3. Klik endpoint `POST /api/v1/auth/login`
4. Klik tombol **"Try it out"**
5. Masukkan credentials:
```json
{
  "username": "testuser",
  "password": "test123456"
}
```
6. Klik **"Execute"**
7. Copy `access_token` dari response

### Step 2: Authorize Swagger dengan Token
1. Klik tombol **"Authorize"** di bagian atas Swagger UI (ikon gembok)
2. Masukkan token dengan format:
```
Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```
3. Klik **"Authorize"**
4. Klik **"Close"**

### Step 3: Test Protected Endpoints
Sekarang semua endpoint yang memerlukan authentication akan otomatis include Authorization header!

Contoh endpoint yang bisa dicoba:
- `GET /api/v1/auth/profile` - Get user profile
- `GET /api/v1/workflows` - List all workflows
- `POST /api/v1/workflow` - Create new workflow (admin only)
- `GET /api/v1/tasks/pending/{assignedTo}` - Get pending tasks

## 📖 API Categories

### 🔑 Auth
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/refresh` - Refresh access token
- `GET /api/v1/auth/profile` - Get user profile (protected)

### 🔄 BPM Engine
- `GET /api/v1/workflows` - List workflows
- `GET /api/v1/workflow/:id` - Get workflow details
- `POST /api/v1/workflow` - Create workflow (admin only)
- `POST /api/v1/process/start` - Start process instance
- `GET /api/v1/tasks/pending/:assignedTo` - Get pending tasks
- `POST /api/v1/task/:id/complete` - Complete task

### 📊 Analytics
- `GET /api/v1/analytics/overview` - Get analytics overview
- `GET /api/v1/analytics/workflow/:id` - Get workflow analytics

### 🏥 Health
- `GET /health` - Overall health check
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe

## 🔧 Generate TypeScript API Client

### Option 1: Using swagger-typescript-api
```bash
npm install -g swagger-typescript-api

# Generate from running server
swagger-typescript-api -p http://localhost:8080/swagger/doc.json -o ./src/api -n api.ts

# Or from local file
swagger-typescript-api -p ./backend-go/docs/swagger.json -o ./frontend-next/lib/generated -n api.ts
```

### Option 2: Using openapi-generator
```bash
npm install @openapitools/openapi-generator-cli -g

openapi-generator-cli generate \
  -i http://localhost:8080/swagger/doc.json \
  -g typescript-axios \
  -o ./frontend-next/lib/generated
```

### Benefits of Generated Client:
- ✅ **Type-safe** - Full TypeScript types for all requests/responses
- ✅ **Auto-complete** - IDE suggestions for all API methods
- ✅ **Validation** - Compile-time checks for API calls
- ✅ **Consistency** - Single source of truth from backend

## 📝 Adding Documentation to New Endpoints

### Example: Document a new handler

```go
// GetWorkflows retrieves all workflows
// @Summary List all workflows
// @Description Get a list of all available workflows in the system
// @Tags BPM Engine
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(active, inactive)
// @Success 200 {array} models.Workflow "List of workflows"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /workflows [get]
func (h *WorkflowHandler) GetWorkflows(c *gin.Context) {
    // Handler implementation
}
```

### Regenerate docs after adding annotations:
```bash
cd backend-go
swag init -g main.go --output ./docs
go build
```

## 🌐 Deployment to Azure

### Ensure Swagger is accessible in production:

1. **Update main.go host annotation** for production:
```go
// @host your-app.azurewebsites.net
// @BasePath /api/v1
```

2. **Regenerate docs**:
```bash
swag init -g main.go --output ./docs
```

3. **Deploy to Azure**:
```bash
docker build -t agileos-backend .
docker tag agileos-backend your-registry.azurecr.io/agileos-backend:latest
docker push your-registry.azurecr.io/agileos-backend:latest
```

4. **Access Swagger UI**:
```
https://your-app.azurewebsites.net/swagger/index.html
```

## 🎯 Best Practices

### 1. Always Document:
- Summary: Short description (1 line)
- Description: Detailed explanation
- Tags: Group related endpoints
- Parameters: All query, path, and body params
- Responses: All possible status codes

### 2. Use Proper Types:
- Reference models: `{object} models.User`
- Arrays: `{array} models.Workflow`
- Primitives: `{string}`, `{integer}`, `{boolean}`

### 3. Security:
- Add `@Security BearerAuth` to protected endpoints
- Document authentication requirements clearly

### 4. Examples:
- Provide example request/response bodies
- Show common error scenarios

## 📚 Resources

- [Swag Documentation](https://github.com/swaggo/swag)
- [OpenAPI 3.0 Specification](https://swagger.io/specification/)
- [Swagger UI](https://swagger.io/tools/swagger-ui/)

## ✅ Verification Checklist

- [ ] Swagger UI accessible at `/swagger/index.html`
- [ ] All endpoints documented with annotations
- [ ] JWT Bearer authentication configured
- [ ] Request/response schemas visible
- [ ] Can test endpoints directly from Swagger UI
- [ ] Production deployment includes Swagger docs

---

**Status**: ✅ Swagger Integration Complete

**Next Steps**:
1. Restart backend: `.\run-local.ps1`
2. Open browser: `http://localhost:8080/swagger/index.html`
3. Test API endpoints with JWT authentication
4. Generate TypeScript client for frontend (optional)
