package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"

	"diplom/config"
	"diplom/models"
	"diplom/utils"
)


var (
	ticketRooms     = make(map[uint]map[*websocket.Conn]string)
	ticketRoomsMu   = utils.NewMutex()
	newTicketConns   = make(map[*websocket.Conn]bool)
	newTicketConnsMu = utils.NewMutex()
)

func coalesceTicket(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func saveBase64ImageTicket(dataURL string, userID uint) (*string, error) {
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("bad dataURL")
	}

	raw, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll("./public/uploads", 0755); err != nil {
		return nil, err
	}
	name := fmt.Sprintf("%d_%d.jpg", userID, time.Now().UnixNano())
	path := filepath.Join("public", "uploads", name)

	if err := os.WriteFile(path, raw, 0644); err != nil {
		return nil, err
	}
	url := "/uploads/" + name
	return &url, nil
}

type CreateTicketInput struct {
	Subject string `json:"subject"`
	Content string `json:"content"`
}

type ITicketInfo struct {
	ID            uint      `json:"id"`
	Subject       string    `json:"subject"`
	Status        string    `json:"status"`
	UserID        uint      `json:"user_id"`
	UserName      string    `json:"user_name"`
	OperatorID    *uint     `json:"operator_id"`
	OperatorName  *string   `json:"operator_name"`
	LastMessageAt time.Time `json:"last_message_at"`
}

func CreateTicket(c *fiber.Ctx) error {
	claimsAny := c.Locals("user")
	if claimsAny == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	claims := claimsAny.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var input CreateTicketInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Невозможно разобрать JSON"})
	}

	ticket := models.Ticket{
		UserID:        userID,
		Subject:       input.Subject,
		Status:        "new",
		LastMessageAt: time.Now(),
	}
	if err := config.DB.Create(&ticket).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при создании тикета"})
	}

	msg := models.TicketMessage{
		TicketID:   ticket.ID,
		SenderID:   userID,
		SenderRole: "user",
		Content:    input.Content,
		CreatedAt:  time.Now(),
	}
	if err := config.DB.Create(&msg).Error; err != nil {
		config.DB.Delete(&ticket)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Тикет создан, но не удалось сохранить сообщение"})
	}

	notifyNewTicketWS(ticket)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"ticket_id": ticket.ID})
}

func GetMyTickets(c *fiber.Ctx) error {
	claimsAny := c.Locals("user")
	if claimsAny == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	claims := claimsAny.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var tickets []models.Ticket
	if err := config.DB.
		Where("user_id = ?", userID).
		Order("last_message_at DESC").
		Find(&tickets).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось получить тикеты"})
	}

	out := make([]fiber.Map, 0, len(tickets))
	for _, t := range tickets {
		statusLabel := ""
		switch t.Status {
		case "new":
			statusLabel = "Новый"
		case "active":
			statusLabel = "В работе"
		case "closed":
			statusLabel = "Закрыт"
		}
		out = append(out, fiber.Map{
			"id":              t.ID,
			"subject":         t.Subject,
			"status":          t.Status,
			"status_label":    statusLabel,
			"last_message_at": t.LastMessageAt,
			"operator_id":     t.OperatorID,
		})
	}
	return c.JSON(out)
}

func GetTicketInfo(c *fiber.Ctx) error {
	ticketIDParam := c.Params("id")
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketIDParam).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Тикет не найден"})
	}

	claimsAny := c.Locals("user")
	if claimsAny == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	claims := claimsAny.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	role := claims["role"].(string)

	if role != "operator" && role != "admin" && ticket.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ запрещён"})
	}

	var user models.User
	if err := config.DB.First(&user, ticket.UserID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось получить информацию о пользователе"})
	}

	var operatorName *string
	if ticket.OperatorID != nil {
		var op models.User
		if err := config.DB.First(&op, *ticket.OperatorID).Error; err == nil {
			operatorName = &op.Name
		}
	}

	out := ITicketInfo{
		ID:            ticket.ID,
		Subject:       ticket.Subject,
		Status:        ticket.Status,
		UserID:        ticket.UserID,
		UserName:      user.Name,
		OperatorID:    ticket.OperatorID,
		OperatorName:  operatorName,
		LastMessageAt: ticket.LastMessageAt,
	}

	return c.JSON(out)
}

