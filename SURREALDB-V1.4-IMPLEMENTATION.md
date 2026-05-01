# SurrealDB v1.4 Implementation - AgileOS

## Overview
AgileOS telah diupgrade untuk menggunakan fitur-fitur terbaru SurrealDB v1.4, termasuk TYPE RELATION untuk graph edges, IF NOT EXISTS untuk idempotent schema, dan berbagai optimasi performa.

## Fitur Baru v1.4 yang Diimplementasikan

### 1. TYPE RELATION untuk Graph Edges ✅

**Tabel `next`** sekarang menggunakan `TYPE RELATION` untuk mendefinisikan edge graph antara steps:

```sql
DEFINE TABLE next SCHEMAFULL TYPE RELATION
    IN step
    OUT step;
```

**Keuntungan**:
- Referential integrity otomatis (IN dan OUT harus valid)
- Query graph traversal lebih efisien
- Validasi tipe data pada edge

**Contoh Penggunaan**:
```sql
-- Membuat edge
RELATE step:manager_approval->next->step:finance_review SET
    condition = { status: "approved" },
    priority = 1;

-- Graph traversal
SELECT ->next->step.* AS next_steps FROM step:manager_approval;

-- Backward traversal
SELECT <-next<-step.* AS previous_steps FROM step:finance_review;
```

### 2. TYPE NORMAL untuk Regular Tables ✅

Semua tabel non-edge menggunakan `TYPE NORMAL`:

```sql
DEFINE TABLE user SCHEMAFULL TYPE NORMAL;
DEFINE TABLE workflow SCHEMAFULL TYPE NORMAL;
DEFINE TABLE task_instance SCHEMAFULL TYPE NORMAL;
```

**Keuntungan**:
- Eksplisit bahwa tabel ini bukan edge
- Mencegah penggunaan RELATE pada tabel biasa
- Dokumentasi yang lebih jelas

### 3. IF NOT EXISTS untuk Idempotent Schema ✅

Semua DEFINE statements menggunakan `IF NOT EXISTS`:

```sql
DEFINE TABLE IF NOT EXISTS user SCHEMAFULL TYPE NORMAL;
DEFINE FIELD IF NOT EXISTS username ON user TYPE string;
DEFINE INDEX IF NOT EXISTS idx_user_username ON user FIELDS username UNIQUE;
DEFINE EVENT IF NOT EXISTS evt_workflow_audit ON workflow;
DEFINE FUNCTION IF NOT EXISTS fn::is_admin($user_id: string) RETURNS bool;
```

**Keuntungan**:
- Schema dapat dijalankan berulang kali tanpa error
- Deployment yang lebih aman
- Rollback yang lebih mudah

### 4. Automated Audit Trail dengan EVENTS ✅

Event otomatis mencatat semua perubahan:

```sql
DEFINE EVENT evt_workflow_audit ON workflow
    WHEN $event IN ["CREATE", "UPDATE", "DELETE"]
    THEN {
        CREATE audit_trails SET
            actor_id = $auth.id,
            action = $event,
            resource_type = "workflow",
            resource_id = $value.id,
            old_value = $before,
            new_value = $after,
            timestamp = time::now();
    };
```

**Tabel yang Diaudit**:
- workflow
- task_instance
- process_instance

### 5. Custom Functions ✅

Fungsi bisnis logic yang dapat digunakan di query:

```sql
-- Check if user is admin
DEFINE FUNCTION fn::is_admin($user_id: string) RETURNS bool {
    LET $user = (SELECT role FROM ONLY type::thing("user", $user_id));
    RETURN $user.role = "admin";
};

-- Check if user can approve task
DEFINE FUNCTION fn::can_approve($user_id: string, $task_id: string) RETURNS bool {
    LET $task = (SELECT assigned_to, status FROM ONLY type::thing("task_instance", $task_id));
    RETURN $task.assigned_to = $user_id AND $task.status = "pending";
};

-- Get pending tasks count
DEFINE FUNCTION fn::pending_tasks_count($user_id: string) RETURNS int {
    LET $count = (SELECT count() FROM task_instance WHERE assigned_to = $user_id AND status = "pending");
    RETURN $count[0].count;
};

-- Calculate task duration
DEFINE FUNCTION fn::task_duration($task_id: string) RETURNS duration {
    LET $task = (SELECT created_at, completed_at FROM ONLY type::thing("task_instance", $task_id));
    IF $task.completed_at != NONE {
        RETURN $task.completed_at - $task.created_at;
    } ELSE {
        RETURN time::now() - $task.created_at;
    };
};
```

