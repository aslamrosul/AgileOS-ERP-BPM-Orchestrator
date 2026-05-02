-- AgileOS Database Schema - SurrealDB v1.4
-- Menggunakan fitur baru: TYPE RELATION, TYPE NORMAL, IF NOT EXISTS

-- ============================================
-- NAMESPACE & DATABASE
-- ============================================
DEFINE NAMESPACE IF NOT EXISTS agileos;
USE NS agileos;

DEFINE DATABASE IF NOT EXISTS main;
USE DB main;

-- ============================================
-- TABLE: USER (TYPE NORMAL)
-- ============================================
DEFINE TABLE IF NOT EXISTS user SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE id = $auth.id OR $auth.role IN ['admin', 'manager']
        FOR create WHERE $auth.role = 'admin'
        FOR update WHERE id = $auth.id OR $auth.role = 'admin'
        FOR delete WHERE $auth.role = 'admin';

DEFINE FIELD IF NOT EXISTS username ON user TYPE string
    ASSERT $value != NONE AND string::len($value) >= 3;
DEFINE FIELD IF NOT EXISTS password_hash ON user TYPE string
    PERMISSIONS FOR select WHERE NONE;
DEFINE FIELD IF NOT EXISTS email ON user TYPE string
    ASSERT $value != NONE AND string::len($value) > 0;
DEFINE FIELD IF NOT EXISTS role ON user TYPE string
    ASSERT $value IN ['admin', 'manager', 'employee', 'finance', 'procurement']
    DEFAULT 'employee';
DEFINE FIELD IF NOT EXISTS full_name ON user TYPE option<string>;
DEFINE FIELD IF NOT EXISTS department ON user TYPE option<string>;
DEFINE FIELD IF NOT EXISTS is_active ON user TYPE bool DEFAULT true;
DEFINE FIELD IF NOT EXISTS created_at ON user TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS updated_at ON user TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS last_login_at ON user TYPE option<datetime>;

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_user_username ON user FIELDS username UNIQUE;
DEFINE INDEX IF NOT EXISTS idx_user_email ON user FIELDS email UNIQUE;
DEFINE INDEX IF NOT EXISTS idx_user_role ON user FIELDS role;
DEFINE INDEX IF NOT EXISTS idx_user_active ON user FIELDS active;

-- ============================================
-- TABLE: WORKFLOW (TYPE NORMAL)
-- ============================================
DEFINE TABLE IF NOT EXISTS workflow SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select FULL
        FOR create, update, delete WHERE $auth.role IN ['admin', 'manager'];

DEFINE FIELD IF NOT EXISTS name ON workflow TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS description ON workflow TYPE option<string>;
DEFINE FIELD IF NOT EXISTS version ON workflow TYPE int DEFAULT 1;
DEFINE FIELD IF NOT EXISTS status ON workflow TYPE string
    ASSERT $value IN ['draft', 'active', 'archived']
    DEFAULT 'draft';
DEFINE FIELD IF NOT EXISTS created_by ON workflow TYPE record<user>;
DEFINE FIELD IF NOT EXISTS created_at ON workflow TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS updated_at ON workflow TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS metadata ON workflow TYPE object FLEXIBLE;

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_workflow_status ON workflow FIELDS status;
DEFINE INDEX IF NOT EXISTS idx_workflow_created_at ON workflow FIELDS created_at;
DEFINE INDEX IF NOT EXISTS idx_workflow_name ON workflow FIELDS name;

-- ============================================
-- TABLE: STEP (TYPE NORMAL)
-- ============================================
DEFINE TABLE IF NOT EXISTS step SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select FULL
        FOR create, update, delete WHERE $auth.role IN ['admin', 'manager'];

DEFINE FIELD IF NOT EXISTS workflow_id ON step TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS name ON step TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS type ON step TYPE string
    ASSERT $value IN ['start', 'action', 'approval', 'decision', 'notify', 'end']
    DEFAULT 'action';
DEFINE FIELD IF NOT EXISTS assigned_to ON step TYPE option<string>;
DEFINE FIELD IF NOT EXISTS sla ON step TYPE duration DEFAULT 24h;
DEFINE FIELD IF NOT EXISTS config ON step TYPE object FLEXIBLE;
DEFINE FIELD IF NOT EXISTS created_at ON step TYPE datetime DEFAULT time::now();

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_step_workflow_id ON step FIELDS workflow_id;
DEFINE INDEX IF NOT EXISTS idx_step_type ON step FIELDS type;
DEFINE INDEX IF NOT EXISTS idx_step_assigned_to ON step FIELDS assigned_to;

