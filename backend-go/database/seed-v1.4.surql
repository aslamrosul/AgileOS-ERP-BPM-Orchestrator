-- AgileOS Seed Data - SurrealDB v1.4
-- Sample data untuk testing dan development

USE NS agileos DB main;

-- ============================================
-- SEED: USERS
-- ============================================

-- Admin user
CREATE user:admin SET
    username = "admin",
    password_hash = crypto::argon2::generate("password123"),
    email = "admin@agileos.com",
    role = "admin",
    department = "IT",
    is_active = true,
    created_at = time::now();

-- Manager user
CREATE user:manager SET
    username = "manager",
    password_hash = crypto::argon2::generate("password123"),
    email = "manager@agileos.com",
    role = "manager",
    department = "Operations",
    is_active = true,
    created_at = time::now();

-- Employee user
CREATE user:employee SET
    username = "employee",
    password_hash = crypto::argon2::generate("password123"),
    email = "employee@agileos.com",
    role = "employee",
    department = "Sales",
    is_active = true,
    created_at = time::now();

-- Finance user
CREATE user:finance SET
    username = "finance",
    password_hash = crypto::argon2::generate("password123"),
    email = "finance@agileos.com",
    role = "finance",
    department = "Finance",
    is_active = true,
    created_at = time::now();

-- Procurement user
CREATE user:procurement SET
    username = "procurement",
    password_hash = crypto::argon2::generate("password123"),
    email = "procurement@agileos.com",
    role = "procurement",
    department = "Procurement",
    is_active = true,
    created_at = time::now();

-- ============================================
-- SEED: WORKFLOW - Purchase Approval
-- ============================================

CREATE workflow:purchase_approval SET
    name = "Purchase Approval Workflow",
    description = "Standard workflow for purchase requests requiring multi-level approval",
    version = 1,
    status = "active",
    created_by = user:admin,
    created_at = time::now(),
    metadata = {
        category: "procurement",
        max_amount: 100000,
        requires_finance: true
    };

-- ============================================
-- SEED: STEPS for Purchase Approval
-- ============================================

-- Step 1: Start
CREATE step:purchase_start SET
    workflow_id = "purchase_approval",
    name = "Purchase Request Initiated",
    type = "start",
    assigned_to = NONE,
    sla = 0s,
    config = {
        auto_proceed: true
    },
    created_at = time::now();

-- Step 2: Manager Approval
CREATE step:manager_approval SET
    workflow_id = "purchase_approval",
    name = "Manager Approval",
    type = "approval",
    assigned_to = "manager",
    sla = 24h,
    config = {
        approval_required: true,
        can_reject: true,
        notification_enabled: true
    },
    created_at = time::now();

-- Step 3: Finance Review
CREATE step:finance_review SET
    workflow_id = "purchase_approval",
    name = "Finance Review",
    type = "approval",
    assigned_to = "finance",
    sla = 48h,
    config = {
        approval_required: true,
        can_reject: true,
        budget_check: true
    },
    created_at = time::now();

-- Step 4: Procurement Processing
CREATE step:procurement_process SET
    workflow_id = "purchase_approval",
    name = "Procurement Processing",
    type = "action",
    assigned_to = "procurement",
    sla = 72h,
    config = {
        create_po: true,
        vendor_selection: true
    },
    created_at = time::now();

-- Step 5: Completion Notification
CREATE step:purchase_complete SET
    workflow_id = "purchase_approval",
    name = "Purchase Completed",
    type = "notify",
    assigned_to = NONE,
    sla = 1h,
    config = {
        notify_requester: true,
        notify_finance: true,
        notify_manager: true
    },
    created_at = time::now();

-- Step 6: End
CREATE step:purchase_end SET
    workflow_id = "purchase_approval",
    name = "Process Ended",
    type = "end",
    assigned_to = NONE,
    sla = 0s,
    config = {
        archive: true
    },
    created_at = time::now();

-- ============================================
-- SEED: GRAPH RELATIONS (next edges)
-- ============================================

