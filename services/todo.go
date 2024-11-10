package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"todolist/database"
	"todolist/models"
)

type PaginatedTodos struct {
	Todos       []models.TodoList `json:"tasks"`
	CurrentPage int               `json:"current_page"`
	TotalPages  int               `json:"total_pages"`
	TotalTasks  int               `json:"total_tasks"`
}

type PaginationInfo struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	TotalTasks  int `json:"total_tasks"`
}

type TodoInput struct {
	Title       string       `validate:"required,min=3,max=100" json:"title"`
	Description string       `validate:"required" json:"description"`
	Status      string       `validate:"required,oneof=pending completed" json:"status"`
	DueDate     sql.NullTime `validate:"required" json:"due_date"`
}

var (
	validate        = validator.New()
	todoCacheKey    = "todos:all"
	todoByIDCache   = "todo:id:%s"
	cacheExpiration = time.Hour * 1
)

func GetAllTodos(ctx context.Context, page, limit int) (*PaginatedTodos, error) {
	cacheKey := fmt.Sprintf("todos:page:%d:limit:%d", page, limit)
	cacheData, err := database.RedisClient.Get(ctx, cacheKey).Result()

	if errors.Is(err, redis.Nil) {
		// Cache miss, fetch from database
		todos, pagination, err := fetchPaginatedTodosFromDB(page, limit)
		if err != nil {
			return nil, err
		}

		// Cache the retrieved todos
		err = cacheTodos(ctx, cacheKey, todos, pagination)
		if err != nil {
			log.Println("Failed to cache todos:", err)
		}

		return &PaginatedTodos{Todos: todos, CurrentPage: pagination.CurrentPage, TotalPages: pagination.TotalPages, TotalTasks: pagination.TotalTasks}, nil
	} else if err != nil {
		return nil, err
	}

	// Unmarshal cached data
	var paginatedTodos PaginatedTodos
	err = json.Unmarshal([]byte(cacheData), &paginatedTodos)
	if err != nil {
		return nil, err
	}

	return &paginatedTodos, nil
}

// fetchPaginatedTodosFromDB retrieves todos from the database based on pagination
func fetchPaginatedTodosFromDB(page, limit int) ([]models.TodoList, PaginationInfo, error) {
	startRow := (page - 1) * limit
	query := `
        SELECT id, title, description, status, due_date
        FROM (
            SELECT id, title, description, status, due_date, ROW_NUMBER() OVER (ORDER BY id) AS rn FROM todolist
        ) WHERE rn BETWEEN :1 AND :2
    `
	rows, err := database.DB.Query(query, startRow+1, startRow+limit)
	if err != nil {
		return nil, PaginationInfo{}, err
	}

	defer rows.Close()

	var todos []models.TodoList
	for rows.Next() {
		var todo models.TodoList
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Status, &todo.DueDate); err != nil {
			return nil, PaginationInfo{}, err
		}
		todos = append(todos, todo)
	}

	// Calculate pagination info
	var totalTasks int
	countQuery := `SELECT COUNT(*) FROM todolist`
	if err := database.DB.QueryRow(countQuery).Scan(&totalTasks); err != nil {
		return nil, PaginationInfo{}, err
	}

	totalPages := (totalTasks + limit - 1) / limit
	pagination := PaginationInfo{
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalTasks:  totalTasks,
	}

	return todos, pagination, nil
}

func cacheTodos(ctx context.Context, key string, todos []models.TodoList, pagination PaginationInfo) error {
	data := PaginatedTodos{
		Todos:       todos,
		CurrentPage: pagination.CurrentPage,
		TotalPages:  pagination.TotalPages,
		TotalTasks:  pagination.TotalTasks,
	}

	cacheData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return database.RedisClient.Set(ctx, key, cacheData, cacheExpiration).Err()
}

func GetTodoByID(ctx context.Context, id string) (*models.TodoList, error) {
	cacheKey := fmt.Sprintf(todoByIDCache, id)
	cacheData, err := database.RedisClient.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		// Cache miss, query database
		todo, err := fetchTodoByIDFromDB(id)
		if err != nil {
			return nil, err
		}

		todoJSON, err := json.Marshal(todo)
		if err == nil {
			database.RedisClient.Set(ctx, cacheKey, todoJSON, cacheExpiration)
		}

		return todo, nil
	} else if err != nil {
		return nil, err
	}

	// Cache hit
	var todo models.TodoList
	err = json.Unmarshal([]byte(cacheData), &todo)
	return &todo, err
}

func fetchTodoByIDFromDB(id string) (*models.TodoList, error) {
	query := `SELECT id, title, description, status, due_date FROM todolist WHERE id = :1`
	var todo models.TodoList
	err := database.DB.QueryRow(query, id).Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Status, &todo.DueDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &todo, nil
}

func CreateTodo(todo *models.TodoList) (*models.TodoList, error) {
	// Validate the input struct
	input := TodoInput{
		Title:       todo.Title,
		Description: todo.Description,
		Status:      todo.Status,
		DueDate:     todo.DueDate,
	}
	if err := validate.Struct(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	var query string
	var args []interface{}

	if todo.DueDate.Valid {
		// If due date is provided, include it in the query
		query = `INSERT INTO todolist (title, description, status, due_date) 
		         VALUES (:1, :2, :3, TO_DATE(:4, 'YYYY-MM-DD')) RETURNING id INTO :5`
		args = []interface{}{todo.Title, todo.Description, todo.Status, todo.DueDate.Time.Format("2006-01-02"), sql.Out{Dest: &todo.ID}}
	} else {
		// If due date is not provided, omit it from the query
		query = `INSERT INTO todolist (title, description, status) 
		         VALUES (:1, :2, :3) RETURNING id INTO :4`
		args = []interface{}{todo.Title, todo.Description, todo.Status, sql.Out{Dest: &todo.ID}}
	}

	// Execute the query
	_, err := database.DB.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return todo, nil
}

// UpdateTodoByID validates and updates a todo item by ID
func UpdateTodoByID(ctx context.Context, id string, todo *models.TodoList) (*models.TodoList, error) {
	input := TodoInput{
		Title:       todo.Title,
		Description: todo.Description,
		Status:      todo.Status,
		DueDate:     todo.DueDate,
	}
	if err := validate.Struct(input); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	query := `UPDATE todolist SET title = :1, description = :2, status = :3, due_date = :4 WHERE id = :5`
	_, err := database.DB.Exec(query, todo.Title, todo.Description, todo.Status, todo.DueDate, id)
	return todo, err
}

// DeleteTodoByID deletes a todo item by ID
func DeleteTodoByID(ctx context.Context, id string) error {
	query := `DELETE FROM todolist WHERE id = :1`
	_, err := database.DB.Exec(query, id)
	return err
}
