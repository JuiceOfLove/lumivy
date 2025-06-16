package controllers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"diplom/config"
	"diplom/models"
)

type CreateEventInput struct {
	CalendarID  uint   `json:"calendar_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Color       string `json:"color"`
	Private     bool   `json:"private"`
}

type UpdateEventInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Color       string `json:"color"`
	Private     bool   `json:"private"`
}

type MonthQuery struct {
	Month int `query:"month"`
	Year  int `query:"year"`
}

func CreateEvent(c *fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "У пользователя нет семьи"})
	}

	var input CreateEventInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Ошибка парсинга JSON"})
	}

	var cal models.Calendar
	if err := config.DB.First(&cal, input.CalendarID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Календарь не найден"})
	}
	if cal.FamilyID != user.FamilyID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа к календарю"})
	}

	start, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Некорректный формат start_time"})
	}
	end, err := time.Parse(time.RFC3339, input.EndTime)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Некорректный формат end_time"})
	}

	event := models.Event{
		CalendarID:  input.CalendarID,
		FamilyID:    user.FamilyID,
		Title:       input.Title,
		Description: input.Description,
		StartTime:   start,
		EndTime:     end,
		CreatedBy:   userID,
		IsCompleted: false,
		Private:     input.Private,
	}
	if input.Color != "" {
		event.Color = &input.Color
	}

	if err := config.DB.Create(&event).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка сохранения события"})
	}

	return c.JSON(fiber.Map{"event": event})
}

func UpdateEvent(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	eventID := c.Params("id")
	var event models.Event
	if err := config.DB.First(&event, eventID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Событие не найдено"})
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	if user.FamilyID != event.FamilyID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа к событию"})
	}

	var input UpdateEventInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Ошибка парсинга JSON"})
	}

	start, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Некорректный формат start_time"})
	}
	end, err := time.Parse(time.RFC3339, input.EndTime)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Некорректный формат end_time"})
	}

	event.Title = input.Title
	event.Description = input.Description
	event.StartTime = start
	event.EndTime = end
	event.Private = input.Private
	if input.Color == "" {
		event.Color = nil
	} else {
		event.Color = &input.Color
	}

	if err := config.DB.Save(&event).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка обновления события"})
	}

	return c.JSON(fiber.Map{"event": event})
}

func GetAllEvents(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
	}

	var events []models.Event
	if err := config.DB.
		Where("family_id = ? AND (private = ? OR created_by = ?)", user.FamilyID, false, userID).
		Order("start_time ASC").
		Find(&events).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка загрузки событий"})
	}

	return c.JSON(events)
}

func GetEventsForMonth(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
	}

	month := c.QueryInt("month", 0)
	year := c.QueryInt("year", 0)
	if month < 1 || month > 12 || year < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Нужны валидные month и year"})
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	var events []models.Event
	if err := config.DB.
		Where("family_id = ? AND (private = ? OR created_by = ?) AND start_time >= ? AND start_time < ?",
			user.FamilyID, false, userID, startDate, endDate).
		Order("start_time ASC").
		Find(&events).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при запросе"})
	}

	return c.JSON(events)
}

func CompleteEvent(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	eventID := c.Params("id")
	var event models.Event
	if err := config.DB.First(&event, eventID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Событие не найдено"})
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	if user.FamilyID != event.FamilyID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа к событию"})
	}

	event.IsCompleted = true
	if err := config.DB.Save(&event).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка обновления события"})
	}

	return c.JSON(fiber.Map{"message": "Событие выполнено", "event": event})
}

type CreateExtraCalendarInput struct {
	Title string `json:"title"`
}

func CreateExtraCalendar(c *fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "У пользователя нет семьи"})
	}

	var input CreateExtraCalendarInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Ошибка парсинга JSON"})
	}

	cal := models.Calendar{
		FamilyID: user.FamilyID,
		Title:    input.Title,
	}
	if err := config.DB.Create(&cal).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка создания календаря"})
	}

	return c.JSON(fiber.Map{"calendar": cal})
}

func GetCalendarsList(c *fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "У пользователя нет семьи"})
	}

	var cals []models.Calendar
	if err := config.DB.Where("family_id = ?", user.FamilyID).Find(&cals).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка загрузки календарей"})
	}

	return c.JSON(cals)
}

func GetEventsForCalendar(c *fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "У пользователя нет семьи"})
	}

	calendarID := c.Params("calendar_id")
	var cal models.Calendar
	if err := config.DB.First(&cal, calendarID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Календарь не найден"})
	}
	if cal.FamilyID != user.FamilyID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа к календарю"})
	}

	month := c.QueryInt("month", 0)
	year := c.QueryInt("year", 0)
	if month < 1 || month > 12 || year < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Нужны валидные month и year"})
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	var events []models.Event
	if err := config.DB.
		Where("calendar_id = ? AND (private = ? OR created_by = ?) AND start_time >= ? AND start_time < ?",
			cal.ID, false, userID, startDate, endDate).
		Order("start_time ASC").
		Find(&events).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при запросе"})
	}

	return c.JSON(events)
}

func GetEventsForToday(c *fiber.Ctx) error {
    claims, _ := c.Locals("user").(jwt.MapClaims)
    userID := uint(claims["user_id"].(float64))

    var user models.User
    if err := config.DB.First(&user, userID).Error; err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
    }

    now := time.Now().UTC()
    start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
    end   := start.AddDate(0, 0, 1)

    var events []models.Event
    if err := config.DB.
        Where("family_id = ? AND (private = ? OR created_by = ?) AND start_time < ? AND end_time >= ?",
            user.FamilyID, false, userID, end, start).
        Order("start_time ASC").
        Find(&events).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при запросе"})
    }

    return c.JSON(events)
}

func GetNextEvent(c *fiber.Ctx) error {
	claims, _ := c.Locals("user").(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
	}

	now := time.Now().UTC()
	var event models.Event
	if err := config.DB.
		Where("family_id = ? AND (private = ? OR created_by = ?) AND start_time > ?",
			user.FamilyID, false, userID, now).
		Order("start_time ASC").
		First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(fiber.Map{"message": "Нет предстоящих событий"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при запросе"})
	}

	return c.JSON(event)
}
