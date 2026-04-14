# TaskFlow

A task management system built with **Go**, **PostgreSQL**, and **Docker**. Users can register, log in, create projects, add tasks to projects, and assign tasks to themselves or other team members.

---

## Overview

TaskFlow is a RESTful backend service that provides:

- **User authentication** — JWT-based registration and login
- **Project management** — CRUD operations with ownership enforcement
- **Task management** — Create, update, filter, and delete tasks within projects
- **Role-based access control** — Project owners and task creators have distinct permissions

### Tech Stack

Go (Golang), Gin, PostgreSQL, Docker, JWT, bcrypt
---

## Architecture Decisions

### Project Layout

```
taskflow/
├── backend/
│   ├── cmd/api/           # Application entry point (main.go)
│   ├── internal/
│   │   ├── config/        # Environment-based configuration
│   │   ├── controllers/   # HTTP handlers (auth, project, task)
│   │   ├── database/      # Connection pool management
│   │   ├── middleware/     # JWT authentication middleware
│   │   ├── models/        # Data structs (User, Project, Task)
│   │   ├── routes/        # Route registration
│   │   └── security/      # Password hashing (bcrypt, cost 12)
│   ├── Dockerfile         # Multi-stage build
│   ├── entrypoint.sh      # Migration runner + app starter
│   └── go.mod
├── migrations/            # SQL migration files (up + down)
├── docs/                  # Project documentation
├── docker-compose.yml
├── .env.example
└── README.md
```

## Running Locally

> **Prerequisites:** Docker and Docker Compose installed. Nothing else required.

```bash
# 1. Clone the repository
git clone https://github.com/OmkarGeedh/taskflow.git
cd taskflow

# 2. Create environment file
cp .env.example .env

# 3. Start everything
docker compose up --build

# API available at http://localhost:8080
```

The `docker compose up` command will:

1. Start a PostgreSQL 15 container with health checks
2. Wait for Postgres to be ready
3. Run all migrations automatically
4. Seed the database with test data
5. Start the Go API server on port **8080**

### Stopping

```bash
docker compose down           # Stop containers
docker compose down -v        # Stop + remove database volume (fresh start)
```

---

## Running Migrations

**Migrations run automatically on container startup** via `entrypoint.sh`. No manual steps needed.

If you need to run them manually:

```bash
# Connect to the running postgres container
docker exec -it taskflow_db psql -U taskflow -d taskflow

# Or run migration files directly
docker exec -i taskflow_db psql -U taskflow -d taskflow < migrations/20250414_user_table_up.sql
docker exec -i taskflow_db psql -U taskflow -d taskflow < migrations/20250414_projects_table_up.sql
docker exec -i taskflow_db psql -U taskflow -d taskflow < migrations/20250414_task_table_up.sql
```

### Rolling back

```bash
docker exec -i taskflow_db psql -U taskflow -d taskflow < migrations/20250414_task_table_down.sql
docker exec -i taskflow_db psql -U taskflow -d taskflow < migrations/20250414_projects_table_down.sql
docker exec -i taskflow_db psql -U taskflow -d taskflow < migrations/20250414_user_table_down.sql
```

---

## Test Credentials

The seed script creates a test user that can be used immediately:

```
Email:    test@example.com
Password: password123
```

It also creates:
- **1 project** — "Sample Project"
- **3 tasks** — with statuses `done`, `in_progress`, and `todo`

---

## API Reference

All endpoints return `Content-Type: application/json`. Protected endpoints require the `Authorization: Bearer <token>` header.

### Authentication

#### `POST /auth/register`

Register a new user.

```json
// Request
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "secret123"
}

// Response 201
{
  "token": "<jwt>",
  "user": {
    "id": "uuid",
    "name": "Jane Doe",
    "email": "jane@example.com"
  }
}
```

#### `POST /auth/login`

Authenticate and receive a JWT.

