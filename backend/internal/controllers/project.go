package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"taskflow/backend/internal/database"
	"taskflow/backend/internal/models"
)

type ProjectInput struct {
	Name        string `json:"name"        binding:"required"`
	Description string `json:"description"`
}

// GET /projects — returns projects the user owns OR has tasks assigned to them in.
func ListProjects() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		rows, err := database.DB.Query(`
			SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at
			FROM projects p
			LEFT JOIN tasks t ON t.project_id = p.id
			WHERE p.owner_id = $1 OR t.assignee_id = $1
			ORDER BY p.created_at DESC
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		defer rows.Close()

		projects := []models.Project{}
		for rows.Next() {
			var p models.Project
			if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan project"})
				return
			}
			projects = append(projects, p)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"projects": projects})
	}
}

// POST /projects
func CreateProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var input ProjectInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var p models.Project
		err := database.DB.QueryRow(`
			INSERT INTO projects (name, description, owner_id)
			VALUES ($1, $2, $3)
			RETURNING id, name, description, owner_id, created_at
		`, input.Name, input.Description, userID).
			Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		c.JSON(http.StatusCreated, p)
	}
}

// GET /projects/:id — returns project + its tasks
func GetProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")

		var p models.Project
		err := database.DB.QueryRow(`
			SELECT id, name, description, owner_id, created_at
			FROM projects WHERE id = $1
		`, projectID).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}

		rows, err := database.DB.Query(`
			SELECT id, project_id, title, description, status, priority,
			       assignee_id, due_date, creator_id, created_at, updated_at
			FROM tasks WHERE project_id = $1 ORDER BY created_at DESC
		`, projectID)
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

		// Flatten response: merge project fields with tasks array
		c.JSON(http.StatusOK, gin.H{
			"id":          p.ID,
			"name":        p.Name,
			"description": p.Description,
			"owner_id":    p.OwnerID,
			"created_at":  p.CreatedAt,
			"tasks":       tasks,
		})
	}
}

// PATCH /projects/:id — owner only
func UpdateProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		projectID := c.Param("id")

		if err := assertProjectOwner(projectID, userID, c); err != nil {
			return
		}

		var input ProjectInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var p models.Project
		err := database.DB.QueryRow(`
			UPDATE projects SET name = $1, description = $2
			WHERE id = $3
			RETURNING id, name, description, owner_id, created_at
		`, input.Name, input.Description, projectID).
			Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		c.JSON(http.StatusOK, p)
	}
}

// DELETE /projects/:id — owner only; cascades tasks via DB constraint
func DeleteProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		projectID := c.Param("id")

		if err := assertProjectOwner(projectID, userID, c); err != nil {
			return
		}

		if _, err := database.DB.Exec(`DELETE FROM projects WHERE id = $1`, projectID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// assertProjectOwner checks ownership, writes 404/403 and returns non-nil error on failure.
func assertProjectOwner(projectID string, userID string, c *gin.Context) error {
	var ownerID string
	err := database.DB.QueryRow(`SELECT owner_id FROM projects WHERE id = $1`, projectID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return err
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the project owner can do this"})
		return fmt.Errorf("forbidden")
	}
	return nil
}
