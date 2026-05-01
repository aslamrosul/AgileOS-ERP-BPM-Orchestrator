# SurrealDB v1.4 Schema Verification

## ✅ Fitur v1.4 yang Sudah Diimplementasi

### 1. **IF NOT EXISTS** ✓
Semua definisi menggunakan `IF NOT EXISTS` untuk idempotency:
```surql
DEFINE NAMESPACE IF NOT EXISTS agileos;
DEFINE DATABASE IF NOT EXISTS main;
DEFINE TABLE IF NOT EXISTS user ...
DEFINE FIELD IF NOT EXISTS username ...
DEFINE INDEX IF NOT EXISTS idx_user_username ...
DEFINE EVENT IF NOT EXISTS evt_workflow_audit ...
DEFINE FUNCTION IF NOT EXISTS fn::is_admin ...
DEFINE ANALYZER IF NOT EXISTS simple_analyzer ...
```

### 2. **TYPE NORMAL** ✓
Tabel record biasa menggunakan `TYPE NORMAL`:
- `user` - TYPE NORMAL
- `workflow` - TYPE NORMAL
- `step` - TYPE NORMAL
- `process_instance` - TYPE NORMAL
- `task_instance` - TYPE NORMAL
- `audit_trails` - TYPE NORMAL
- `workflow_versions` - TYPE NORMAL
- `documents` - TYPE NORMAL

### 3. **TYPE RELATION** ✓
Tabel edge graph menggunakan `TYPE RELATION`:
```surql
DEFINE TABLE IF NOT EXISTS next SCHEMAFULL TYPE RELATION
    IN step
    OUT step
    PERMISSIONS ...;

-- Field in dan out harus didefinisikan eksplisit
DEFINE FIELD IF NOT EXISTS in ON next TYPE record<step>;
DEFINE FIELD IF NOT EXISTS out ON next TYPE record<step>;
```

### 4. **SCHEMAFULL** ✓
Semua tabel menggunakan `SCHEMAFULL` untuk validasi ketat:
- Semua field harus didefinisikan
- Type checking otomatis
- ASSERT untuk validasi data

### 5. **FLEXIBLE Object** ✓
Field object yang perlu nested bebas menggunakan `FLEXIBLE`:
```surql
DEFINE FIELD IF NOT EXISTS metadata ON workflow TYPE object FLEXIBLE;
DEFINE FIELD IF NOT EXISTS config ON step TYPE object FLEXIBLE;
DEFINE FIELD IF NOT EXISTS data ON process_instance TYPE object FLEXIBLE;
```

### 6. **PERMISSIONS** ✓
Semua tabel memiliki permissions berbasis role:
```surql
PERMISSIONS
    FOR select WHERE id = $auth.id OR $auth.role IN ['admin', 'manager']
    FOR create WHERE $auth.role = 'admin'
    FOR update WHERE id = $auth.id OR $auth.role = 'admin'
    FOR delete WHERE $auth.role = 'admin'
```

### 7. **EVENTS** ✓
Audit trail otomatis menggunakan events:
- `evt_workflow_audit` - Log perubahan workflow
- `evt_task_audit` - Log perubahan task
- `evt_process_audit` - Log perubahan process

### 8. **CUSTOM FUNCTIONS** ✓
Business logic dalam fungsi:
- `fn::is_admin($user_id)` - Cek role admin
- `fn::can_approve($user_id, $task_id)` - Cek hak approve
- `fn::pending_tasks_count($user_id)` - Hitung pending tasks
- `fn::task_duration($task_id)` - Hitung durasi task

### 9. **FULL-TEXT SEARCH** ✓
BM25 search dengan analyzer:
```surql
DEFINE ANALYZER IF NOT EXISTS simple_analyzer
    TOKENIZERS blank
    FILTERS lowercase;

DEFINE INDEX IF NOT EXISTS idx_workflow_fts ON workflow FIELDS name, description
    SEARCH ANALYZER simple_analyzer BM25;
```

### 10. **INDEXES** ✓
Indexes untuk performa query:
- UNIQUE indexes: username, email
- Compound indexes: status + assigned_to
- Timestamp indexes untuk sorting
- Foreign key indexes

## 🔧 Perbaikan yang Dilakukan

### 1. Field `in` dan `out` pada TYPE RELATION
**Sebelum:**
```surql
DEFINE TABLE next TYPE RELATION IN step OUT step;
-- Field in/out tidak didefinisikan eksplisit
```

**Sesudah:**
```surql
DEFINE TABLE next TYPE RELATION IN step OUT step;
DEFINE FIELD IF NOT EXISTS in ON next TYPE record<step>;
DEFINE FIELD IF NOT EXISTS out ON next TYPE record<step>;
```

### 2. Function dengan Array Handling
**Sebelum:**
```surql
DEFINE FUNCTION fn::is_admin($user_id: string) {
    LET $user = (SELECT role FROM user WHERE id = $user_id);
    RETURN $user[0].role = "admin";  -- Error jika array kosong
};
```

**Sesudah:**
```surql
DEFINE FUNCTION fn::is_admin($user_id: string) {
    LET $user = (SELECT VALUE role FROM user WHERE id = $user_id LIMIT 1);
    RETURN array::len($user) > 0 AND $user[0] = "admin";
};
```

### 3. Event dengan Default Values
**Sebelum:**
```surql
CREATE audit_trails SET
    actor_id = $auth.id,  -- Error jika $auth.id = NONE
```

**Sesudah:**
```surql
CREATE audit_trails SET
    actor_id = string::default($auth.id, "system"),
    resource_id = string::default($value.id, "unknown"),
```