-- Start -> Manager Approval
RELATE step:purchase_start->next->step:manager_approval SET
    condition = NONE,
    priority = 1,
    created_at = time::now();

-- Manager Approval -> Finance Review (if approved)
RELATE step:manager_approval->next->step:finance_review SET
    condition = { status: "approved" },
    priority = 1,
    created_at = time::now();

-- Finance Review -> Procurement Processing (if approved)
RELATE step:finance_review->next->step:procurement_process SET
    condition = { status: "approved" },
    priority = 1,
    created_at = time::now();

-- Procurement Processing -> Completion Notification
RELATE step:procurement_process->next->step:purchase_complete SET
    condition = NONE,
    priority = 1,
    created_at = time::now();

-- Completion Notification -> End
RELATE step:purchase_complete->next->step:purchase_end SET
    condition = NONE,
    priority = 1,
    created_at = time::now();

-- ============================================
-- SEED: SAMPLE PROCESS INSTANCE
-- ============================================

CREATE process_instance:sample_001 SET
    workflow_id = "purchase_approval",
    current_step_id = "manager_approval",
    status = "running",
    initiated_by = "employee",
    started_at = time::now(),
    data = {
        item: "Dell Laptop XPS 15",
        quantity: 5,
        unit_price: 25000000,
        total_amount: 125000000,
        currency: "IDR",
        justification: "Replacement for old laptops in development team",
        department: "IT",
        urgency: "medium"
    },
    execution_history = [
        {
            step_id: "purchase_start",
            executed_at: time::now() - 1h,
            executed_by: "employee",
            action: "initiated",
            result: { status: "success" },
            duration: 5s
        }
    ];

-- ============================================
-- SEED: SAMPLE TASK INSTANCE
-- ============================================

CREATE task_instance:task_001 SET
    process_instance_id = "sample_001",
    step_id = "manager_approval",
    step_name = "Manager Approval",
    status = "pending",
    assigned_to = "manager",
    created_at = time::now(),
    due_at = time::now() + 24h,
    result = NONE,
    signature = NONE;

-- ============================================
-- SEED: WORKFLOW VERSION
-- ============================================

CREATE workflow_versions:purchase_v1 SET
    workflow_id = "purchase_approval",
    version = 1,
    definition = {
        name: "Purchase Approval Workflow",
        steps: [
            { id: "purchase_start", name: "Purchase Request Initiated", type: "start" },
            { id: "manager_approval", name: "Manager Approval", type: "approval" },
            { id: "finance_review", name: "Finance Review", type: "approval" },
            { id: "procurement_process", name: "Procurement Processing", type: "action" },
            { id: "purchase_complete", name: "Purchase Completed", type: "notify" },
            { id: "purchase_end", name: "Process Ended", type: "end" }
        ],
        edges: [
            { from: "purchase_start", to: "manager_approval" },
            { from: "manager_approval", to: "finance_review", condition: "approved" },
            { from: "finance_review", to: "procurement_process", condition: "approved" },
            { from: "procurement_process", to: "purchase_complete" },
            { from: "purchase_complete", to: "purchase_end" }
        ]
    },
    created_by = "admin",
    created_at = time::now(),
    change_notes = "Initial version of purchase approval workflow";

-- ============================================
-- VERIFICATION QUERIES
-- ============================================

-- Count users
SELECT count() FROM user;

-- Count workflows
SELECT count() FROM workflow;

-- Count steps
SELECT count() FROM step;

-- Count graph edges
SELECT count() FROM next;

-- Show workflow with steps (graph traversal)
SELECT 
    *,
    (SELECT * FROM step WHERE workflow_id = $parent.id) AS steps
FROM workflow:purchase_approval;

-- Show next steps from manager_approval
SELECT ->next->step.* AS next_steps FROM step:manager_approval;

-- Show pending tasks
SELECT * FROM task_instance WHERE status = "pending";

-- Show process instances
SELECT * FROM process_instance;