**Penggunaan**:
```sql
-- Check if user is admin
SELECT fn::is_admin("admin") AS is_admin;

-- Get pending tasks count
SELECT fn::pending_tasks_count("manager") AS pending_count;

-- Calculate task duration
SELECT fn::task_duration("task_001") AS duration;
```

### 6. Full-Text Search dengan BM25 ✅

Analyzer dan index untuk full-text search:

```sql
-- Define analyzers
DEFINE ANALYZER simple_analyzer
    TOKENIZERS blank
    FILTERS lowercase;

DEFINE ANALYZER english_analyzer
    TOKENIZERS blank
    FILTERS lowercase, snowball(English);

-- Full-text search indexes
DEFINE INDEX idx_workflow_fts ON workflow FIELDS name, description
    SEARCH ANALYZER simple_analyzer BM25;

DEFINE INDEX idx_task_fts ON task_instance FIELDS step_name
    SEARCH ANALYZER simple_analyzer BM25;
```

**Penggunaan**:
```sql
-- Search workflows
SELECT *, search::score(1) AS score
FROM workflow
WHERE name @1@ "purchase approval"
ORDER BY score DESC;

-- Search tasks
SELECT *, search::score(1) AS score
FROM task_instance
WHERE step_name @1@ "manager approval"
ORDER BY score DESC;
```

### 7. Improved Permissions Model ✅

Permissions yang lebih granular per tabel:

```sql
DEFINE TABLE task_instance SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE assigned_to = $auth.id OR $auth.role IN ['admin', 'manager']
        FOR create WHERE $auth.role IN ['admin', 'manager']
        FOR update WHERE assigned_to = $auth.id OR $auth.role IN ['admin', 'manager']
        FOR delete WHERE $auth.role = 'admin';
```

**Variabel yang Tersedia**:
- `$auth` - Data user yang login
- `$auth.id` - User ID
- `$auth.role` - User role
- `$value` - Nilai yang sedang diproses
- `$before` - Nilai sebelum update
- `$after` - Nilai setelah update

### 8. FLEXIBLE TYPE untuk Dynamic Fields ✅

Field yang dapat menyimpan struktur dinamis:

```sql
DEFINE FIELD metadata ON workflow FLEXIBLE TYPE object;
DEFINE FIELD config ON step FLEXIBLE TYPE object;
DEFINE FIELD data ON process_instance FLEXIBLE TYPE object;
```

**Keuntungan**:
- Nested object tanpa perlu define setiap field
- Fleksibilitas untuk data yang berubah-ubah
- Tetap type-safe pada level root

### 9. READONLY Fields ✅

Field yang tidak bisa diubah setelah dibuat:

```sql
DEFINE FIELD timestamp ON audit_trails TYPE datetime DEFAULT time::now()
    READONLY;

DEFINE FIELD created_at ON workflow_versions TYPE datetime DEFAULT time::now()
    READONLY;
```

### 10. Composite Indexes ✅

Index pada multiple columns untuk query yang lebih cepat:

```sql
-- Composite index untuk query: WHERE status = 'pending' AND assigned_to = 'manager'
DEFINE INDEX idx_task_status_assigned ON task_instance FIELDS status, assigned_to;

-- Composite index untuk audit trail
DEFINE INDEX idx_audit_actor_timestamp ON audit_trails FIELDS actor_id, timestamp;
```

## Database Schema

### Tables

| Table | Type | Description |
|-------|------|-------------|
| `user` | NORMAL | User accounts dengan role-based access |
| `workflow` | NORMAL | Workflow definitions |
| `step` | NORMAL | Workflow steps |
| `next` | RELATION | Graph edges antara steps (IN: step, OUT: step) |
| `process_instance` | NORMAL | Running workflow instances |
| `task_instance` | NORMAL | Individual tasks untuk approval/action |
| `audit_trails` | NORMAL | Immutable audit log |
| `workflow_versions` | NORMAL | Workflow version history |
| `documents` | NORMAL | Document attachments |

