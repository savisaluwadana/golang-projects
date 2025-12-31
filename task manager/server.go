package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateTaskRequest struct {
	Description    string   `json:"description"`
	ProjectID      int      `json:"project_id"`
	Category       string   `json:"category,omitempty"`
	Priority       string   `json:"priority,omitempty"`
	Status         string   `json:"status,omitempty"`
	DueDate        string   `json:"due_date,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	Assignee       string   `json:"assignee,omitempty"`
	EstimatedHours float64  `json:"estimated_hours,omitempty"`
}

type UpdateTaskRequest struct {
	Description    string   `json:"description,omitempty"`
	ProjectID      int      `json:"project_id,omitempty"`
	Category       string   `json:"category,omitempty"`
	Priority       string   `json:"priority,omitempty"`
	Status         string   `json:"status,omitempty"`
	DueDate        string   `json:"due_date,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	Assignee       string   `json:"assignee,omitempty"`
	EstimatedHours float64  `json:"estimated_hours,omitempty"`
	Position       int      `json:"position,omitempty"`
}

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

type TimeTrackingRequest struct {
	TaskID int    `json:"task_id"`
	Note   string `json:"note,omitempty"`
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "message": "Error encoding response"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func handleGetTasks(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	tasks, err := loadTasks()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load tasks",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    tasks,
	})
}

func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if strings.TrimSpace(req.Description) == "" {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Description cannot be empty",
		})
		return
	}

	priority := Medium
	if req.Priority != "" {
		priority = parsePriority(req.Priority)
	}

	status := StatusTodo
	if req.Status != "" {
		status = TaskStatus(req.Status)
	}

	var dueDate *time.Time
	if req.DueDate != "" {
		parsed, err := parseDate(req.DueDate)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid due date format",
			})
			return
		}
		dueDate = parsed
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load data",
		})
		return
	}

	projectID := req.ProjectID
	if projectID == 0 {
		// Create or get default project
		var defaultProject *Project
		for i := range appData.Projects {
			if appData.Projects[i].Name == "Default" {
				defaultProject = &appData.Projects[i]
				break
			}
		}
		if defaultProject == nil {
			newProject := Project{
				ID:        nextProjectID(appData.Projects),
				Name:      "Default",
				Color:     "#6366f1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			appData.Projects = append(appData.Projects, newProject)
			defaultProject = &newProject
		}
		projectID = defaultProject.ID
	}

	task := Task{
		ID:             nextID(appData.Tasks),
		ProjectID:      projectID,
		Description:    req.Description,
		Category:       req.Category,
		Priority:       priority,
		Status:         status,
		DueDate:        dueDate,
		Tags:           req.Tags,
		Assignee:       req.Assignee,
		EstimatedHours: req.EstimatedHours,
		Done:           false,
		CreatedAt:      time.Now(),
		Position:       len(appData.Tasks),
	}

	appData.Tasks = append(appData.Tasks, task)
	if err := saveAppData(appData); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to save task",
		})
		return
	}

	respondJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Task created successfully",
		Data:    task,
	})
}

func handleMarkDone(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	idStr = strings.TrimSuffix(idStr, "/done")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid task ID",
		})
		return
	}

	tasks, err := loadTasks()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load tasks",
		})
		return
	}

	found := false
	now := time.Now()
	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Done = true
			tasks[i].CompletedAt = &now
			found = true
			break
		}
	}

	if !found {
		respondJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Task not found",
		})
		return
	}

	if err := saveTasks(tasks); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to save tasks",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Task marked as done",
	})
}

func handleMarkUndone(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	idStr = strings.TrimSuffix(idStr, "/undone")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid task ID",
		})
		return
	}

	tasks, err := loadTasks()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load tasks",
		})
		return
	}

	found := false
	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Done = false
			tasks[i].CompletedAt = nil
			found = true
			break
		}
	}

	if !found {
		respondJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Task not found",
		})
		return
	}

	if err := saveTasks(tasks); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to save tasks",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Task marked as undone",
	})
}