-- ============================================
-- TABLE: NEXT (TYPE RELATION) - Graph Edge
-- ============================================
DEFINE TABLE IF NOT EXISTS next SCHEMAFULL TYPE RELATION
    IN step
    OUT step
    PERMISSIONS
        FOR select FULL
        FOR create, update, delete WHERE $auth.role IN ['admin', 'manager'];

DEFINE FIELD IF NOT EXISTS in ON next TYPE record<step>;
DEFINE FIELD IF NOT EXISTS out ON next TYPE record<step>;
DEFINE FIELD IF NOT EXISTS condition ON next TYPE option<object> FLEXIBLE;
DEFINE FIELD IF NOT EXISTS priority ON next TYPE int DEFAULT 0;
DEFINE FIELD IF NOT EXISTS created_at ON next TYPE datetime DEFAULT time::now();

-- ============================================
-- TABLE: PROCESS_INSTANCE (TYPE NORMAL)
-- ============================================
DEFINE TABLE IF NOT EXISTS process_instance SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE initiated_by = $auth.id OR $auth.role IN ['admin', 'manager']
        FOR create WHERE $auth.id != NONE
        FOR update WHERE initiated_by = $auth.id OR $auth.role IN ['admin', 'manager']
        FOR delete WHERE $auth.role = 'admin';

DEFINE FIELD IF NOT EXISTS workflow_id ON process_instance TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS current_step_id ON process_instance TYPE option<string>;
DEFINE FIELD IF NOT EXISTS status ON process_instance TYPE string
    ASSERT $value IN ['running', 'completed', 'failed', 'cancelled']
    DEFAULT 'running';
DEFINE FIELD IF NOT EXISTS initiated_by ON process_instance TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS started_at ON process_instance TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS completed_at ON process_instance TYPE option<datetime>;
DEFINE FIELD IF NOT EXISTS data ON process_instance TYPE object FLEXIBLE;
DEFINE FIELD IF NOT EXISTS execution_history ON process_instance TYPE array DEFAULT [];

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_process_status ON process_instance FIELDS status;
DEFINE INDEX IF NOT EXISTS idx_process_workflow_id ON process_instance FIELDS workflow_id;
DEFINE INDEX IF NOT EXISTS idx_process_initiated_by ON process_instance FIELDS initiated_by;
DEFINE INDEX IF NOT EXISTS idx_process_started_at ON process_instance FIELDS started_at;

-- ============================================
-- TABLE: TASK_INSTANCE (TYPE NORMAL)
-- ============================================
DEFINE TABLE IF NOT EXISTS task_instance SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE assigned_to = $auth.id OR $auth.role IN ['admin', 'manager']
        FOR create WHERE $auth.role IN ['admin', 'manager']
        FOR update WHERE assigned_to = $auth.id OR $auth.role IN ['admin', 'manager']
        FOR delete WHERE $auth.role = 'admin';

DEFINE FIELD IF NOT EXISTS process_instance_id ON task_instance TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS step_id ON task_instance TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS step_name ON task_instance TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS status ON task_instance TYPE string
    ASSERT $value IN ['pending', 'in_progress', 'completed', 'rejected', 'cancelled']
    DEFAULT 'pending';
DEFINE FIELD IF NOT EXISTS assigned_to ON task_instance TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS created_at ON task_instance TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS due_at ON task_instance TYPE option<datetime>;
DEFINE FIELD IF NOT EXISTS completed_at ON task_instance TYPE option<datetime>;
DEFINE FIELD IF NOT EXISTS completed_by ON task_instance TYPE option<string>;
DEFINE FIELD IF NOT EXISTS result ON task_instance TYPE option<object> FLEXIBLE;
DEFINE FIELD IF NOT EXISTS signature ON task_instance TYPE option<string>;

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_task_status ON task_instance FIELDS status;
DEFINE INDEX IF NOT EXISTS idx_task_assigned_to ON task_instance FIELDS assigned_to;
DEFINE INDEX IF NOT EXISTS idx_task_process_id ON task_instance FIELDS process_instance_id;
DEFINE INDEX IF NOT EXISTS idx_task_created_at ON task_instance FIELDS created_at;
DEFINE INDEX IF NOT EXISTS idx_task_status_assigned ON task_instance FIELDS status, assigned_to;
DEFINE INDEX IF NOT EXISTS idx_task_due_at ON task_instance FIELDS due_at;

