// Package ws provides the WebSocket hub for real-time event broadcasting.
// Events are broadcast to rooms (keyed by schedule_id) for efficient delivery.
package ws

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/ilham/presensi-online/backend/internal/domain"
)

// EventType defines the type of WebSocket event.
type EventType string

const (
	EventSessionOpened      EventType = "SESSION_OPENED"
	EventSessionClosed      EventType = "SESSION_CLOSED"
	EventQRRefreshed        EventType = "QR_REFRESHED"
	EventAttendanceRecorded EventType = "ATTENDANCE_RECORDED"
)

// Event is the message broadcasted over WebSocket.
type Event struct {
	Type      EventType   `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"ts"`
}

// QRRefreshPayload is sent to the lecturer when the QR code rotates.
type QRRefreshPayload struct {
	SessionID  uuid.UUID `json:"session_id"`
	Token      string    `json:"token"`
	ExpiresIn  int64     `json:"expires_in_seconds"`
}

// AttendancePayload is broadcast to the lecturer and admin dashboard
// when a student successfully records attendance.
type AttendancePayload struct {
	StudentID    uuid.UUID              `json:"student_id"`
	Status       domain.AttendanceStatus `json:"status"`
	ScannedAt    time.Time              `json:"scanned_at"`
}

// Client represents a single WebSocket connection.
type Client struct {
	conn       *websocket.Conn
	scheduleID uuid.UUID
	userID     uuid.UUID
	send       chan []byte
	hub        *Hub
}

// Hub manages all WebSocket connections, grouped by schedule_id rooms.
type Hub struct {
	mu      sync.RWMutex
	rooms   map[uuid.UUID]map[*Client]struct{} // scheduleID → clients
	register   chan *Client
	unregister chan *Client
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[uuid.UUID]map[*Client]struct{}),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
	}
}

// Run starts the hub's event loop (should be called in a goroutine).
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.scheduleID] == nil {
				h.rooms[client.scheduleID] = make(map[*Client]struct{})
			}
			h.rooms[client.scheduleID][client] = struct{}{}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if room, ok := h.rooms[client.scheduleID]; ok {
				delete(room, client)
				if len(room) == 0 {
					delete(h.rooms, client.scheduleID)
				}
			}
			close(client.send)
			h.mu.Unlock()
		}
	}
}

// BroadcastToRoom sends an event to all clients in a given schedule room.
func (h *Hub) BroadcastToRoom(scheduleID uuid.UUID, event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.rooms[scheduleID] {
		select {
		case client.send <- data:
		default:
			// Client's send buffer is full — unregister it
			h.unregister <- client
		}
	}
}

// PublishAttendanceRecorded implements app.WSPublisher.
// It broadcasts an ATTENDANCE_RECORDED event to the schedule room.
func (h *Hub) PublishAttendanceRecorded(sessionID uuid.UUID, attendance *domain.Attendance) error {
	// Note: We need the scheduleID from the session — this requires the handler
	// to call BroadcastAttendance with explicit scheduleID.
	// For now, we use sessionID as a proxy (the hub can maintain a sessionID→scheduleID map).
	var scannedAt time.Time
	if attendance.ScannedAt != nil {
		scannedAt = *attendance.ScannedAt
	} else {
		scannedAt = time.Now()
	}

	event := Event{
		Type: EventAttendanceRecorded,
		Payload: AttendancePayload{
			StudentID: attendance.StudentID,
			Status:    attendance.Status,
			ScannedAt: scannedAt,
		},
		Timestamp: time.Now().Unix(),
	}
	// Broadcast to the session's room (using sessionID as room key for flexibility)
	h.BroadcastToRoom(sessionID, event)
	return nil
}

// BroadcastQRToken sends a QR_REFRESHED event to the room.
func (h *Hub) BroadcastQRToken(roomID, sessionID uuid.UUID, token string, expiresIn int64) {
	h.BroadcastToRoom(roomID, Event{
		Type: EventQRRefreshed,
		Payload: QRRefreshPayload{
			SessionID: sessionID,
			Token:     token,
			ExpiresIn: expiresIn,
		},
		Timestamp: time.Now().Unix(),
	})
}

// BroadcastSessionOpened sends a SESSION_OPENED event to the room.
func (h *Hub) BroadcastSessionOpened(roomID, sessionID uuid.UUID, meetingNo int) {
	h.BroadcastToRoom(roomID, Event{
		Type: EventSessionOpened,
		Payload: map[string]interface{}{
			"session_id": sessionID,
			"meeting_no": meetingNo,
		},
		Timestamp: time.Now().Unix(),
	})
}

// BroadcastSessionClosed sends a SESSION_CLOSED event to the room.
func (h *Hub) BroadcastSessionClosed(roomID, sessionID uuid.UUID) {
	h.BroadcastToRoom(roomID, Event{
		Type: EventSessionClosed,
		Payload: map[string]interface{}{
			"session_id": sessionID,
		},
		Timestamp: time.Now().Unix(),
	})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev/API clients
	},
}

// ServeWS upgrades the HTTP connection to a WebSocket and registers the client.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, roomID, userID uuid.UUID) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	client := NewClient(conn, roomID, userID, h)
	h.Register(client)

	go client.WritePump()
	go client.ReadPump()

	return nil
}

// Register adds a new client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// NewClient creates a new WebSocket client.
func NewClient(conn *websocket.Conn, scheduleID, userID uuid.UUID, hub *Hub) *Client {
	return &Client{
		conn:       conn,
		scheduleID: scheduleID,
		userID:     userID,
		send:       make(chan []byte, 256),
		hub:        hub,
	}
}

// WritePump pumps messages from the hub to the WebSocket connection.
func (c *Client) WritePump() {
	defer func() {
		c.conn.Close()
	}()

	for msg := range c.send {
		c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

// ReadPump reads messages from the WebSocket (handles ping/pong and cleanup).
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log error
			}
			break
		}
	}
}
