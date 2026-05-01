-- AgileOS Audit Trail & Governance Schema
-- This script creates immutable audit trail tables and workflow versioning

-- ============================================
-- AUDIT TRAILS TABLE (Immutable)
-- ============================================
DEFINE TABLE audit_trails SCHEMAFULL;

-- Define fields for audit_trails
DEFINE FIELD timestamp ON audit_trails TYPE datetime;
DEFINE FIELD actor_id ON audit_trails TYPE string;
DEFINE FIELD actor_username ON audit_trails TYPE string;
DEFINE FIELD actor_role ON audit_trails TYPE string;
DEFINE FIELD action ON audit_trails TYPE string;
DEFINE FIELD resource_type ON audit_trails TYPE string;
DEFINE FIELD resource_id ON audit_trails TYPE string;
DEFINE FIELD old_value ON audit_trails TYPE object;
DEFINE FIELD new_value ON audit_trails TYPE object;
DEFINE FIELD ip_address ON audit_trails TYPE string;
DEFINE FIELD user_agent ON audit_trails TYPE string;
DEFINE FIELD compliance_status ON audit_trails TYPE string;
DEFINE FIELD compliance_notes ON audit_trails TYPE string;
DEFINE FIELD metadata ON audit_trails TYPE object;

-- Create indexes for efficient querying
DEFINE INDEX idx_audit_timestamp ON audit_trails FIELDS timestamp;
DEFINE INDEX idx_audit_actor ON audit_trails FIELDS actor_id;
DEFINE INDEX idx_audit_action ON audit_trails FIELDS action;
DEFINE INDEX idx_audit_resource ON audit_trails FIELDS resource_type, resource_id;
DEFINE INDEX idx_audit_compliance ON audit_trails FIELDS compliance_status;

-- ============================================
-- WORKFLOW VERSIONS TABLE (Versioning)
-- ============================================
DEFINE TABLE workflow_versions SCHEMAFULL;

-- Define fields for workflow_versions
DEFINE FIELD workflow_id ON workflow_versions TYPE string;
DEFINE FIELD version ON workflow_versions TYPE string;
DEFINE FIELD version_number ON workflow_versions TYPE number;
DEFINE FIELD name ON workflow_versions TYPE string;
DEFINE FIELD description ON workflow_versions TYPE string;
DEFINE FIELD definition ON workflow_versions TYPE object;
DEFINE FIELD created_by ON workflow_versions TYPE string;
DEFINE FIELD created_at ON workflow_versions TYPE datetime;
DEFINE FIELD change_reason ON workflow_versions TYPE string;
DEFINE FIELD is_active ON workflow_versions TYPE bool;
DEFINE FIELD approved_by ON workflow_versions TYPE string;
DEFINE FIELD approved_at ON workflow_versions TYPE datetime;

-- Create indexes for workflow versions
DEFINE INDEX idx_workflow_version ON workflow_versions FIELDS workflow_id, version_number;
DEFINE INDEX idx_workflow_active ON workflow_versions FIELDS workflow_id, is_active;

-- ============================================
-- SAMPLE AUDIT TRAIL DATA
-- ============================================

-- System initialization audit
CREATE audit_trails CONTENT {
    timestamp: time::now(),
    actor_id: "system",
    actor_username: "system",
    actor_role: "system",
    action: "SYSTEM_INIT",
    resource_type: "system",
    resource_id: "audit_system",
    new_value: {
        message: "Audit trail system initialized",
        version: "1.0.0"
    },
    ip_address: "127.0.0.1",
    compliance_status: "PASS",
    compliance_notes: "System initialization - automated",
    metadata: {
        initialization_time: time::now(),
        schema_version: "1.0"
    }
};

-- ============================================
-- DATA RETENTION POLICY
-- ============================================

-- Note: In production, implement automated archival process
-- Audit trails should be retained for at least 1 year for compliance
-- After retention period, move to cold storage or archive database

-- ============================================
-- GOVERNANCE RULES
-- ============================================

-- IMPORTANT: Audit trails are IMMUTABLE
-- Only INSERT operations are allowed
-- UPDATE and DELETE operations should be blocked at application level
-- This ensures data integrity and compliance with audit requirements

-- ============================================
-- COMPLIANCE CHECKS
-- ============================================

-- The following compliance checks are automated in the application:
-- 1. Unauthorized actions detection
-- 2. Self-approval prevention (conflict of interest)
-- 3. Role-based authorization validation
-- 4. Workflow change documentation requirements
-- 5. Business hours monitoring
-- 6. Rapid action detection (potential automation/bot)

-- ============================================
-- SAMPLE WORKFLOW VERSION
-- ============================================

-- Create initial version for purchase_approval workflow
CREATE workflow_versions CONTENT {
    workflow_id: "purchase_approval",
    version: "v1.0",
    version_number: 1,
    name: "Purchase Approval Workflow",
    description: "Standard purchase approval process for procurement",
    definition: {
        nodes: [
            {
                id: "start",
                type: "start",
                data: { label: "Start Purchase Request" }
            },
            {
                id: "manager_review",
                type: "approval",
                data: { 
                    label: "Manager Review",
                    assignee_role: "manager"
                }
            },
            {
                id: "finance_approval",
                type: "approval",
                data: { 
                    label: "Finance Approval",
                    assignee_role: "finance"
                }
            },
            {
                id: "end",
                type: "end",
                data: { label: "Purchase Approved" }
            }
        ],
        edges: [
            { source: "start", target: "manager_review" },
            { source: "manager_review", target: "finance_approval" },
            { source: "finance_approval", target: "end" }
        ]
    },
    created_by: "admin",
    created_at: time::now(),
    change_reason: "Initial workflow version",
    is_active: true
};

-- ============================================
-- AUDIT TRAIL QUERIES (Examples)
-- ============================================

-- Query all compliance violations
-- SELECT * FROM audit_trails WHERE compliance_status = 'FAIL' ORDER BY timestamp DESC;

-- Query audit trails by user
-- SELECT * FROM audit_trails WHERE actor_id = 'user123' ORDER BY timestamp DESC;

-- Query workflow changes
-- SELECT * FROM audit_trails WHERE action = 'WORKFLOW_CHANGE' ORDER BY timestamp DESC;

-- Query audit trails by date range
-- SELECT * FROM audit_trails WHERE timestamp >= '2024-01-01T00:00:00Z' AND timestamp <= '2024-12-31T23:59:59Z';

-- Count audit trails by action type
-- SELECT action, count() FROM audit_trails GROUP BY action;

-- ============================================
-- END OF AUDIT SCHEMA
-- ============================================