### Graph Structure

```
step:purchase_start
    ↓ (next)
step:manager_approval
    ↓ (next, condition: approved)
step:finance_review
    ↓ (next, condition: approved)
step:procurement_process
    ↓ (next)
step:purchase_complete
    ↓ (next)
step:purchase_end
```

## Deployment

### 1. Apply Schema

```powershell
# Apply schema v1.4
.\backend-go\scripts\apply-schema-v1.4.ps1

# With custom parameters
.\backend-go\scripts\apply-schema-v1.4.ps1 `
    -SurrealURL "http://localhost:8002" `
    -Username "root" `
    -Password "root" `
    -Namespace "agileos" `
    -Database "main"
```

### 2. Verify Schema

```sql
-- Check database info
INFO FOR DB;

-- Check table info
INFO FOR TABLE user;
INFO FOR TABLE next;

-- Verify graph edges
SELECT count() FROM next;
SELECT * FROM next;

-- Test graph traversal
SELECT ->next->step.* AS next_steps FROM step:manager_approval;
```

### 3. Test Functions

```sql
-- Test custom functions
SELECT fn::is_admin("admin") AS is_admin;
SELECT fn::pending_tasks_count("manager") AS pending;
SELECT fn::task_duration("task_001") AS duration;
```

### 4. Test Full-Text Search

```sql
-- Search workflows
SELECT *, search::score(1) AS score
FROM workflow
WHERE name @1@ "purchase"
ORDER BY score DESC;
```

## Query Examples

### Graph Traversal

```sql
-- Get next steps from current step
SELECT ->next->step.* AS next_steps 
FROM step:manager_approval;

-- Get previous steps
SELECT <-next<-step.* AS previous_steps 
FROM step:finance_review;

-- Multi-hop traversal
SELECT ->next->step->next->step.* AS two_steps_ahead
FROM step:manager_approval;

-- Get all steps in workflow (recursive)
SELECT * FROM step WHERE workflow_id = "purchase_approval";
```

### Permissions Testing

```sql
-- As manager, get my tasks
SELECT * FROM task_instance 
WHERE assigned_to = $auth.id;

-- As admin, get all tasks
SELECT * FROM task_instance;

-- Try to update task (will check permissions)
UPDATE task_instance:task_001 SET status = "completed";
```

### Audit Trail Queries

```sql
-- Get all actions by user
SELECT * FROM audit_trails 
WHERE actor_id = "admin"
ORDER BY timestamp DESC;

-- Get changes to specific resource
SELECT * FROM audit_trails 
WHERE resource_type = "workflow" 
  AND resource_id = "purchase_approval"
ORDER BY timestamp DESC;

-- Get recent audit trail (last 24 hours)
SELECT * FROM audit_trails 
WHERE timestamp > time::now() - 24h
ORDER BY timestamp DESC;
```

### Workflow Version History

```sql
-- Get all versions of workflow
SELECT * FROM workflow_versions 
WHERE workflow_id = "purchase_approval"
ORDER BY version DESC;

-- Get specific version
SELECT * FROM workflow_versions 
WHERE workflow_id = "purchase_approval" 
  AND version = 1;
```

## Performance Optimizations

### 1. Indexes

Semua query yang sering digunakan sudah dioptimasi dengan index:

- `idx_user_username` - UNIQUE index untuk login
- `idx_task_status_assigned` - Composite index untuk pending tasks
- `idx_audit_actor_timestamp` - Composite index untuk audit queries
- `idx_workflow_fts` - Full-text search index

### 2. Query Optimization

```sql
-- ❌ Slow: Full table scan
SELECT * FROM task_instance WHERE status = "pending";

-- ✅ Fast: Uses index
SELECT * FROM task_instance 
WHERE status = "pending" AND assigned_to = "manager";

-- ✅ Fast: Graph traversal dengan index
SELECT ->next->step.* FROM step:manager_approval;
```

### 3. EXPLAIN untuk Debugging

```sql
-- Check query plan
EXPLAIN SELECT * FROM task_instance 
WHERE status = "pending" AND assigned_to = "manager";
```

## Migration dari Schema Lama

Jika Anda sudah memiliki data di schema lama:

### Option 1: Fresh Install (Recommended untuk Development)

```powershell
# Stop containers
docker-compose down -v

