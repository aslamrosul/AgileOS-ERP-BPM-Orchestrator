# SurrealDB v1.4 Upgrade Summary - AgileOS

## ✅ Yang Sudah Diupdate

### 1. Docker Image
- **File**: `docker-compose.yml`
- **Perubahan**: `surrealdb/surrealdb:latest` → `surrealdb/surrealdb:v1.4.2`
- **Status**: ✅ DONE

### 2. Database Schema v1.4
- **File**: `backend-go/database/schema-v1.4.surql`
- **Fitur Baru**:
  - ✅ `TYPE RELATION` untuk tabel `next` (graph edges)
  - ✅ `TYPE NORMAL` untuk semua tabel regular
  - ✅ `IF NOT EXISTS` untuk semua DEFINE statements
  - ✅ `FLEXIBLE` fields untuk dynamic data
  - ✅ `READONLY` fields untuk audit trails
  - ✅ Composite indexes
  - ✅ Full-text search dengan BM25
  - ✅ Automated audit trail dengan EVENTS
  - ✅ Custom functions (fn::is_admin, fn::can_approve, dll)
  - ✅ Granular permissions per tabel
- **Status**: ✅ DONE

### 3. Seed Data v1.4
- **File**: `backend-go/database/seed-v1.4.surql`
- **Isi**:
  - ✅ 5 default users
  - ✅ Purchase approval workflow
  - ✅ 6 workflow steps dengan graph relations
  - ✅ Sample process instance dan task
- **Status**: ✅ DONE

### 4. Deployment Script
- **File**: `backend-go/scripts/apply-schema-v1.4.ps1`
- **Fitur**:
  - ✅ Automated schema deployment
  - ✅ Connectivity check
  - ✅ Optional seed data
  - ✅ Schema verification
  - ✅ Statistics display
- **Status**: ✅ DONE (dengan perbaikan syntax)

### 5. Dokumentasi
- **File**: `SURREALDB-V1.4-IMPLEMENTATION.md`
- **Isi**: Panduan lengkap implementasi v1.4
- **Status**: ✅ DONE

## ⚠️ Yang Belum Diupdate (Perlu Dilakukan)

### 1. Go Backend Code
**File yang perlu diupdate**:
- `backend-go/database/surreal.go` - Belum menggunakan fitur v1.4
- `backend-go/models/*.go` - Belum ada model untuk audit_trails, workflow_versions

**Yang perlu ditambahkan**:
```go
// Model untuk audit trails
type AuditTrail struct {
    ID           string                 `json:"id,omitempty"`
    ActorID      string                 `json:"actor_id"`
    Action       string                 `json:"action"`
    ResourceType string                 `json:"resource_type"`
    ResourceID   string                 `json:"resource_id"`
    OldValue     map[string]interface{} `json:"old_value,omitempty"`
    NewValue     map[string]interface{} `json:"new_value,omitempty"`
    Timestamp    time.Time              `json:"timestamp"`
    IPAddress    string                 `json:"ip_address,omitempty"`
    UserAgent    string                 `json:"user_agent,omitempty"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Model untuk workflow versions
type WorkflowVersion struct {
    ID           string                 `json:"id,omitempty"`
    WorkflowID   string                 `json:"workflow_id"`
    Version      int                    `json:"version"`
    Definition   map[string]interface{} `json:"definition"`
    CreatedBy    string                 `json:"created_by"`
    CreatedAt    time.Time              `json:"created_at"`
    ChangeNotes  string                 `json:"change_notes,omitempty"`
}
```

### 2. Graph Traversal Functions
**File**: `backend-go/database/surreal.go`

**Yang perlu ditambahkan**:
```go
// GetNextStepsGraph - menggunakan graph traversal v1.4
func (s *SurrealDB) GetNextStepsGraph(stepID string) ([]Step, error) {
    query := `SELECT ->next->step.* AS next_steps FROM $step`
    
    result, err := s.client.Query(query, map[string]interface{}{
        "step": stepID,
    })
    if err != nil {
        return nil, err
    }
    
    // Parse result...
    return steps, nil
}

