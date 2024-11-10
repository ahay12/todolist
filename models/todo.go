package models

import (
	"database/sql"
	_ "github.com/go-playground/validator/v10"
)

type TodoList struct {
	ID          int
	Title       string
	Description string
	Status      string
	DueDate     sql.NullTime
}

//func (todo *TodoList) GetFormattedDueDate() map[string]interface{} {
//	if todo.DueDate.Valid {
//		return map[string]interface{}{
//			"Time":  todo.DueDate.Time.Format("2006-01-02"),
//			"Valid": true,
//		}
//	}
//	return map[string]interface{}{
//		"Time":  "0001-01-01",
//		"Valid": false,
//	}
//}
