# Database Layer - AgileOS BPM

Repository pattern implementation untuk SurrealDB dengan fokus pada BPM workflow management.

## Struktur

```
database/
├── surreal.go      # SurrealDB client & CRUD operations
├── seed.surql      # Sample data untuk testing
└── README.md       # Dokumentasi ini
```

## Fitur Utama

### 1. Workflow Management
- Create/Update workflow definitions
- Version control untuk workflows
- Active/inactive status

### 2. Step Management
- Define workflow steps dengan tipe berbeda (approval, action, decision, notify)
- SLA tracking per step
- Flexible configuration per step

### 3. Graph Relations
- Menggunakan SurrealDB RELATE untuk menghubungkan steps
- Support conditional branching (approved/rejected paths)
- Graph traversal untuk mendapatkan next steps

### 4. Process Instance
- Track running workflow instances
- Execution history logging
- Process variables (data)

## Cara Menggunakan

### 1. Seed Database

Buka SurrealDB Dashboard di `http://localhost:8000` atau gunakan CLI:

```bash
surreal sql --conn http://localhost:8000 --user root --pass root --ns agileos --db main --file seed.surql
```

### 2. Verifikasi Data

Di SurrealDB Dashboard, jalankan query:

```sql
-- Lihat semua workflows
SELECT * FROM workflow;

-- Lihat semua steps untuk workflow tertentu
SELECT * FROM step WHERE workflow_id = "workflow:purchase_approval";

-- Lihat graph relationships
SELECT id, name, ->next->step.name AS next_steps 
FROM step 
WHERE workflow_id = "workflow:purchase_approval";

-- Test graph traversal
SELECT ->next->step.* AS next_steps FROM step:pr_submit;
```

### 3. Gunakan di Go Code

```go
package main

import (
    "agileos-backend/database"
    "agileos-backend/models"
    "log"
    "time"
)

func main() {
    // Connect to database
    db, err := database.ConnectDB(
        "ws://localhost:8000/rpc",
        "root",
        "root",
        "agileos",
        "main",
    )
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create workflow
    wf := &models.Workflow{
        Name:        "Leave Request",
        Version:     "1.0.0",
        Description: "Employee leave request workflow",
        IsActive:    true,
    }
    db.SaveWorkflow(wf)

    // Add steps
    step1 := &models.Step{
        WorkflowID:  wf.ID,
        Name:        "Submit Request",
        Type:        models.StepTypeAction,
        AssignedTo:  "role:employee",
        SLA:         1 * time.Hour,
        Description: "Employee submits leave request",
    }
    db.AddStep(step1)

    step2 := &models.Step{
        WorkflowID:  wf.ID,
        Name:        "Manager Approval",
        Type:        models.StepTypeApproval,
        AssignedTo:  "role:manager",
        SLA:         24 * time.Hour,
        Description: "Manager approves leave",
    }
    db.AddStep(step2)

    // Link steps
    db.LinkSteps(step1.ID, step2.ID, nil)

    // Get next step
    nextSteps, _ := db.GetNextStep(step1.ID)
    log.Printf("Next steps: %+v", nextSteps)
}
```

## Verifikasi Relasi

### Method 1: SurrealDB Dashboard

1. Buka `http://localhost:8000`
2. Login dengan `root` / `root`
3. Pilih namespace `agileos` dan database `main`
4. Jalankan query:

```sql
SELECT id, name, ->next->step.name AS next_steps FROM step;
```

### Method 2: Via Go Code

```go
nextSteps, err := db.GetNextStep("step:pr_submit")
if err != nil {
    log.Fatal(err)
}

for _, step := range nextSteps {
    log.Printf("Next: %s (%s)", step.Name, step.Type)
}
```

### Method 3: cURL ke Backend API

```bash
curl http://localhost:8080/api/v1/steps/step:pr_submit/next
```

## Graph Visualization

Workflow graph untuk Purchase Approval:

```
[Submit PR] 
    ↓
[Manager Review]
    ↓ (approved)        ↓ (rejected)
[Finance Approval]      [Rejected]
    ↓ (approved)        
[Procurement]
    ↓
[Completed]
```

## Best Practices

1. Selalu gunakan transaction untuk operasi multi-step
2. Validate workflow sebelum activate
3. Track execution history untuk audit trail
4. Implement retry mechanism untuk failed steps
5. Use SLA monitoring untuk alerting