func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid task ID",
		})
		return
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load data",
		})
		return
	}

	out := appData.Tasks[:0]
	found := false
	for _, t := range appData.Tasks {
		if t.ID == id {
			found = true
			continue
		}
		out = append(out, t)
	}

	if !found {
		respondJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Task not found",
		})
		return
	}

	appData.Tasks = out
	if err := saveAppData(appData); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to save data",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Task deleted successfully",
	})
}

func handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid task ID",
		})
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load data",
		})
		return
	}

	found := false
	for i := range appData.Tasks {
		if appData.Tasks[i].ID == id {
			// Update description
			if req.Description != "" {
				appData.Tasks[i].Description = req.Description
			}

			// Update project
			if req.ProjectID > 0 {
				appData.Tasks[i].ProjectID = req.ProjectID
			}

			// Update category
			if req.Category != "" {
				appData.Tasks[i].Category = req.Category
			}

			// Update priority
			if req.Priority != "" {
				appData.Tasks[i].Priority = parsePriority(req.Priority)
			}

			// Update status
			if req.Status != "" {
				appData.Tasks[i].Status = TaskStatus(req.Status)
			}

			// Update assignee
			if req.Assignee != "" {
				appData.Tasks[i].Assignee = req.Assignee
			}

			// Update estimated hours
			if req.EstimatedHours > 0 {
				appData.Tasks[i].EstimatedHours = req.EstimatedHours
			}

			// Update position
			if req.Position >= 0 {
				appData.Tasks[i].Position = req.Position
			}

			// Update due date
			if req.DueDate != "" {
				if req.DueDate == "null" || req.DueDate == "clear" {
					appData.Tasks[i].DueDate = nil
				} else {
					parsed, err := parseDate(req.DueDate)
					if err != nil {
						respondJSON(w, http.StatusBadRequest, APIResponse{
							Success: false,
							Message: "Invalid due date format",
						})
						return
					}
					appData.Tasks[i].DueDate = parsed
				}
			}

			// Update tags
			if req.Tags != nil {
				appData.Tasks[i].Tags = req.Tags
			}

			found = true
			break
		}
	}

	if !found {
		respondJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Task not found",
		})
		return
	}

	if err := saveAppData(appData); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to save data",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Task updated successfully",
	})
}

func handleGetStats(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load data",
		})
		return
	}

	tasks := appData.Tasks
	projects := appData.Projects
	timeEntries := appData.TimeEntries

	totalTasks := len(tasks)
	completedTasks := 0
	overdue := 0
	byPriority := map[string]int{"low": 0, "medium": 0, "high": 0, "urgent": 0}
	categories := make(map[string]int)

	for _, t := range tasks {
		if t.Done {
			completedTasks++
		} else {
			byPriority[strings.ToLower(t.Priority.String())]++
			if t.DueDate != nil && t.DueDate.Before(time.Now()) {
				overdue++
			}
		}
		if t.Category != "" {
			categories[t.Category]++
		}
	}

	// Calculate total hours tracked
	var totalHours float64
	for _, entry := range timeEntries {
		if entry.EndTime != nil {
			duration := entry.EndTime.Sub(entry.StartTime)
			totalHours += duration.Hours()
		}
	}

	stats := map[string]interface{}{
		"total_tasks":         totalTasks,
		"completed_tasks":     completedTasks,
		"pending_tasks":       totalTasks - completedTasks,
		"overdue_tasks":       overdue,
		"total_projects":      len(projects),
		"total_hours_tracked": totalHours,
		"byPriority":          byPriority,
		"categories":          categories,
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    stats,
	})
}

// Project Handlers
func handleGetProjects(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load projects",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    appData.Projects,
	})
}