-- ============================================
-- TABLE: AUDIT_TRAILS (TYPE NORMAL)
-- ============================================
DEFINE TABLE IF NOT EXISTS audit_trails SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE $auth.role IN ['admin', 'manager']
        FOR create FULL
        FOR update, delete WHERE NONE;

DEFINE FIELD IF NOT EXISTS actor_id ON audit_trails TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS action ON audit_trails TYPE string
    ASSERT $value IN ['CREATE', 'UPDATE', 'DELETE', 'APPROVE', 'REJECT', 'COMPLETE']
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS resource_type ON audit_trails TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS resource_id ON audit_trails TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS old_value ON audit_trails TYPE option<object> FLEXIBLE;
DEFINE FIELD IF NOT EXISTS new_value ON audit_trails TYPE option<object> FLEXIBLE;
DEFINE FIELD IF NOT EXISTS timestamp ON audit_trails TYPE datetime DEFAULT time::now()
    READONLY;
DEFINE FIELD IF NOT EXISTS ip_address ON audit_trails TYPE option<string>;
DEFINE FIELD IF NOT EXISTS user_agent ON audit_trails TYPE option<string>;
DEFINE FIELD IF NOT EXISTS metadata ON audit_trails TYPE option<object> FLEXIBLE;

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_audit_actor ON audit_trails FIELDS actor_id;
DEFINE INDEX IF NOT EXISTS idx_audit_action ON audit_trails FIELDS action;
DEFINE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_trails FIELDS timestamp;
DEFINE INDEX IF NOT EXISTS idx_audit_resource ON audit_trails FIELDS resource_type, resource_id;
DEFINE INDEX IF NOT EXISTS idx_audit_actor_timestamp ON audit_trails FIELDS actor_id, timestamp;

-- ============================================
-- TABLE: WORKFLOW_VERSIONS (TYPE NORMAL)
-- ============================================
DEFINE TABLE IF NOT EXISTS workflow_versions SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE $auth.role IN ['admin', 'manager']
        FOR create WHERE $auth.role IN ['admin', 'manager']
        FOR update, delete WHERE NONE;

DEFINE FIELD IF NOT EXISTS workflow_id ON workflow_versions TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS version ON workflow_versions TYPE int
    ASSERT $value > 0;
DEFINE FIELD IF NOT EXISTS definition ON workflow_versions TYPE object FLEXIBLE
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS created_by ON workflow_versions TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS created_at ON workflow_versions TYPE datetime DEFAULT time::now()
    READONLY;
DEFINE FIELD IF NOT EXISTS change_notes ON workflow_versions TYPE option<string>;

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_version_workflow_id ON workflow_versions FIELDS workflow_id;
DEFINE INDEX IF NOT EXISTS idx_version_number ON workflow_versions FIELDS version;
DEFINE INDEX IF NOT EXISTS idx_version_created_at ON workflow_versions FIELDS created_at;

-- ============================================
-- TABLE: DOCUMENTS (TYPE NORMAL)
-- ============================================
DEFINE TABLE IF NOT EXISTS documents SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE owner_id = $auth.id OR $auth.role IN ['admin', 'manager']
        FOR create WHERE $auth.id != NONE
        FOR update WHERE owner_id = $auth.id OR $auth.role IN ['admin', 'manager']
        FOR delete WHERE owner_id = $auth.id OR $auth.role = 'admin';

DEFINE FIELD IF NOT EXISTS workflow_id ON documents TYPE option<string>;
DEFINE FIELD IF NOT EXISTS process_instance_id ON documents TYPE option<string>;
DEFINE FIELD IF NOT EXISTS task_instance_id ON documents TYPE option<string>;
DEFINE FIELD IF NOT EXISTS owner_id ON documents TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS filename ON documents TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS file_path ON documents TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS file_size ON documents TYPE int
    ASSERT $value > 0;
DEFINE FIELD IF NOT EXISTS mime_type ON documents TYPE string;
DEFINE FIELD IF NOT EXISTS signature_hash ON documents TYPE option<string>;
DEFINE FIELD IF NOT EXISTS signature_history ON documents TYPE array DEFAULT [];
DEFINE FIELD IF NOT EXISTS created_at ON documents TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS metadata ON documents TYPE option<object> FLEXIBLE;

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_document_workflow_id ON documents FIELDS workflow_id;
DEFINE INDEX IF NOT EXISTS idx_document_process_id ON documents FIELDS process_instance_id;
DEFINE INDEX IF NOT EXISTS idx_document_owner_id ON documents FIELDS owner_id;
DEFINE INDEX IF NOT EXISTS idx_document_created_at ON documents FIELDS created_at;

