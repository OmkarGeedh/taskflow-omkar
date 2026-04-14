package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"taskflow/backend/internal/database"
	"taskflow/backend/internal/models"
)

type TaskCreateInput struct {
	Title       string     `json:"title"       binding:"required"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssigneeID  *string    `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date"`
}

type TaskUpdateInput struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Status      *string    `json:"status"`
	Priority    *string    `json:"priority"`
	AssigneeID  *string    `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date"`
}

// GET /projects/:id/tasks?status=&assignee=
func ListTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")
		status := c.Query("status")
		assignee := c.Query("assignee")

		query := `
			SELECT id, project_id, title, description, status, priority,
			       assignee_id, due_date, creator_id, created_at, updated_at
			FROM tasks WHERE project_id = $1`
		args := []interface{}{projectID}
		idx := 2

		if status != "" {
			query += fmt.Sprintf(" AND status = $%d", idx)
			args = append(args, status)
			idx++
		}
		if assignee != "" {
			query += fmt.Sprintf(" AND assignee_id = $%d", idx)
			args = append(args, assignee)
		}

		query += " ORDER BY created_at DESC"

		rows, err := database.DB.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		defer rows.Close()

		tasks := []models.Task{}
		for rows.Next() {
			var t models.Task
			if err := rows.Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description,
				&t.Status, &t.Priority, &t.AssigneeID, &t.DueDate,
				&t.CreatorID, &t.CreatedAt, &t.UpdatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan task"})
				return
			}
			tasks = append(tasks, t)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"tasks": tasks})
	}
}

// POST /projects/:id/tasks
func CreateTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		projectID := c.Param("id")

		var input TaskCreateInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if input.Status == "" {
			input.Status = "todo"
		}
		if input.Priority == "" {
			input.Priority = "medium"
		}

		var t models.Task
		err := database.DB.QueryRow(`
			INSERT INTO tasks
			  (project_id, title, description, status, priority, assignee_id, due_date, creator_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id, project_id, title, description, status, priority,
			          assignee_id, due_date, creator_id, created_at, updated_at
		`, projectID, input.Title, input.Description, input.Status, input.Priority,
			input.AssigneeID, input.DueDate, userID).
			Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description,
				&t.Status, &t.Priority, &t.AssigneeID, &t.DueDate,
				&t.CreatorID, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		c.JSON(http.StatusCreated, t)
	}
}

// PATCH /tasks/:id
func UpdateTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("id")

		var input TaskUpdateInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// COALESCE keeps existing value when the field is null (not sent by client)
		var t models.Task
		err := database.DB.QueryRow(`
			UPDATE tasks SET
				title       = COALESCE($1, title),
				description = COALESCE($2, description),
				status      = COALESCE($3, status),
				priority    = COALESCE($4, priority),
				assignee_id = COALESCE($5, assignee_id),
				due_date    = COALESCE($6, due_date),
				updated_at  = NOW()
			WHERE id = $7
			RETURNING id, project_id, title, description, status, priority,
			          assignee_id, due_date, creator_id, created_at, updated_at
		`, input.Title, input.Description, input.Status, input.Priority,
			input.AssigneeID, input.DueDate, taskID).
			Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description,
				&t.Status, &t.Priority, &t.AssigneeID, &t.DueDate,
				&t.CreatorID, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}

		c.JSON(http.StatusOK, t)
	}
}

// DELETE /tasks/:id — project owner OR task creator
func DeleteTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		taskID := c.Param("id")

		var creatorID, projectID string
		err := database.DB.QueryRow(
			`SELECT creator_id, project_id FROM tasks WHERE id = $1`, taskID,
		).Scan(&creatorID, &projectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}

		var projectOwnerID string
		if err := database.DB.QueryRow(
			`SELECT owner_id FROM projects WHERE id = $1`, projectID,
		).Scan(&projectOwnerID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		if userID != creatorID && userID != projectOwnerID {
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to delete this task"})
			return
		}

		if _, err := database.DB.Exec(`DELETE FROM tasks WHERE id = $1`, taskID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}
