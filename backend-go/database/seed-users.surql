-- AgileOS BPM - User Seed Script
-- Run this in SurrealDB Dashboard or via CLI

USE NS agileos DB main;

-- Create sample users with hashed passwords
-- Password for all users: "password123"
-- Hash generated with bcrypt cost 12

-- Admin User
CREATE user:admin SET
    username = "admin",
    email = "admin@agileos.com",
    password_hash = "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIeWU7u3oi",
    role = "admin",
    full_name = "System Administrator",
    department = "IT",
    is_active = true,
    created_at = time::now(),
    updated_at = time::now();

-- Manager User
CREATE user:manager1 SET
    username = "manager",
    email = "manager@agileos.com",
    password_hash = "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIeWU7u3oi",
    role = "manager",
    full_name = "John Manager",
    department = "Operations",
    is_active = true,
    created_at = time::now(),
    updated_at = time::now();

-- Employee User
CREATE user:employee1 SET
    username = "employee",
    email = "employee@agileos.com",
    password_hash = "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIeWU7u3oi",
    role = "employee",
    full_name = "Jane Employee",
    department = "Sales",
    is_active = true,
    created_at = time::now(),
    updated_at = time::now();

-- Finance User
CREATE user:finance1 SET
    username = "finance",
    email = "finance@agileos.com",
    password_hash = "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIeWU7u3oi",
    role = "finance",
    full_name = "Bob Finance",
    department = "Finance",
    is_active = true,
    created_at = time::now(),
    updated_at = time::now();

-- Procurement User
CREATE user:procurement1 SET
    username = "procurement",
    email = "procurement@agileos.com",
    password_hash = "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIeWU7u3oi",
    role = "procurement",
    full_name = "Alice Procurement",
    department = "Procurement",
    is_active = true,
    created_at = time::now(),
    updated_at = time::now();

-- Verify users created
SELECT * FROM user;

-- Show user count by role
SELECT role, count() AS user_count FROM user GROUP BY role;