func GetTicketMessages(c *fiber.Ctx) error {
	ticketIDParam := c.Params("id")
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketIDParam).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Тикет не найден"})
	}

	claimsAny := c.Locals("user")
	if claimsAny == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	claims := claimsAny.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	role := claims["role"].(string)

	isOwner := ticket.UserID == userID
	isAdmin := role == "admin"
	if !isOwner && role != "operator" && !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ запрещён"})
	}

	var msgs []models.TicketMessage
	if err := config.DB.
		Where("ticket_id = ?", ticket.ID).
		Order("created_at ASC").
		Find(&msgs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось загрузить сообщения"})
	}
	return c.JSON(msgs)
}

func ListOperatorTickets(c *fiber.Ctx) error {
	claimsAny := c.Locals("user")
	if claimsAny == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	claims := claimsAny.(jwt.MapClaims)
	role := claims["role"].(string)
	if role != "operator" && role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ только для операторов"})
	}

	status := c.Query("status", "new")
	if status != "new" && status != "active" && status != "closed" {
		status = "new"
	}

	var tickets []models.Ticket
	if err := config.DB.
		Where("status = ?", status).
		Order("last_message_at DESC").
		Find(&tickets).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось получить тикеты"})
	}

	out := make([]fiber.Map, 0, len(tickets))
	for _, t := range tickets {
		var user models.User
		config.DB.First(&user, t.UserID)
		out = append(out, fiber.Map{
			"id":              t.ID,
			"subject":         t.Subject,
			"user_id":         t.UserID,
			"user_name":       user.Name,
			"last_message_at": t.LastMessageAt,
			"status":          t.Status,
			"operator_id":     t.OperatorID,
		})
	}
	return c.JSON(out)
}

func AssignTicket(c *fiber.Ctx) error {
	claimsAny := c.Locals("user")
	if claimsAny == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	claims := claimsAny.(jwt.MapClaims)
	role := claims["role"].(string)
	if role != "operator" && role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ только для операторов"})
	}
	operatorID := uint(claims["user_id"].(float64))

	ticketIDParam := c.Params("id")
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketIDParam).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Тикет не найден"})
	}
	if ticket.Status != "new" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Тикет уже взят или закрыт"})
	}

	ticket.OperatorID = &operatorID
	ticket.Status = "active"
	ticket.LastMessageAt = time.Now()
	if err := config.DB.Save(&ticket).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось присвоить тикет"})
	}

	notifyTicketAssignedWS(ticket)
	return c.JSON(fiber.Map{"message": "Тикет взят в работу"})
}

func CloseTicket(c *fiber.Ctx) error {
	claimsAny := c.Locals("user")
	if claimsAny == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	claims := claimsAny.(jwt.MapClaims)
	role := claims["role"].(string)
	if role != "operator" && role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ только для операторов"})
	}

	ticketIDParam := c.Params("id")
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketIDParam).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Тикет не найден"})
	}
	if ticket.Status != "active" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Тикет не в работе"})
	}

	ticket.Status = "closed"
	ticket.LastMessageAt = time.Now()
	if err := config.DB.Save(&ticket).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось закрыть тикет"})
	}

	notifyTicketClosedWS(ticket)
	return c.JSON(fiber.Map{"message": "Тикет закрыт"})
}

type CreateTicketMessageInput struct {
	Content   string  `json:"content"`
	MediaURL  *string `json:"media_url,omitempty"`
	ReplyToID *uint   `json:"reply_to_id,omitempty"`
}

func CreateTicketMessage(c *fiber.Ctx) error {
	ticketIDParam := c.Params("id")
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketIDParam).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Тикет не найден"})
	}

	claimsAny := c.Locals("user")
	if claimsAny == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	claims := claimsAny.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	role := claims["role"].(string)

	isOwner := ticket.UserID == userID
	isAssignedOperator := role == "operator" && ticket.OperatorID != nil && *ticket.OperatorID == userID
	isAdmin := role == "admin"
	if !isOwner && !isAssignedOperator && !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ запрещён"})
	}

	var input CreateTicketMessageInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверные данные сообщения"})
	}

	msg := models.TicketMessage{
		TicketID:   ticket.ID,
		SenderID:   userID,
		SenderRole: func() string { if role == "operator" || role == "admin" { return "operator" } else { return "user" } }(),
		Content:    input.Content,
		MediaURL:   input.MediaURL,
		ReplyToID:  input.ReplyToID,
		CreatedAt:  time.Now(),
	}
	if err := config.DB.Create(&msg).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось сохранить сообщение"})
	}

	ticket.LastMessageAt = time.Now()
	config.DB.Save(&ticket)

	notifyTicketMessageWS(msg)
	return c.Status(fiber.StatusCreated).JSON(msg)
}

