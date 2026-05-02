-- AgileOS SurrealDB Performance Indexes
-- Run this script to optimize query performance
-- Usage: surreal import --conn http://localhost:8002 --user root --pass root --ns agileos --db main create-indexes.surql

-- ============================================
-- WORKFLOW INDEXES
-- ============================================

-- Index on workflow status for filtering active workflows
DEFINE INDEX idx_workflow_status ON TABLE workflow COLUMNS status;

-- Index on workflow created_at for sorting and time-based queries
DEFINE INDEX idx_workflow_created_at ON TABLE workflow COLUMNS created_at;

-- ============================================
-- STEP INDEXES
-- ============================================

-- Index on step workflow_id for fetching all steps of a workflow
DEFINE INDEX idx_step_workflow_id ON TABLE step COLUMNS workflow_id;

-- Index on step type for filtering by step type
DEFINE INDEX idx_step_type ON TABLE step COLUMNS type;

-- ============================================
-- TASK INSTANCE INDEXES
-- ============================================

-- Index on task status for filtering pending/completed tasks
DEFINE INDEX idx_task_status ON TABLE task_instance COLUMNS status;

-- Index on task assigned_to for user-specific task queries
DEFINE INDEX idx_task_assigned_to ON TABLE task_instance COLUMNS assigned_to;

-- Index on task process_instance_id for fetching all tasks of a process
DEFINE INDEX idx_task_process_id ON TABLE task_instance COLUMNS process_instance_id;

-- Index on task created_at for sorting and time-based queries
DEFINE INDEX idx_task_created_at ON TABLE task_instance COLUMNS created_at;

-- Composite index for common query: pending tasks for a user
DEFINE INDEX idx_task_status_assigned ON TABLE task_instance COLUMNS status, assigned_to;

-- ============================================
-- PROCESS INSTANCE INDEXES
-- ============================================

-- Index on process status for filtering active processes
DEFINE INDEX idx_process_status ON TABLE process_instance COLUMNS status;

-- Index on process workflow_id for analytics
DEFINE INDEX idx_process_workflow_id ON TABLE process_instance COLUMNS workflow_id;

-- Index on process started_at for time-based analytics
DEFINE INDEX idx_process_started_at ON TABLE process_instance COLUMNS started_at;

-- Index on process initiated_by for user activity tracking
DEFINE INDEX idx_process_initiated_by ON TABLE process_instance COLUMNS initiated_by;

-- ============================================
-- AUDIT TRAIL INDEXES
-- ============================================

-- Index on audit actor_id for user activity tracking
DEFINE INDEX idx_audit_actor ON TABLE audit_trails COLUMNS actor_id;

-- Index on audit action for filtering by action type
DEFINE INDEX idx_audit_action ON TABLE audit_trails COLUMNS action;

-- Index on audit timestamp for time-based queries
DEFINE INDEX idx_audit_timestamp ON TABLE audit_trails COLUMNS timestamp;

-- Composite index for common query: user actions in time range
DEFINE INDEX idx_audit_actor_timestamp ON TABLE audit_trails COLUMNS actor_id, timestamp;

-- ============================================
-- USER INDEXES
-- ============================================

-- Index on user username for login queries (unique)
DEFINE INDEX idx_user_username ON TABLE user COLUMNS username UNIQUE;

-- Index on user role for role-based queries
DEFINE INDEX idx_user_role ON TABLE user COLUMNS role;

-- ============================================
-- WORKFLOW VERSION INDEXES
-- ============================================

-- Index on workflow_versions workflow_id for version history
DEFINE INDEX idx_version_workflow_id ON TABLE workflow_versions COLUMNS workflow_id;

-- Index on workflow_versions version for sorting
DEFINE INDEX idx_version_number ON TABLE workflow_versions COLUMNS version;

-- Index on workflow_versions created_at for time-based queries
DEFINE INDEX idx_version_created_at ON TABLE workflow_versions COLUMNS created_at;

-- ============================================
-- DOCUMENT INDEXES
-- ============================================

-- Index on document workflow_id for document queries
DEFINE INDEX idx_document_workflow_id ON TABLE documents COLUMNS workflow_id;

-- Index on document created_at for sorting
DEFINE INDEX idx_document_created_at ON TABLE documents COLUMNS created_at;

-- ============================================
-- VERIFICATION QUERIES
-- ============================================

-- Run these queries to verify indexes are created:
-- INFO FOR TABLE workflow;
-- INFO FOR TABLE step;
-- INFO FOR TABLE task_instance;
-- INFO FOR TABLE process_instance;
-- INFO FOR TABLE audit_trails;
-- INFO FOR TABLE user;

-- ============================================
-- PERFORMANCE TESTING
-- ============================================

-- Test query performance with EXPLAIN:
-- SELECT * FROM task_instance WHERE status = 'pending' AND assigned_to = 'manager' EXPLAIN;
-- SELECT * FROM audit_trails WHERE actor_id = 'admin' AND timestamp > time::now() - 7d EXPLAIN;
