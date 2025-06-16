package controllers

import (
	"github.com/gofiber/websocket/v2"
	"sync"
)

var TicketRooms   = make(map[uint]map[*websocket.Conn]uint)
var TicketRoomsMu sync.Mutex