func handleCreateProject(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Project name cannot be empty",
		})
		return
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load data",
		})
		return
	}

	color := req.Color
	if color == "" {
		color = "#6366f1"
	}

	project := Project{
		ID:          nextProjectID(appData.Projects),
		Name:        req.Name,
		Description: req.Description,
		Color:       color,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	appData.Projects = append(appData.Projects, project)
	if err := saveAppData(appData); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to save project",
		})
		return
	}

	respondJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Project created successfully",
		Data:    project,
	})
}

func handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid project ID",
		})
		return
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load data",
		})
		return
	}

	// Remove project
	out := appData.Projects[:0]
	found := false
	for _, p := range appData.Projects {
		if p.ID == id {
			found = true
			continue
		}
		out = append(out, p)
	}

	if !found {
		respondJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Project not found",
		})
		return
	}

	// Also remove tasks in this project
	tasks := appData.Tasks[:0]
	for _, t := range appData.Tasks {
		if t.ProjectID != id {
			tasks = append(tasks, t)
		}
	}

	appData.Projects = out
	appData.Tasks = tasks
	if err := saveAppData(appData); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to save data",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Project deleted successfully",
	})
}

func handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid project ID",
		})
		return
	}

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if req.Name == "" {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Project name is required",
		})
		return
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load data",
		})
		return
	}

	// Find and update project
	found := false
	for i, p := range appData.Projects {
		if p.ID == id {
			appData.Projects[i].Name = req.Name
			appData.Projects[i].Description = req.Description
			if req.Color != "" {
				appData.Projects[i].Color = req.Color
			}
			found = true
			break
		}
	}

	if !found {
		respondJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Project not found",
		})
		return
	}

	if err := saveAppData(appData); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to save data",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Project updated successfully",
	})
}

