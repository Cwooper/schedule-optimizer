package api

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"schedule-optimizer/internal/store"
	"strings"

	"github.com/gin-gonic/gin"
)

// AnnouncementResponse wraps the announcement for JSON serialization.
type AnnouncementResponse struct {
	Announcement *AnnouncementInfo `json:"announcement"`
}

// AnnouncementInfo is the public shape of an announcement.
type AnnouncementInfo struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Type  string `json:"type"`
}

// GetAnnouncement returns the current active announcement, if any.
func (h *Handlers) GetAnnouncement(c *gin.Context) {
	row, err := h.queries.GetActiveAnnouncement(c.Request.Context())
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusOK, AnnouncementResponse{Announcement: nil})
		return
	}
	if err != nil {
		slog.Error("Failed to fetch announcement", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch announcement"})
		return
	}

	c.JSON(http.StatusOK, AnnouncementResponse{
		Announcement: &AnnouncementInfo{
			ID:    row.ID,
			Title: row.Title,
			Body:  row.Body,
			Type:  row.Type,
		},
	})
}

const maxFeedbackLength = 1000

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

// FeedbackRequest is the request body for submitting feedback.
type FeedbackRequest struct {
	Message string `json:"message" binding:"required"`
}

// SubmitFeedback stores user feedback in the database.
func (h *Handlers) SubmitFeedback(c *gin.Context) {
	var req FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	// Strip HTML tags, then validate
	msg := htmlTagPattern.ReplaceAllString(req.Message, "")
	msg = strings.TrimSpace(msg)

	if msg == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message cannot be empty"})
		return
	}
	if len(msg) > maxFeedbackLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message exceeds 1000 character limit"})
		return
	}

	sessionID := getSessionID(c)
	go func() {
		if err := h.queries.InsertFeedback(context.Background(), store.InsertFeedbackParams{
			SessionID: sessionID,
			Message:   msg,
		}); err != nil {
			slog.Warn("Failed to store feedback", "error", err)
		}
	}()

	c.Status(http.StatusNoContent)
}
