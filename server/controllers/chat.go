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
	roomsChat   = make(map[uint]map[*websocket.Conn]uint)
	roomsChatMu = utils.NewMutex()
)

func coalesceChat(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func saveBase64ImageChat(dataURL string, userID uint) (*string, error) {
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

func ChatHistory(c *fiber.Ctx) error {
	claims := c.Locals("user").(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var u models.User
	if err := config.DB.First(&u, userID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	var msgs []models.ChatMessage
	if err := config.DB.
		Where("family_id = ?", u.FamilyID).
		Order("created_at").
		Find(&msgs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot load history"})
	}
	return c.JSON(msgs)
}

func ChatWebSocket(c *websocket.Conn) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.Close()
		return
	}
	secretKey := []byte(os.Getenv("JWT_SECRET"))
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil || !tok.Valid {
		c.Close()
		return
	}
	claims := tok.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.Close()
		return
	}
	familyID := user.FamilyID
	if familyID == 0 {
		c.Close()
		return
	}
	roomsChatMu.Lock()
	if roomsChat[familyID] == nil {
		roomsChat[familyID] = make(map[*websocket.Conn]uint)
	}
	roomsChat[familyID][c] = userID
	roomsChatMu.Unlock()
	broadcastChatPresence(familyID)

	for {
		_, raw, err := c.ReadMessage()
		if err != nil {
			break
		}

		var envelope struct {
			Content  *string `json:"content"`
			ReplyTo  *uint   `json:"reply_to"`
			MediaB64 *string `json:"media"`
			DeleteID *uint   `json:"delete_id"`
		}
		if err := json.Unmarshal(raw, &envelope); err != nil {
			continue
		}

		if envelope.DeleteID != nil {
			var m models.ChatMessage
			if err := config.DB.First(&m, *envelope.DeleteID).Error; err == nil && m.UserID == userID {
				config.DB.Delete(&m)
				broadcastChatDelete(familyID, m.ID)
			}
			continue
		}

		if envelope.Content == nil && envelope.MediaB64 == nil {
			continue
		}

		var mediaURL *string
		if envelope.MediaB64 != nil {
			if url, err := saveBase64ImageChat(*envelope.MediaB64, userID); err == nil {
				mediaURL = url
			} else {
				log.Println("save image chat error:", err)
			}
		}

		msg := models.ChatMessage{
			FamilyID:  familyID,
			UserID:    userID,
			Content:   coalesceChat(envelope.Content),
			MediaURL:  mediaURL,
			ReplyToID: envelope.ReplyTo,
			CreatedAt: time.Now(),
		}
		if err := config.DB.Create(&msg).Error; err != nil {
			log.Println("db create chat error:", err)
			continue
		}
		broadcastChatMessage(familyID, msg)
	}

	roomsChatMu.Lock()
	delete(roomsChat[familyID], c)
	roomsChatMu.Unlock()
	broadcastChatPresence(familyID)
}

func broadcastChatMessage(fam uint, msg models.ChatMessage) {
	roomsChatMu.Lock()
	defer roomsChatMu.Unlock()

	payload, _ := json.Marshal(struct {
		Type string             `json:"type"`
		Data models.ChatMessage `json:"data"`
	}{"message", msg})

	for conn := range roomsChat[fam] {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil && !websocket.IsCloseError(err) {
			log.Printf("WS chat write error: %v\n", err)
		}
	}
}

func broadcastChatDelete(fam uint, id uint) {
	roomsChatMu.Lock()
	defer roomsChatMu.Unlock()

	payload, _ := json.Marshal(struct {
		Type string `json:"type"`
		Data uint   `json:"data"`
	}{"delete", id})

	for conn := range roomsChat[fam] {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil && !websocket.IsCloseError(err) {
			log.Printf("WS chat delete error: %v\n", err)
		}
	}
}

func broadcastChatPresence(fam uint) {
	roomsChatMu.Lock()
	defer roomsChatMu.Unlock()

	online := make([]uint, 0, len(roomsChat[fam]))
	for _, uid := range roomsChat[fam] {
		online = append(online, uid)
	}

	payload, _ := json.Marshal(struct {
		Type string `json:"type"`
		Data []uint `json:"data"`
	}{"presence", online})

	for conn := range roomsChat[fam] {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil && !websocket.IsCloseError(err) {
			log.Printf("WS chat presence error: %v\n", err)
		}
	}
}
