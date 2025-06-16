package controllers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"diplom/config"
	"diplom/models"
)

func isAdmin(c *fiber.Ctx) bool {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	return ok && claims["role"] == "admin"
}

func ListOperators(c *fiber.Ctx) error {
	if !isAdmin(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа"})
	}

	var ops []models.User
	if err := config.DB.
		Where("role = ?", "operator").
		Select("id", "name", "email", "created_at").
		Order("created_at DESC").
		Find(&ops).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Ошибка БД"})
	}
	return c.JSON(ops)
}

func SearchUsersByEmail(c *fiber.Ctx) error {
	if !isAdmin(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа"})
	}

	q := strings.TrimSpace(c.Query("q"))
	if len(q) < 2 {
		return c.JSON([]models.User{})
	}

	var users []models.User
	if err := config.DB.
		Where("LOWER(email) LIKE LOWER(?)", "%"+q+"%").
		Where("role != ?", "operator").
		Select("id", "name", "email").
		Limit(10).Find(&users).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Ошибка БД"})
	}
	return c.JSON(users)
}

func MakeOperator(c *fiber.Ctx) error {
	if !isAdmin(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа"})
	}

	var user models.User
	if err := config.DB.First(&user, c.Params("id")).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	if user.Role == "operator" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Уже оператор"})
	}

	user.Role = "operator"
	user.UpdatedAt = time.Now()
	if err := config.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось обновить роль"})
	}
	return c.JSON(fiber.Map{"message": "Пользователь назначен оператором"})
}