// Kanban Handlers
func handleGetKanban(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	projectIDStr := r.URL.Query().Get("project_id")
	var projectID int
	if projectIDStr != "" {
		var err error
		projectID, err = strconv.Atoi(projectIDStr)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid project ID",
			})
			return
		}
	}

	tasks, err := loadTasks()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load tasks",
		})
		return
	}

	// Filter by project if specified
	if projectID > 0 {
		filtered := []Task{}
		for _, t := range tasks {
			if t.ProjectID == projectID {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

	// Group tasks by status
	kanban := map[string][]Task{
		"backlog":     {},
		"todo":        {},
		"in_progress": {},
		"in_review":   {},
		"done":        {},
	}

	for _, task := range tasks {
		status := string(task.Status)
		if status == "" {
			if task.Done {
				status = "done"
			} else {
				status = "todo"
			}
		}
		kanban[status] = append(kanban[status], task)
	}

	// Sort by position within each column
	for status := range kanban {
		sort.Slice(kanban[status], func(i, j int) bool {
			return kanban[status][i].Position < kanban[status][j].Position
		})
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    kanban,
	})
}

func handleMoveTask(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		TaskID    int    `json:"task_id"`
		NewStatus string `json:"new_status"`
		Position  int    `json:"position"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load data",
		})
		return
	}

	// Update task status and position
	found := false
	for i := range appData.Tasks {
		if appData.Tasks[i].ID == req.TaskID {
			appData.Tasks[i].Status = TaskStatus(req.NewStatus)
			appData.Tasks[i].Position = req.Position
			if req.NewStatus == "done" {
				appData.Tasks[i].Done = true
				now := time.Now()
				appData.Tasks[i].CompletedAt = &now
			} else {
				appData.Tasks[i].Done = false
				appData.Tasks[i].CompletedAt = nil
			}
			found = true
			break
		}
	}

	if !found {
		respondJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Task not found",
		})
		return
	}

	if err := saveAppData(appData); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to save data",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Task moved successfully",
	})
}

// Time Tracking Handlers
func handleStartTimer(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req TimeTrackingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load data",
		})
		return
	}

	// Check if there's already an active timer
	for _, entry := range appData.TimeEntries {
		if entry.EndTime == nil {
			respondJSON(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "There's already an active timer running",
			})
			return
		}
	}

	timeEntry := TimeEntry{
		ID:        nextTimeEntryID(appData.TimeEntries),
		TaskID:    req.TaskID,
		StartTime: time.Now(),
		Note:      req.Note,
	}

	appData.TimeEntries = append(appData.TimeEntries, timeEntry)
	if err := saveAppData(appData); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to save time entry",
		})
		return
	}

	respondJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Timer started",
		Data:    timeEntry,
	})
}

func handleStopTimer(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/time/")
	idStr = strings.TrimSuffix(idStr, "/stop")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid time entry ID",
		})
		return
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load data",
		})
		return
	}

	found := false
	now := time.Now()
	for i := range appData.TimeEntries {
		if appData.TimeEntries[i].ID == id {
			appData.TimeEntries[i].EndTime = &now
			duration := int(now.Sub(appData.TimeEntries[i].StartTime).Seconds())
			appData.TimeEntries[i].Duration = duration
			found = true
			break
		}
	}

	if !found {
		respondJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Time entry not found",
		})
		return
	}

	if err := saveAppData(appData); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to save data",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Timer stopped",
	})
}

func handleGetTimeEntries(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	taskIDStr := r.URL.Query().Get("task_id")
	var taskID int
	if taskIDStr != "" {
		var err error
		taskID, err = strconv.Atoi(taskIDStr)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid task ID",
			})
			return
		}
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load data",
		})
		return
	}

	entries := appData.TimeEntries
	if taskID > 0 {
		filtered := []TimeEntry{}
		for _, e := range entries {
			if e.TaskID == taskID {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    entries,
	})
}

// Reports Handler
func handleGetReports(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load data",
		})
		return
	}

	// Calculate various metrics
	totalTasks := len(appData.Tasks)
	completedTasks := 0
	totalEstimatedHours := 0.0
	totalTrackedSeconds := 0
	tasksByProject := make(map[string]int)
	tasksByStatus := make(map[string]int)
	tasksByPriority := make(map[string]int)
	overdueTasks := 0

	for _, task := range appData.Tasks {
		if task.Done {
			completedTasks++
		}
		totalEstimatedHours += task.EstimatedHours

		// Count by status
		status := string(task.Status)
		if status == "" {
			status = "todo"
		}
		tasksByStatus[status]++

		// Count by priority
		tasksByPriority[strings.ToLower(task.Priority.String())]++

		// Check overdue
		if !task.Done && task.DueDate != nil && task.DueDate.Before(time.Now()) {
			overdueTasks++
		}

		// Find project name
		for _, proj := range appData.Projects {
			if proj.ID == task.ProjectID {
				tasksByProject[proj.Name]++
				break
			}
		}
	}

	// Calculate time tracking stats
	for _, entry := range appData.TimeEntries {
		totalTrackedSeconds += entry.Duration
	}

	completionRate := 0.0
	if totalTasks > 0 {
		completionRate = float64(completedTasks) / float64(totalTasks) * 100
	}

	report := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_tasks":           totalTasks,
			"completed_tasks":       completedTasks,
			"completion_rate":       completionRate,
			"overdue_tasks":         overdueTasks,
			"total_estimated_hours": totalEstimatedHours,
			"total_tracked_hours":   float64(totalTrackedSeconds) / 3600,
			"total_projects":        len(appData.Projects),
		},
		"by_project":  tasksByProject,
		"by_status":   tasksByStatus,
		"by_priority": tasksByPriority,
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    report,
	})
}

// Comment Handlers
func handleGetComments(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "task_id is required",
		})
		return
	}

	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid task ID",
		})
		return
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load data",
		})
		return
	}

	// Find task and return its comments
	for _, task := range appData.Tasks {
		if task.ID == taskID {
			comments := task.Comments
			if comments == nil {
				comments = []Comment{}
			}
			respondJSON(w, http.StatusOK, APIResponse{
				Success: true,
				Data:    comments,
			})
			return
		}
	}

	respondJSON(w, http.StatusNotFound, APIResponse{
		Success: false,
		Message: "Task not found",
	})
}

func handleAddComment(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		TaskID int    `json:"task_id"`
		Author string `json:"author"`
		Text   string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if req.TaskID == 0 || req.Text == "" {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "task_id and text are required",
		})
		return
	}

	appData, err := loadAppData()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to load data",
		})
		return
	}

	// Find task and add comment
	found := false
	for i, task := range appData.Tasks {
		if task.ID == req.TaskID {
			// Generate comment ID
			commentID := 1
			if len(task.Comments) > 0 {
				for _, c := range task.Comments {
					if c.ID >= commentID {
						commentID = c.ID + 1
					}
				}
			}

			comment := Comment{
				ID:        commentID,
				TaskID:    req.TaskID,
				Author:    req.Author,
				Text:      req.Text,
				CreatedAt: time.Now(),
			}

			if appData.Tasks[i].Comments == nil {
				appData.Tasks[i].Comments = []Comment{}
			}
			appData.Tasks[i].Comments = append(appData.Tasks[i].Comments, comment)
			found = true
			break
		}
	}

	if !found {
		respondJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Task not found",
		})
		return
	}

	if err := saveAppData(appData); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to save data",
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Comment added successfully",
	})
}

func routeHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	path := r.URL.Path

	switch {
	// Task endpoints
	case path == "/api/tasks" && r.Method == "GET":
		handleGetTasks(w, r)
	case path == "/api/tasks" && r.Method == "POST":
		handleCreateTask(w, r)
	case strings.HasPrefix(path, "/api/tasks/") && strings.HasSuffix(path, "/done") && r.Method == "PUT":
		handleMarkDone(w, r)
	case strings.HasPrefix(path, "/api/tasks/") && strings.HasSuffix(path, "/undone") && r.Method == "PUT":
		handleMarkUndone(w, r)
	case strings.HasPrefix(path, "/api/tasks/") && r.Method == "DELETE":
		handleDeleteTask(w, r)
	case strings.HasPrefix(path, "/api/tasks/") && r.Method == "PUT":
		handleUpdateTask(w, r)

	// Project endpoints
	case path == "/api/projects" && r.Method == "GET":
		handleGetProjects(w, r)
	case path == "/api/projects" && r.Method == "POST":
		handleCreateProject(w, r)
	case strings.HasPrefix(path, "/api/projects/") && r.Method == "PUT":
		handleUpdateProject(w, r)
	case strings.HasPrefix(path, "/api/projects/") && r.Method == "DELETE":
		handleDeleteProject(w, r)

	// Kanban endpoints
	case path == "/api/kanban" && r.Method == "GET":
		handleGetKanban(w, r)
	case path == "/api/kanban/move" && r.Method == "PUT":
		handleMoveTask(w, r)

	// Time tracking endpoints
	case path == "/api/time/start" && r.Method == "POST":
		handleStartTimer(w, r)
	case strings.HasPrefix(path, "/api/time/") && strings.HasSuffix(path, "/stop") && r.Method == "PUT":
		handleStopTimer(w, r)
	case path == "/api/time" && r.Method == "GET":
		handleGetTimeEntries(w, r)

	// Reports and stats
	case path == "/api/stats" && r.Method == "GET":
		handleGetStats(w, r)
	case path == "/api/reports" && r.Method == "GET":
		handleGetReports(w, r)

	// Comment endpoints
	case path == "/api/comments" && r.Method == "GET":
		handleGetComments(w, r)
	case path == "/api/comments" && r.Method == "POST":
		handleAddComment(w, r)

	default:
		respondJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Endpoint not found",
		})
	}
}

func startServer() {
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	http.HandleFunc("/api/", routeHandler)

	port := "8080"
	fmt.Printf("\n🚀 Server starting on http://localhost:%s\n", port)
	fmt.Printf("📋 Task Manager UI: http://localhost:%s\n", port)
	fmt.Printf("🔌 API endpoint: http://localhost:%s/api\n\n", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
