package controllers

import (
	"os"
	"time"

	"diplom/config"
	"diplom/mail"
	"diplom/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CreateFamilyInput struct {
	Name string `json:"name"`
}

func CreateFamily(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка извлечения данных из токена",
		})
	}
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка извлечения user_id из токена",
		})
	}
	userID := uint(userIDFloat)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Пользователь не найден",
		})
	}

	if user.FamilyID != 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Вы уже состоите в семье",
		})
	}

	var input CreateFamilyInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Невозможно разобрать JSON",
		})
	}

	family := models.Family{
		Name:      input.Name,
		OwnerID:   user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := config.DB.Create(&family).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка создания семьи",
		})
	}

	calendar := models.Calendar{
		FamilyID:  family.ID,
		Title:     "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := config.DB.Create(&calendar).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка создания календаря для семьи",
		})
	}

	user.FamilyID = family.ID
	config.DB.Save(&user)

	return c.JSON(fiber.Map{
		"family": family,
	})
}

type InviteInput struct {
	Email string `json:"email"`
}

func InviteMember(c *fiber.Ctx) error {
	var input InviteInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Невозможно разобрать JSON",
		})
	}

	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка извлечения данных из токена",
		})
	}
	senderIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка извлечения user_id из токена",
		})
	}
	senderID := uint(senderIDFloat)

	var inviter models.User
	if err := config.DB.First(&inviter, senderID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка извлечения отправителя",
		})
	}
	if inviter.FamilyID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Вы не состоите в семье, не можете приглашать",
		})
	}

	var invitee models.User
	if err := config.DB.Where("email = ?", input.Email).First(&invitee).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Пользователь с таким email не найден",
		})
	}
	if invitee.FamilyID != 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Пользователь уже состоит в семье",
		})
	}

	inviteToken := uuid.New().String()
	invitation := models.FamilyInvitation{
		FamilyID:  inviter.FamilyID,
		Email:     invitee.Email,
		Token:     inviteToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := config.DB.Create(&invitation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка создания приглашения",
		})
	}
	inviteLink := os.Getenv("CLIENT_URL") + "/dashboard/family/invite/" + inviteToken
	mailService := mail.NewMailService()
	if err := mailService.SendFamilyInviteMail(invitee.Email, inviteLink); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка отправки приглашения",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Приглашение отправлено",
	})
}

func AcceptInvitation(c *fiber.Ctx) error {
	token := c.Params("token")
	var invitation models.FamilyInvitation
	if err := config.DB.Where("token = ?", token).First(&invitation).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный токен приглашения",
		})
	}
	var user models.User
	if err := config.DB.Where("email = ?", invitation.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Пользователь не найден",
		})
	}

	user.FamilyID = invitation.FamilyID
	config.DB.Save(&user)
	config.DB.Delete(&invitation)

	return c.JSON(fiber.Map{"message": "Приглашение принято. Вы вступили в семью."})
}

func GetFamilyDetails(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	if user.FamilyID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Вы не состоите в семье"})
	}

	var family models.Family
	if err := config.DB.First(&family, user.FamilyID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Семья не найдена"})
	}
	var members []models.User
	if err := config.DB.Where("family_id = ?", user.FamilyID).Find(&members).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка загрузки членов семьи"})
	}

	return c.JSON(fiber.Map{
		"family":  family,
		"members": members,
	})
}
