package handler

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"todolist/helper"
	"todolist/models"
	"todolist/services"
)

func CreateUserHandler(ctx *fiber.Ctx) error {
	user := new(models.User)
	if err := ctx.BodyParser(user); err != nil {
		helper.RespondJSON(ctx, fiber.StatusBadRequest, "cannot parse JSON", nil, err.Error())
		return err
	}

	data, err := services.CreateUser(user)
	if err != nil {
		helper.RespondJSON(ctx, fiber.StatusInternalServerError, "failed to create user", nil, err.Error())
		return err
	}

	helper.RespondJSON(ctx, fiber.StatusCreated, "User created successfully", data, nil)
	return nil
}

func GetAllTodosHandler(c *fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	paginatedTodos, err := services.GetAllTodos(c.Context(), page, limit)
	if err != nil {
		helper.RespondJSON(c, fiber.StatusInternalServerError, "Failed to get todos", nil, err.Error())
		return err
	}

	return c.Status(fiber.StatusOK).JSON(paginatedTodos)
}

func GetTodoByIDHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	todo, err := services.GetTodoByID(context.Background(), id)
	if err != nil {
		helper.RespondJSON(c, fiber.StatusInternalServerError, "Failed to get todo", nil, err.Error())
		return err
	}
	return c.JSON(fiber.Map{"todo": todo})
}

func CreateTodoHandler(c *fiber.Ctx) error {
	var todo models.TodoList
	if err := c.BodyParser(&todo); err != nil {
		helper.RespondJSON(c, fiber.StatusBadRequest, "Failed to parse request body", nil, err.Error())
		return err
	}

	createdTodo, err := services.CreateTodo(&todo)
	if err != nil {
		helper.RespondJSON(c, fiber.StatusInternalServerError, "Failed to create todo", nil, err.Error())
		return err
	}

	helper.RespondJSON(c, fiber.StatusCreated, "Task created successfully", createdTodo, nil)
	return nil
}

func UpdateTodoHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	var todo models.TodoList
	if err := c.BodyParser(&todo); err != nil {
		helper.RespondJSON(c, fiber.StatusBadRequest, "Failed to parse request body", nil, err.Error())
		return err
	}

	updatedTodo, err := services.UpdateTodoByID(context.Background(), id, &todo)
	if err != nil {
		helper.RespondJSON(c, fiber.StatusInternalServerError, "Failed to update todo", nil, err.Error())
		return err
	}

	helper.RespondJSON(c, fiber.StatusOK, "Todo updated successfully", updatedTodo, nil)
	return nil
}

func DeleteTodoHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	err := services.DeleteTodoByID(context.Background(), id)
	if err != nil {
		helper.RespondJSON(c, fiber.StatusInternalServerError, "Failed to delete todo", nil, err.Error())
		return err
	}

	helper.RespondJSON(c, fiber.StatusOK, "Todo deleted successfully", nil, nil)
	return nil
}