func DeleteTicketMessageHTTP(c *fiber.Ctx) error {
	tid, _ := c.ParamsInt("id")
	mid, _ := c.ParamsInt("msgId")

	claimsAny := c.Locals("user")
	if claimsAny == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	claims := claimsAny.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	role := claims["role"].(string)

	var msg models.TicketMessage
	if err := config.DB.First(&msg, mid).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Сообщение не найдено"})
	}
	if msg.TicketID != uint(tid) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Сообщение не к этому тикету"})
	}
	if role != "operator" && role != "admin" && msg.SenderID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа"})
	}

	if err := config.DB.Delete(&msg).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось удалить сообщение"})
	}
	broadcastTicketDelete(uint(tid), msg.ID)
	return c.JSON(fiber.Map{"message": "OK"})
}

func SupportNewTicketsWS(c *websocket.Conn) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.Close()
		return
	}
	secret := []byte(os.Getenv("JWT_SECRET"))
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !tok.Valid {
		c.Close()
		return
	}
	claims := tok.Claims.(jwt.MapClaims)
	if claims["role"].(string) != "operator" && claims["role"].(string) != "admin" {
		c.Close()
		return
	}

	newTicketConnsMu.Lock()
	newTicketConns[c] = true
	newTicketConnsMu.Unlock()

	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}

	newTicketConnsMu.Lock()
	delete(newTicketConns, c)
	newTicketConnsMu.Unlock()
	c.Close()
}

func SupportTicketChatWS(c *websocket.Conn) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		log.Println("WS тикета: токен отсутствует — закрываем соединение")
		c.Close()
		return
	}
	secret := []byte(os.Getenv("JWT_SECRET"))
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !tok.Valid {
		log.Printf("WS тикета: некорректный или просроченный токен: %v\n", err)
		c.Close()
		return
	}
	claims := tok.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	role := claims["role"].(string)
	log.Printf("WS тикета: токен валиден, userID=%d, role=%s\n", userID, role)

	ticketIDParam := c.Params("id")
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketIDParam).Error; err != nil {
		log.Println("WS тикета: тикет не найден, ID=", ticketIDParam)
		c.Close()
		return
	}

	isOwner := ticket.UserID == userID
	isOperator := role == "operator"
	isAdmin := role == "admin"
	if !isOwner && !isOperator && !isAdmin {
		log.Printf("WS тикета: пользователь %d (role=%s) не имеет прав на этот тикет ID=%d\n", userID, role, ticket.ID)
		c.Close()
		return
	}
	log.Printf("WS тикета: права доступа подтверждены для userID=%d, ticketID=%d\n", userID, ticket.ID)

	ticketRoomsMu.Lock()
	if ticketRooms[ticket.ID] == nil {
		ticketRooms[ticket.ID] = make(map[*websocket.Conn]string)
	}
	ticketRooms[ticket.ID][c] = role
	ticketRoomsMu.Unlock()
	log.Printf("WS тикета: соединение зарегистрировано (ticketID=%d, conn=%v)\n", ticket.ID, c.RemoteAddr())

	for {
		_, raw, err := c.ReadMessage()
		if err != nil {
			log.Printf("WS тикета: ReadMessage вернул ошибку: %v\n", err)
			break
		}

		var envelope struct {
			Content  *string `json:"content"`
			ReplyTo  *uint   `json:"reply_to"`
			MediaB64 *string `json:"media"`
			DeleteID *uint   `json:"delete_id"`
		}
		if err := json.Unmarshal(raw, &envelope); err != nil {
			log.Println("WS тикета: не удалось распарсить JSON:", err)
			continue
		}
		log.Printf("WS тикета: получен пакет: %+v\n", envelope)

		if envelope.DeleteID != nil {
			msgID := *envelope.DeleteID
			var msg models.TicketMessage
			if err := config.DB.First(&msg, msgID).Error; err == nil {
				if msg.TicketID == ticket.ID && msg.SenderID == userID {
					config.DB.Delete(&msg)
					broadcastTicketDelete(ticket.ID, msg.ID)
					log.Printf("WS тикета: сообщение ID=%d удалено пользователем %d\n", msgID, userID)
				}
			}
			continue
		}

		if envelope.Content == nil && envelope.MediaB64 == nil {
			continue
		}
		var mediaURL *string
		if envelope.MediaB64 != nil {
			base64Data := *envelope.MediaB64
			parts := strings.SplitN(base64Data, ",", 2)
			if len(parts) == 2 {
				rawImg, err := base64.StdEncoding.DecodeString(parts[1])
				if err == nil {
					if err := os.MkdirAll("./public/uploads", 0755); err == nil {
						fileName := fmt.Sprintf("%d_%d.jpg", userID, time.Now().UnixNano())
						savePath := filepath.Join("public", "uploads", fileName)
						if err := os.WriteFile(savePath, rawImg, 0644); err == nil {
							url := "/uploads/" + fileName
							mediaURL = &url
						} else {
							log.Println("WS тикета: ошибка записи файла:", err)
						}
					} else {
						log.Println("WS тикета: не удалось создать папку ./public/uploads:", err)
					}
				} else {
					log.Println("WS тикета: не удалось декодировать Base64 media:", err)
				}
			}
		}
		newMsg := models.TicketMessage{
			TicketID:   ticket.ID,
			SenderID:   userID,
			SenderRole: role,
			Content:    "",
			MediaURL:   mediaURL,
			ReplyToID:  envelope.ReplyTo,
			CreatedAt:  time.Now(),
		}
		if envelope.Content != nil {
			newMsg.Content = *envelope.Content
		}
		if err := config.DB.Create(&newMsg).Error; err != nil {
			log.Println("WS тикета: ошибка при создании TicketMessage:", err)
			continue
		}
		log.Printf("WS тикета: создано сообщение ID=%d\n", newMsg.ID)
		ticket.LastMessageAt = time.Now()
		if err := config.DB.Save(&ticket).Error; err != nil {
			log.Println("WS тикета: не удалось обновить LastMessageAt:", err)
		}

		notifyTicketMessageWS(newMsg)
		log.Println("WS тикета: разослали событие support:ticket_message")
	}

	ticketRoomsMu.Lock()
	delete(ticketRooms[ticket.ID], c)
	if len(ticketRooms[ticket.ID]) == 0 {
		delete(ticketRooms, ticket.ID)
	}
	ticketRoomsMu.Unlock()

	c.Close()
}

