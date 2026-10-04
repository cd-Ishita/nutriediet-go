package admin

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cd-Ishita/nutriediet-go/display"
	"github.com/cd-Ishita/nutriediet-go/helpers"
	"github.com/gin-gonic/gin"
)

// Shared in-memory hub for the single-clinic waiting-room display.
var displayHub = display.NewHub()

type updateDisplayRequest struct {
	Status               string `json:"status"`
	Message              string `json:"message"`
	AnnouncementsEnabled *bool  `json:"announcementsEnabled"`
}

func GetDisplayStatus(c *gin.Context) {
	if !helpers.CheckUserType(c, "ADMIN") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized access by client"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"state":   displayHub.Current(),
	})
}

func UpdateDisplayStatus(c *gin.Context) {
	if !helpers.CheckUserType(c, "ADMIN") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized access by client"})
		return
	}

	var req updateDisplayRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := strings.TrimSpace(req.Status)
	message := strings.TrimSpace(req.Message)

	// Allow announcement-only toggles without rewriting status (avoids re-announcing).
	if status == "" {
		if req.AnnouncementsEnabled == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status or announcementsEnabled is required"})
			return
		}
		state := displayHub.SetAnnouncementsEnabled(*req.AnnouncementsEnabled)
		c.JSON(http.StatusOK, gin.H{"success": true, "state": state})
		return
	}

	switch status {
	case display.StatusPleaseWait, display.StatusNextPatient, display.StatusCustom:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	if status == display.StatusCustom && message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required for custom status"})
		return
	}

	state := displayHub.SetStatus(status, message)
	if req.AnnouncementsEnabled != nil {
		state = displayHub.SetAnnouncementsEnabled(*req.AnnouncementsEnabled)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"state":   state,
	})
}

func StreamDisplayStatus(c *gin.Context) {
	if !helpers.CheckUserType(c, "ADMIN") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized access by client"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	initial, err := displayHub.SnapshotJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode state"})
		return
	}
	if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", initial); err != nil {
		return
	}
	flusher.Flush()

	updates := displayHub.Subscribe()
	defer displayHub.Unsubscribe(updates)

	keepalive := time.NewTicker(25 * time.Second)
	defer keepalive.Stop()

	clientGone := c.Request.Context().Done()
	for {
		select {
		case <-clientGone:
			return
		case payload, open := <-updates:
			if !open {
				return
			}
			if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", payload); err != nil {
				return
			}
			flusher.Flush()
		case <-keepalive.C:
			if _, err := io.WriteString(c.Writer, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
