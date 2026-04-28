# Backend Architecture - AgileOS BPM Engine

## Overview

Backend AgileOS dibangun dengan Go menggunakan clean architecture dan repository pattern untuk memastikan scalability dan maintainability.

## Layer Architecture

```
┌─────────────────────────────────────────┐
│         API Layer (Gin Router)          │
│         - REST Endpoints                │
│         - Request Validation            │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│         Service Layer (BPM Logic)       │
│         - Workflow Orchestration        │
│         - Business Rules                │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│      Repository Layer (Database)        │
│      - CRUD Operations                  │
│      - Graph Queries                    │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│         Data Layer (SurrealDB)          │
│         - Multi-model Database          │
│         - Graph Relations               │
└─────────────────────────────────────────┘
```

## Directory Structure

```
backend-go/
├── main.go              # Entry point
├── models/              # Data models
│   └── workflow.go      # Workflow, Step, ProcessInstance
├── database/            # Repository layer
│   ├── surreal.go       # SurrealDB client & operations
│   ├── seed.surql       # Sample data
│   └── README.md        # Database documentation
├── scripts/             # Utility scripts
│   ├── test-db.ps1      # Test database connection
│   └── seed-db.ps1      # Seed database
├── Dockerfile           # Container image
├── go.mod               # Dependencies
└── run-local.ps1        # Local development runner
```

## Data Models

### Workflow
Represents a business process definition.

```go
type Workflow struct {
    ID          string
    Name        string
    Version     string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    IsActive    bool
}
```

### Step
Represents a single step in a workflow.

```go
type Step struct {
    ID          string
    WorkflowID  string
    Name        string
    Type        StepType      // approval, action, decision, notify
    AssignedTo  string        // User/Role ID
    SLA         time.Duration // Max duration
    Description string
    Config      interface{}   // Flexible config
    CreatedAt   time.Time
}
```

### ProcessInstance
Represents a running instance of a workflow.

```go
type ProcessInstance struct {
    ID              string
    WorkflowID      string
    CurrentStepID   string
    Status          ProcessStatus
    StartedAt       time.Time
    CompletedAt     *time.Time
    InitiatedBy     string
    Data            map[string]interface{}
    ExecutionHistory []ExecutionLog
}
```

## Graph Relations

SurrealDB RELATE digunakan untuk menghubungkan steps:

```sql
RELATE step:A->next->step:B;
```

Ini memungkinkan graph traversal untuk BPM orchestration:

```sql
SELECT ->next->step.* AS next_steps FROM step:current;
```

## Repository Pattern

### Interface (Future)

```go
type WorkflowRepository interface {
    SaveWorkflow(wf *Workflow) error
    GetWorkflow(id string) (*Workflow, error)
    AddStep(step *Step) error
    LinkSteps(from, to string, condition map[string]interface{}) error
    GetNextStep(currentID string) ([]Step, error)
}
```

### Implementation

`database/surreal.go` implements all CRUD operations:

- `SaveWorkflow()` - Create/update workflow
- `AddStep()` - Add step to workflow
- `LinkSteps()` - Create NEXT relation
- `GetNextStep()` - Graph traversal untuk BPM
- `CreateProcessInstance()` - Start workflow instance
- `UpdateProcessInstance()` - Update instance state

## BPM Core Logic

### Workflow Execution Flow

1. **Create Workflow Definition**
   ```go
   wf := &Workflow{Name: "Purchase Approval", Version: "1.0.0"}
   db.SaveWorkflow(wf)
   ```

2. **Add Steps**
   ```go
   step1 := &Step{WorkflowID: wf.ID, Name: "Submit", Type: StepTypeAction}
   db.AddStep(step1)
   ```

3. **Link Steps (Graph)**
   ```go
   db.LinkSteps(step1.ID, step2.ID, nil)
   ```

4. **Start Process Instance**
   ```go
   instance := &ProcessInstance{
       WorkflowID: wf.ID,
       CurrentStepID: step1.ID,
       InitiatedBy: "user:john",
   }
   db.CreateProcessInstance(instance)
   ```

5. **Get Next Step (Orchestration)**
   ```go
   nextSteps, _ := db.GetNextStep(instance.CurrentStepID)
   ```

## Testing

### 1. Test Database Connection

```bash
.\scripts\test-db.ps1
```

### 2. Seed Sample Data

```bash
.\scripts\seed-db.ps1
```

Atau manual via SurrealDB Dashboard:
1. Open `http://localhost:8000`
2. Login: `root` / `root`
3. Copy paste content dari `database/seed.surql`
4. Execute

### 3. Verify Data

```sql
-- List workflows
SELECT * FROM workflow;

-- List steps with next steps
SELECT id, name, ->next->step.name AS next_steps FROM step;

-- Test graph traversal
SELECT ->next->step.* AS next_steps FROM step:pr_submit;
```

## Next Steps (Prompt 3)

- [ ] REST API endpoints untuk workflow management
- [ ] BPM orchestration service layer
- [ ] Authentication & authorization
- [ ] NATS integration untuk async processing
- [ ] Webhook support untuk external integrations

## Performance Considerations

1. **Graph Queries**: SurrealDB graph traversal sangat cepat untuk BPM
2. **Indexing**: Index pada `workflow_id` dan `status` untuk query performance
3. **Connection Pooling**: Reuse SurrealDB connection
4. **Caching**: Future - cache workflow definitions

## Security

1. **Input Validation**: Validate semua input di API layer
2. **SQL Injection**: SurrealDB parameterized queries
3. **Authentication**: JWT tokens (coming in next prompt)
4. **Authorization**: Role-based access control (RBAC)