func notifyNewTicketWS(ticket models.Ticket) {
	payload, _ := json.Marshal(fiber.Map{
		"event": "support:new_ticket",
		"data": fiber.Map{
			"ticket_id":       ticket.ID,
			"subject":         ticket.Subject,
			"user_id":         ticket.UserID,
			"last_message_at": ticket.LastMessageAt,
		},
	})
	newTicketConnsMu.Lock()
	for conn := range newTicketConns {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("WS new_ticket error: %v\n", err)
		}
	}
	newTicketConnsMu.Unlock()
}

func notifyTicketAssignedWS(ticket models.Ticket) {
	payload, _ := json.Marshal(fiber.Map{
		"event": "support:ticket_assigned",
		"data": fiber.Map{
			"ticket_id":   ticket.ID,
			"operator_id": ticket.OperatorID,
		},
	})
	newTicketConnsMu.Lock()
	for conn := range newTicketConns {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("WS ticket_assigned error: %v\n", err)
		}
	}
	newTicketConnsMu.Unlock()
}

func notifyTicketClosedWS(ticket models.Ticket) {
	payload, _ := json.Marshal(fiber.Map{
		"event": "support:ticket_closed",
		"data": fiber.Map{
			"ticket_id": ticket.ID,
		},
	})
	tID := ticket.ID
	ticketRoomsMu.Lock()
	for conn := range ticketRooms[tID] {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("WS ticket_closed error: %v\n", err)
		}
	}
	ticketRoomsMu.Unlock()
}

func notifyTicketMessageWS(msg models.TicketMessage) {
    payload, _ := json.Marshal(fiber.Map{
        "event": "support:ticket_message",
        "data":  msg,
    })
    tID := msg.TicketID

    ticketRoomsMu.Lock()
    for conn := range ticketRooms[tID] {
        if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
            log.Printf("WS ticket_message error: %v\n", err)
        }
    }
    ticketRoomsMu.Unlock()
}



func broadcastTicketDelete(ticketID, messageID uint) {
	payload, _ := json.Marshal(fiber.Map{
		"event": "support:ticket_delete",
		"data": fiber.Map{
			"message_id": messageID,
		},
	})
	ticketRoomsMu.Lock()
	for conn := range ticketRooms[ticketID] {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("WS ticket_delete error: %v\n", err)
		}
	}
	ticketRoomsMu.Unlock()
}