// GetPreviousStepsGraph - backward traversal
func (s *SurrealDB) GetPreviousStepsGraph(stepID string) ([]Step, error) {
    query := `SELECT <-next<-step.* AS previous_steps FROM $step`
    // ...
}
```

### 3. Custom Functions Usage
**File**: `backend-go/handlers/*.go`

**Yang perlu ditambahkan**:
```go
// Gunakan custom functions di query
func (h *Handler) CheckUserPermission(userID, taskID string) (bool, error) {
    query := `SELECT fn::can_approve($user_id, $task_id) AS can_approve`
    
    result, err := h.db.Query(query, map[string]interface{}{
        "user_id": userID,
        "task_id": taskID,
    })
    // ...
}
```

### 4. Full-Text Search
**File**: `backend-go/handlers/workflow.go`

**Yang perlu ditambahkan**:
```go
func (h *Handler) SearchWorkflows(searchTerm string) ([]Workflow, error) {
    query := `
        SELECT *, search::score(1) AS score
        FROM workflow
        WHERE name @1@ $term OR description @1@ $term
        ORDER BY score DESC
        LIMIT 20
    `
    
    result, err := h.db.Query(query, map[string]interface{}{
        "term": searchTerm,
    })
    // ...
}
```

### 5. Audit Trail Handler
**File**: `backend-go/handlers/audit.go` (BARU)

**Yang perlu dibuat**:
```go
package handlers

type AuditHandler struct {
    db *database.SurrealDB
}

func (h *AuditHandler) GetAuditTrails(c *gin.Context) {
    // Query audit trails dengan filter
    query := `
        SELECT * FROM audit_trails
        WHERE actor_id = $actor_id
        ORDER BY timestamp DESC
        LIMIT 100
    `
    // ...
}

func (h *AuditHandler) GetResourceHistory(c *gin.Context) {
    // Get history untuk resource tertentu
    query := `
        SELECT * FROM audit_trails
        WHERE resource_type = $type AND resource_id = $id
        ORDER BY timestamp DESC
    `
    // ...
}
```

## 🚀 Langkah-Langkah Upgrade

### Step 1: Update Docker Image
```bash
# Pull SurrealDB v1.4.2
docker pull surrealdb/surrealdb:v1.4.2

# Restart dengan image baru
docker-compose down
docker-compose up -d agileos-db
```

### Step 2: Apply Schema v1.4
```powershell
# Apply schema
.\backend-go\scripts\apply-schema-v1.4.ps1

# Pilih 'y' untuk seed data
```

### Step 3: Update Go Code (Manual)
1. Tambahkan model baru (AuditTrail, WorkflowVersion)
2. Update database functions untuk graph traversal
3. Tambahkan handler untuk audit trails
4. Implementasi full-text search
5. Gunakan custom functions

### Step 4: Test
```bash
# Test graph traversal
curl -X POST http://localhost:8002/sql \
  -H "NS: agileos" -H "DB: main" \
  -u root:root \
  -d "SELECT ->next->step.* FROM step:manager_approval;"

# Test custom function
curl -X POST http://localhost:8002/sql \
  -H "NS: agileos" -H "DB: main" \
  -u root:root \
  -d "SELECT fn::is_admin('admin') AS is_admin;"

# Test audit trails
curl -X POST http://localhost:8002/sql \
  -H "NS: agileos" -H "DB: main" \
  -u root:root \
  -d "SELECT * FROM audit_trails ORDER BY timestamp DESC LIMIT 10;"
```

## 📊 Perbandingan Schema Lama vs Baru

| Fitur | Schema Lama | Schema v1.4 |
|-------|-------------|-------------|
| Table Types | Implicit | Explicit (TYPE NORMAL/RELATION) |
| Graph Edges | Manual RELATE | TYPE RELATION dengan IN/OUT |
| Idempotency | ❌ Error jika sudah ada | ✅ IF NOT EXISTS |
| Audit Trail | Manual di aplikasi | ✅ Automated dengan EVENTS |
| Custom Functions | ❌ Tidak ada | ✅ fn::is_admin, fn::can_approve |
| Full-Text Search | ❌ Tidak ada | ✅ BM25 analyzer |
| Permissions | Basic | ✅ Granular per tabel |
| Dynamic Fields | ❌ Rigid | ✅ FLEXIBLE TYPE |
| Readonly Fields | ❌ Tidak ada | ✅ READONLY |

## 🎯 Keuntungan Upgrade ke v1.4

1. **Type Safety**: TYPE RELATION memastikan graph edges valid
2. **Idempotent Schema**: Bisa apply schema berulang kali tanpa error
3. **Automated Audit**: Semua perubahan tercatat otomatis
4. **Better Performance**: Composite indexes dan full-text search
5. **Cleaner Code**: Custom functions mengurangi duplikasi logic
6. **Better Security**: Granular permissions per tabel
7. **Flexibility**: FLEXIBLE fields untuk data dinamis

## ⚠️ Breaking Changes

### 1. Graph Edge Table
**Lama**:
```sql
CREATE next CONTENT { from: step:a, to: step:b };
```

**Baru**:
```sql
RELATE step:a->next->step:b SET condition = {...};
```

### 2. Email Validation
**Lama**:
```sql
ASSERT string::is::email($value)
```

**Baru**:
```sql
ASSERT string::is_email($value) = true
```

### 3. FLEXIBLE Syntax
**Lama**:
```sql
DEFINE FIELD metadata FLEXIBLE TYPE object;
```

**Baru**:
```sql
DEFINE FIELD metadata TYPE object FLEXIBLE;
```

## 📝 TODO List

- [ ] Update Go models (AuditTrail, WorkflowVersion)
- [ ] Implement graph traversal functions
- [ ] Add audit trail handler
- [ ] Implement full-text search
- [ ] Use custom functions in queries
- [ ] Update frontend untuk menampilkan audit trail
- [ ] Add tests untuk fitur v1.4
- [ ] Update documentation

## 🔗 Resources

- [SurrealDB v1.4 Release Notes](https://surrealdb.com/releases/1.4.0)
- [TYPE RELATION Documentation](https://surrealdb.com/docs/surrealql/statements/define/table)
- [IF NOT EXISTS Documentation](https://surrealdb.com/docs/surrealql/statements/define)
- [Custom Functions](https://surrealdb.com/docs/surrealql/statements/define/function)
- [Full-Text Search](https://surrealdb.com/docs/surrealql/statements/define/indexes)

---

**Status**: Schema v1.4 sudah siap, Go backend perlu update
**Last Updated**: 2026-04-29
**Version**: 1.4.2