```json
// Request
{
  "email": "test@example.com",
  "password": "password123"
}

// Response 200
{
  "token": "<jwt>",
  "user": {
    "id": "uuid",
    "name": "Test User",
    "email": "test@example.com"
  }
}
```

---

### Profile

#### `GET /profile` 🔒

Returns the authenticated user's profile.

```json
// Response 200
{
  "id": "uuid",
  "name": "Test User",
  "email": "test@example.com",
  "created_at": "2026-04-14T10:00:00Z"
}
```

---

### Projects

#### `GET /projects` 🔒

List all projects the user owns or has tasks assigned in.

```json
// Response 200
{
  "projects": [
    {
      "id": "uuid",
      "name": "Sample Project",
      "description": "A test project with sample tasks",
      "owner_id": "uuid",
      "created_at": "2026-04-14T10:00:00Z"
    }
  ]
}
```

#### `POST /projects` 🔒

Create a new project (current user becomes owner).

```json
// Request
{
  "name": "New Project",
  "description": "Optional description"
}

// Response 201
{
  "id": "uuid",
  "name": "New Project",
  "description": "Optional description",
  "owner_id": "uuid",
  "created_at": "2026-04-14T10:00:00Z"
}
```

#### `GET /projects/:id` 🔒

Get project details with all its tasks.

```json
// Response 200
{
  "id": "uuid",
  "name": "Sample Project",
  "description": "A test project with sample tasks",
  "owner_id": "uuid",
  "created_at": "2026-04-14T10:00:00Z",
  "tasks": [
    {
      "id": "uuid",
      "project_id": "uuid",
      "title": "Setup development environment",
      "description": "Install dependencies and configure local environment",
      "status": "done",
      "priority": "high",
      "assignee_id": "uuid",
      "due_date": "2026-04-10T00:00:00Z",
      "creator_id": "uuid",
      "created_at": "2026-04-14T10:00:00Z",
      "updated_at": "2026-04-14T10:00:00Z"
    }
  ]
}
```

#### `PATCH /projects/:id` 🔒

Update project name and/or description. **Owner only.**

```json
// Request
{
  "name": "Updated Name",
  "description": "Updated description"
}

// Response 200 — returns updated project object
```

#### `DELETE /projects/:id` 🔒

Delete project and all its tasks (cascade). **Owner only.**

```
Response: 204 No Content
```

---

### Tasks

#### `GET /projects/:id/tasks` 🔒

List tasks in a project. Supports query filters:

| Parameter  | Description                     |
| ---------- | ------------------------------- |
| `status`   | Filter by `todo`, `in_progress`, or `done` |
| `assignee` | Filter by assignee user UUID    |

```
GET /projects/:id/tasks?status=todo&assignee=uuid
```

```json
// Response 200
{
  "tasks": [
    {
      "id": "uuid",
      "project_id": "uuid",
      "title": "Write API documentation",
      "description": "Document all REST endpoints with examples",
      "status": "todo",
      "priority": "medium",
      "assignee_id": null,
      "due_date": "2026-04-25T00:00:00Z",
      "creator_id": "uuid",
      "created_at": "2026-04-14T10:00:00Z",
      "updated_at": "2026-04-14T10:00:00Z"
    }
  ]
}
```

#### `POST /projects/:id/tasks` 🔒

Create a new task in a project.

```json
// Request
{
  "title": "Design homepage",
  "description": "Create wireframes and mockups",
  "priority": "high",
  "assignee_id": "uuid",
  "due_date": "2026-04-20T00:00:00Z"
}

// Response 201 — returns created task object
```

Defaults: `status = "todo"`, `priority = "medium"`.

#### `PATCH /tasks/:id` 🔒

Partial update — all fields are optional. Only provided fields are updated (uses `COALESCE`).

```json
// Request
{
  "status": "done",
  "priority": "low"
}

// Response 200 — returns updated task object
```

#### `DELETE /tasks/:id` 🔒

Delete a task. Allowed for **project owner** or **task creator** only.

```
Response: 204 No Content
```

---



## License

This project was built as a take-home assignment.
