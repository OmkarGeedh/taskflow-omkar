-- Seed data for TaskFlow
-- Creates test user, project, and tasks for immediate testing

-- Insert test user (password: password123)
-- Password hash generated with bcrypt cost 12
INSERT INTO users (id, name, email, password_hash, created_at)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Test User',
    'test@example.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIq8fPOQKm',
    NOW()
)
ON CONFLICT (email) DO NOTHING;

-- Insert test project
INSERT INTO projects (id, name, description, owner_id, created_at)
VALUES (
    'b1ffcd99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'Sample Project',
    'A test project with sample tasks',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    NOW()
)
ON CONFLICT (id) DO NOTHING;

-- Insert 3 tasks with different statuses
INSERT INTO tasks (id, project_id, title, description, status, priority, assignee_id, due_date, creator_id, created_at, updated_at)
VALUES 
    (
        'c2ggde99-9c0b-4ef8-bb6d-6bb9bd380a33',
        'b1ffcd99-9c0b-4ef8-bb6d-6bb9bd380a22',
        'Setup development environment',
        'Install dependencies and configure local environment',
        'done',
        'high',
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        '2026-04-10',
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        NOW(),
        NOW()
    ),
    (
        'd3hhef99-9c0b-4ef8-bb6d-6bb9bd380a44',
        'b1ffcd99-9c0b-4ef8-bb6d-6bb9bd380a22',
        'Implement authentication',
        'Build JWT-based authentication system',
        'in_progress',
        'high',
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        '2026-04-20',
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        NOW(),
        NOW()
    ),
    (
        'e4iifg99-9c0b-4ef8-bb6d-6bb9bd380a55',
        'b1ffcd99-9c0b-4ef8-bb6d-6bb9bd380a22',
        'Write API documentation',
        'Document all REST endpoints with examples',
        'todo',
        'medium',
        NULL,
        '2026-04-25',
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        NOW(),
        NOW()
    )
ON CONFLICT (id) DO NOTHING;