## 📊 Struktur Database

### Graph Relationship
```
workflow (1) ──┐
               ├──> step (N) ──> next (RELATION) ──> step (N)
               │
               └──> process_instance (N) ──> task_instance (N)
```

### Audit Trail
```
user ──> workflow ──> EVENT ──> audit_trails
user ──> task_instance ──> EVENT ──> audit_trails
user ──> process_instance ──> EVENT ──> audit_trails
```

## 🧪 Testing Schema

### 1. Apply Schema
```bash
# PowerShell
.\backend-go\scripts\apply-schema-v1.4.ps1

# Atau manual
surreal sql --endpoint ws://localhost:8000 \
    --username root --password root \
    --namespace agileos --database main \
    --file backend-go/database/schema-v1.4.surql
```

### 2. Verify Tables
```surql
-- List all tables
INFO FOR DB;

-- Check specific table
INFO FOR TABLE user;
INFO FOR TABLE next;
INFO FOR TABLE task_instance;

-- Verify TYPE RELATION
SELECT * FROM next LIMIT 1;
```

### 3. Test Graph Traversal
```surql
-- Create workflow with steps
CREATE workflow:test SET name = "Test Workflow";
CREATE step:start SET workflow_id = "workflow:test", name = "Start", type = "start";
CREATE step:approve SET workflow_id = "workflow:test", name = "Approve", type = "approval";

-- Create edge
RELATE step:start->next->step:approve SET priority = 1;

-- Traverse graph
SELECT ->next->step.* FROM step:start;
SELECT <-next<-step.* FROM step:approve;
```

### 4. Test Permissions
```surql
-- As admin
DEFINE USER admin ON DATABASE PASSWORD 'admin123' ROLES OWNER;

-- As regular user (should fail for some operations)
DEFINE USER employee ON DATABASE PASSWORD 'emp123' ROLES VIEWER;
```

### 5. Test Events
```surql
-- Create workflow (should trigger audit event)
CREATE workflow SET name = "Test", created_by = user:admin;

-- Check audit trail
SELECT * FROM audit_trails WHERE resource_type = "workflow" ORDER BY timestamp DESC LIMIT 5;
```

### 6. Test Functions
```surql
-- Test is_admin
RETURN fn::is_admin("user:admin");

-- Test pending tasks count
RETURN fn::pending_tasks_count("user:employee");

-- Test task duration
RETURN fn::task_duration("task_instance:abc123");
```

### 7. Test Full-Text Search
```surql
-- Create test data
CREATE workflow SET name = "Purchase Order Approval", description = "Workflow for approving purchase orders";
CREATE workflow SET name = "Leave Request", description = "Employee leave request workflow";

-- Search
SELECT *, search::score(1) AS score 
FROM workflow 
WHERE name @1@ "approval" OR description @1@ "approval"
ORDER BY score DESC;
```

## ✅ Checklist Kompatibilitas v1.4

- [x] IF NOT EXISTS pada semua DEFINE statements
- [x] TYPE NORMAL untuk tabel record
- [x] TYPE RELATION untuk edge graph dengan IN/OUT
- [x] Field `in` dan `out` eksplisit pada TYPE RELATION
- [x] SCHEMAFULL untuk validasi ketat
- [x] FLEXIBLE untuk nested objects
- [x] PERMISSIONS berbasis role
- [x] EVENTS untuk audit automation
- [x] CUSTOM FUNCTIONS dengan error handling
- [x] FULL-TEXT SEARCH dengan BM25
- [x] INDEXES untuk performa
- [x] READONLY fields untuk immutable data
- [x] DEFAULT values untuk semua field
- [x] ASSERT untuk validasi data
- [x] option<T> untuk nullable fields
- [x] array untuk collections
- [x] record<table> untuk foreign keys

## 🚀 Migration dari v0.x ke v1.4

Jika Anda memiliki data lama:

```surql
-- 1. Export data lama
surreal export --conn http://localhost:8000 \
    --user root --pass root \
    --ns agileos --db main \
    old_data.surql

-- 2. Backup database
cp -r data/ data_backup/

-- 3. Apply new schema
surreal sql --file backend-go/database/schema-v1.4.surql

-- 4. Import data (akan otomatis validasi dengan schema baru)
surreal import --conn http://localhost:8000 \
    --user root --pass root \
    --ns agileos --db main \
    old_data.surql
```

## 📚 Referensi

- [SurrealDB v1.4 Release Notes](https://surrealdb.com/releases/1.4.0)
- [TYPE RELATION Documentation](https://surrealdb.com/docs/surrealql/statements/define/table#type-relation)
- [IF NOT EXISTS Documentation](https://surrealdb.com/docs/surrealql/statements/define#if-not-exists)
- [Permissions Documentation](https://surrealdb.com/docs/surrealql/statements/define/table#permissions)
- [Events Documentation](https://surrealdb.com/docs/surrealql/statements/define/event)
- [Functions Documentation](https://surrealdb.com/docs/surrealql/statements/define/function)

## 🎯 Kesimpulan

Schema database AgileOS **sudah sepenuhnya kompatibel dengan SurrealDB v1.4** dengan implementasi:

1. ✅ Semua fitur baru v1.4 (TYPE RELATION, IF NOT EXISTS)
2. ✅ Best practices (SCHEMAFULL, PERMISSIONS, EVENTS)
3. ✅ Graph database untuk workflow orchestration
4. ✅ Audit trail otomatis
5. ✅ Full-text search
6. ✅ Custom business logic functions
7. ✅ Proper error handling dan validasi

Schema siap digunakan untuk production! 🚀
