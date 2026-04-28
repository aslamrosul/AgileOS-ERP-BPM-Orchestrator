-- AgileOS BPM Database Seed Script
-- Run this in SurrealDB Dashboard or via CLI

-- Use namespace and database
USE NS agileos DB main;

-- Create a sample workflow: Purchase Request Approval
CREATE workflow:purchase_approval SET
    name = "Purchase Request Approval",
    version = "1.0.0",
    description = "Workflow for approving purchase requests",
    created_at = time::now(),
    updated_at = time::now(),
    is_active = true;

-- Create steps for the workflow
CREATE step:pr_submit SET
    workflow_id = "workflow:purchase_approval",
    name = "Submit Purchase Request",
    type = "action",
    assigned_to = "role:employee",
    sla = 1h,
    description = "Employee submits purchase request",
    created_at = time::now();

CREATE step:manager_review SET
    workflow_id = "workflow:purchase_approval",
    name = "Manager Review",
    type = "approval",
    assigned_to = "role:manager",
    sla = 24h,
    description = "Manager reviews and approves/rejects request",
    created_at = time::now();

CREATE step:finance_approval SET
    workflow_id = "workflow:purchase_approval",
    name = "Finance Approval",
    type = "approval",
    assigned_to = "role:finance",
    sla = 48h,
    description = "Finance team approves budget allocation",
    created_at = time::now();

CREATE step:procurement SET
    workflow_id = "workflow:purchase_approval",
    name = "Procurement Process",
    type = "action",
    assigned_to = "role:procurement",
    sla = 72h,
    description = "Procurement team processes the purchase",
    created_at = time::now();

CREATE step:completed SET
    workflow_id = "workflow:purchase_approval",
    name = "Completed",
    type = "notify",
    assigned_to = "role:employee",
    sla = 1h,
    description = "Notify requester of completion",
    created_at = time::now();

CREATE step:rejected SET
    workflow_id = "workflow:purchase_approval",
    name = "Rejected",
    type = "notify",
    assigned_to = "role:employee",
    sla = 1h,
    description = "Notify requester of rejection",
    created_at = time::now();

-- Create NEXT relationships (workflow flow)
RELATE step:pr_submit->next->step:manager_review;
RELATE step:manager_review->next->step:finance_approval SET condition = { decision: "approved" };
RELATE step:manager_review->next->step:rejected SET condition = { decision: "rejected" };
RELATE step:finance_approval->next->step:procurement SET condition = { decision: "approved" };
RELATE step:finance_approval->next->step:rejected SET condition = { decision: "rejected" };
RELATE step:procurement->next->step:completed;

-- Query to verify workflow structure
SELECT * FROM workflow:purchase_approval;

-- Query to verify all steps
SELECT * FROM step WHERE workflow_id = "workflow:purchase_approval";

-- Query to verify relationships (graph traversal)
SELECT id, name, ->next->step.name AS next_steps FROM step WHERE workflow_id = "workflow:purchase_approval";

-- Create a sample process instance
CREATE process_instance:test_001 SET
    workflow_id = "workflow:purchase_approval",
    current_step_id = "step:pr_submit",
    status = "running",
    started_at = time::now(),
    initiated_by = "user:john_doe",
    data = {
        amount: 5000,
        description: "New laptops for development team",
        quantity: 5
    };

-- Query to get next step from current step
SELECT ->next->step.* AS next_steps FROM step:pr_submit;
