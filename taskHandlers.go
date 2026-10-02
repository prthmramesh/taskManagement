package main

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
)

func createTask(w http.ResponseWriter, r *http.Request) {

	var task Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		slog.Warn("failed to decode request body", "error", err)
		return
	}

	if task.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		slog.Warn("title empty", "title", task.Title)
		return
	}

	if task.Status == "" {
		task.Status = "pending"
	}

	if task.Status != "pending" && task.Status != "in_progress" && task.Status != "completed" {
		http.Error(w, "Invalid status value", http.StatusBadRequest)
		slog.Warn("invalid status value", "status", task.Status)
		return
	}

	task.UserID = r.Context().Value(userIDKey).(int)

	sqlstatement := `INSERT INTO tasks (user_id, title, description, status, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id, created_at, updated_at`
	err := db.QueryRow(sqlstatement, task.UserID, task.Title, task.Description, task.Status).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		slog.Error("failed to create task", "error", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	slog.Info("task created successfully", "task_id", task.ID, "user_id", task.UserID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	UserID := r.Context().Value(userIDKey).(int)

	row, err := db.Query("SELECT id, user_id, title, description, status, created_at, updated_at FROM tasks WHERE user_id=$1", UserID)
	if err != nil {
		slog.Error("failed to retrieve tasks", "error", err)
		http.Error(w, "Failed to retrieve tasks", http.StatusInternalServerError)
		return
	}
	defer row.Close()

	tasks := make([]Task, 0)

	for row.Next() {
		var task Task
		if err := row.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt); err != nil {
			slog.Error("failed to scan task", "error", err)
			http.Error(w, "Failed to retrieve tasks", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	slog.Info("tasks fetched", "user_id", UserID, "count", len(tasks))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

func getTaskByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)

	taskID := r.PathValue("id")

	var task Task
	err := db.QueryRow("SELECT id, user_id, title, description, status, created_at, updated_at FROM tasks WHERE id=$1 AND user_id=$2", taskID, userID).Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {

		if err == sql.ErrNoRows {
			slog.Warn("task not found", "task_id", taskID, "user_id", userID)
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, "Failed to fetch task", http.StatusInternalServerError)
			slog.Error("get task failed", "user_id", userID, "task_id", taskID, "error", err.Error())
		}
		return

	}

	slog.Info("task fetched", "user_id", userID, "task_id", taskID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}