-- ============================================
-- EVENTS - Audit Trail Automation
-- ============================================

-- Event: Log workflow changes
DEFINE EVENT IF NOT EXISTS evt_workflow_audit ON workflow
    WHEN $event IN ["CREATE", "UPDATE", "DELETE"]
    THEN {
        CREATE audit_trails SET
            actor_id = $auth.id ?? "system",
            action = $event,
            resource_type = "workflow",
            resource_id = $value.id ?? "unknown",
            old_value = $before,
            new_value = $after,
            timestamp = time::now();
    };

-- Event: Log task completion
DEFINE EVENT IF NOT EXISTS evt_task_audit ON task_instance
    WHEN $event IN ["CREATE", "UPDATE", "DELETE"]
    THEN {
        CREATE audit_trails SET
            actor_id = $auth.id ?? "system",
            action = $event,
            resource_type = "task_instance",
            resource_id = $value.id ?? "unknown",
            old_value = $before,
            new_value = $after,
            timestamp = time::now();
    };

-- Event: Log process instance changes
DEFINE EVENT IF NOT EXISTS evt_process_audit ON process_instance
    WHEN $event IN ["CREATE", "UPDATE", "DELETE"]
    THEN {
        CREATE audit_trails SET
            actor_id = $auth.id ?? "system",
            action = $event,
            resource_type = "process_instance",
            resource_id = $value.id ?? "unknown",
            old_value = $before,
            new_value = $after,
            timestamp = time::now();
    };

-- ============================================
-- FUNCTIONS - Custom Business Logic
-- ============================================

-- Function: Check if user is admin
DEFINE FUNCTION IF NOT EXISTS fn::is_admin($user_id: string) {
    LET $user = (SELECT VALUE role FROM user WHERE id = $user_id LIMIT 1);
    RETURN array::len($user) > 0 AND $user[0] = "admin";
};

-- Function: Check if user can approve
DEFINE FUNCTION IF NOT EXISTS fn::can_approve($user_id: string, $task_id: string) {
    LET $task = (SELECT assigned_to, status FROM task_instance WHERE id = $task_id LIMIT 1);
    RETURN array::len($task) > 0 AND $task[0].assigned_to = $user_id AND $task[0].status = "pending";
};

-- Function: Get pending tasks count
DEFINE FUNCTION IF NOT EXISTS fn::pending_tasks_count($user_id: string) {
    LET $tasks = (SELECT VALUE count() FROM task_instance WHERE assigned_to = $user_id AND status = "pending" GROUP ALL);
    RETURN IF array::len($tasks) > 0 THEN $tasks[0] ELSE 0 END;
};

-- Function: Calculate task duration
DEFINE FUNCTION IF NOT EXISTS fn::task_duration($task_id: string) {
    LET $task = (SELECT created_at, completed_at FROM task_instance WHERE id = $task_id LIMIT 1);
    IF array::len($task) > 0 {
        IF $task[0].completed_at != NONE {
            RETURN $task[0].completed_at - $task[0].created_at;
        } ELSE {
            RETURN time::now() - $task[0].created_at;
        };
    } ELSE {
        RETURN 0s;
    };
};

-- ============================================
-- ANALYZERS - Full-Text Search
-- ============================================

-- Simple analyzer for basic text search
DEFINE ANALYZER IF NOT EXISTS simple_analyzer
    TOKENIZERS blank
    FILTERS lowercase;

-- English analyzer with stemming
DEFINE ANALYZER IF NOT EXISTS english_analyzer
    TOKENIZERS blank
    FILTERS lowercase, snowball(English);

-- ============================================
-- FULL-TEXT SEARCH INDEXES
-- ============================================

-- Full-text search on workflow name and description
DEFINE INDEX IF NOT EXISTS idx_workflow_fts ON workflow FIELDS name, description
    SEARCH ANALYZER simple_analyzer BM25;

-- Full-text search on task step name
DEFINE INDEX IF NOT EXISTS idx_task_fts ON task_instance FIELDS step_name
    SEARCH ANALYZER simple_analyzer BM25;

-- ============================================
-- VERIFICATION QUERIES
-- ============================================

-- Verify tables
INFO FOR TABLE user;
INFO FOR TABLE workflow;
INFO FOR TABLE step;
INFO FOR TABLE next;
INFO FOR TABLE process_instance;
INFO FOR TABLE task_instance;
INFO FOR TABLE audit_trails;
INFO FOR TABLE workflow_versions;
INFO FOR TABLE documents;