# Start fresh
docker-compose up -d

# Apply new schema
.\backend-go\scripts\apply-schema-v1.4.ps1
```

### Option 2: Data Migration (untuk Production)

```sql
-- 1. Export data lama
-- (gunakan surreal export)

-- 2. Apply new schema
-- (jalankan apply-schema-v1.4.ps1)

-- 3. Import data dengan transformasi
-- (sesuaikan dengan struktur baru)
```

## Troubleshooting

### Error: "Table already exists"

Gunakan `IF NOT EXISTS` sudah diterapkan, tapi jika masih error:

```sql
-- Drop dan recreate
REMOVE TABLE old_table;
DEFINE TABLE IF NOT EXISTS old_table ...;
```

### Error: "Invalid relation"

Pastikan TYPE RELATION hanya digunakan untuk edge tables:

```sql
-- ❌ Wrong
DEFINE TABLE user TYPE RELATION;

-- ✅ Correct
DEFINE TABLE user TYPE NORMAL;
DEFINE TABLE next TYPE RELATION IN step OUT step;
```

### Error: "Permission denied"

Check permissions pada tabel:

```sql
-- View permissions
INFO FOR TABLE task_instance;

-- Test dengan user berbeda
-- (gunakan JWT token dengan role berbeda)
```

## Best Practices

### 1. Gunakan TYPE yang Tepat

```sql
-- Edge tables
DEFINE TABLE likes TYPE RELATION IN user OUT post;

-- Regular tables
DEFINE TABLE user TYPE NORMAL;
DEFINE TABLE post TYPE NORMAL;
```

### 2. Selalu Gunakan IF NOT EXISTS

```sql
DEFINE TABLE IF NOT EXISTS user ...;
DEFINE FIELD IF NOT EXISTS username ON user ...;
DEFINE INDEX IF NOT EXISTS idx_username ON user ...;
```

### 3. Gunakan FLEXIBLE untuk Dynamic Data

```sql
-- ✅ Good: Flexible metadata
DEFINE FIELD metadata ON workflow FLEXIBLE TYPE object;

-- ❌ Bad: Rigid structure untuk data yang berubah
DEFINE FIELD metadata ON workflow TYPE object;
DEFINE FIELD metadata.key1 ON workflow TYPE string;
DEFINE FIELD metadata.key2 ON workflow TYPE string;
```

### 4. Gunakan Custom Functions untuk Business Logic

```sql
-- ✅ Good: Reusable function
DEFINE FUNCTION fn::can_approve($user_id, $task_id) { ... };
SELECT * FROM task_instance WHERE fn::can_approve($auth.id, id);

-- ❌ Bad: Duplicate logic
SELECT * FROM task_instance WHERE assigned_to = $auth.id AND status = "pending";
```

### 5. Gunakan Events untuk Audit Trail

```sql
-- ✅ Good: Automatic audit
DEFINE EVENT evt_audit ON table WHEN $event IN ["CREATE", "UPDATE", "DELETE"] THEN { ... };

-- ❌ Bad: Manual audit di aplikasi
-- (mudah lupa, tidak konsisten)
```

## Resources

- [SurrealDB v1.4 Documentation](https://surrealdb.com/docs)
- [SurrealQL Reference](https://surrealdb.com/docs/surrealql)
- [Graph Relations](https://surrealdb.com/docs/surrealql/statements/relate)
- [Full-Text Search](https://surrealdb.com/docs/surrealql/statements/define/indexes)

## Next Steps

1. ✅ Apply schema v1.4
2. ✅ Test graph traversal
3. ✅ Test custom functions
4. ✅ Test full-text search
5. ✅ Test permissions
6. ✅ Test audit trail
7. 🔄 Update Go backend untuk menggunakan fitur baru
8. 🔄 Update frontend untuk menampilkan audit trail
9. 🔄 Load testing dengan schema baru

---

**Last Updated**: 2026-04-29
**Version**: 1.4.0
**Status**: Production Ready
