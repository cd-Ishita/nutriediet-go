package display

import (
	"encoding/json"
	"sync"
	"time"
)

const (
	StatusPleaseWait  = "please_wait"
	StatusNextPatient = "next_patient"
	StatusCustom      = "custom"
)

// State is the waiting-room display payload pushed over SSE.
// Extra fields can be added later (token number, room id, locale, etc.).
type State struct {
	Status               string    `json:"status"`
	Title                string    `json:"title"`
	Message              string    `json:"message"`
	SpokenText           string    `json:"spokenText"`
	AnnouncementsEnabled bool      `json:"announcementsEnabled"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type Hub struct {
	mu      sync.RWMutex
	state   State
	clients map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{
		state:   defaultState(),
		clients: make(map[chan []byte]struct{}),
	}
}

func defaultState() State {
	return BuildState(StatusPleaseWait, "", true)
}

func BuildState(status, customMessage string, announcementsEnabled bool) State {
	title, message, spoken := resolveCopy(status, customMessage)
	return State{
		Status:               status,
		Title:                title,
		Message:              message,
		SpokenText:           spoken,
		AnnouncementsEnabled: announcementsEnabled,
		UpdatedAt:            time.Now().UTC(),
	}
}

func resolveCopy(status, customMessage string) (title, message, spoken string) {
	switch status {
	case StatusNextPatient:
		return "NEXT PATIENT", "Please come in.", "Next patient, please come in."
	case StatusCustom:
		msg := customMessage
		if msg == "" {
			msg = "Please wait."
		}
		return "NOTICE", msg, msg
	default:
		return "PLEASE WAIT", "Another consultation is currently in progress.", "Please wait. Another consultation is currently in progress."
	}
}

func (h *Hub) Current() State {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.state
}

func (h *Hub) SetStatus(status, customMessage string) State {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.state = BuildState(status, customMessage, h.state.AnnouncementsEnabled)
	h.broadcastLocked()
	return h.state
}

func (h *Hub) SetAnnouncementsEnabled(enabled bool) State {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.state.AnnouncementsEnabled = enabled
	// Do not bump UpdatedAt — display clients use it to avoid re-speaking.
	h.broadcastLocked()
	return h.state
}

func (h *Hub) Subscribe() chan []byte {
	ch := make(chan []byte, 4)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(ch chan []byte) {
	h.mu.Lock()
	if _, ok := h.clients[ch]; ok {
		delete(h.clients, ch)
		close(ch)
	}
	h.mu.Unlock()
}

func (h *Hub) SnapshotJSON() ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return json.Marshal(h.state)
}

func (h *Hub) broadcastLocked() {
	payload, err := json.Marshal(h.state)
	if err != nil {
		return
	}
	for ch := range h.clients {
		select {
		case ch <- payload:
		default:
			// Slow client — drop this update; next one or reconnect will catch up.
		}
	}
}
