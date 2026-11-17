package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"schedule-optimizer/internal/store"
	"schedule-optimizer/internal/termcode"

	"github.com/gin-gonic/gin"
)

// A valid event is under 200 bytes; the cap keeps a forged one from making
// the server buffer more.
const maxEventBodyBytes = 1024

// icsOptions are the .ics export options in the order the client joins them
// into an ics_export target, so each option set has one spelling to group by.
var icsOptions = []string{"classes", "finals", "holidays", "reg1", "reg2", "reg3", "deadlines"}

// eventTargets is the event vocabulary, mirrored in frontend/src/lib/track.ts.
// Each name maps to its target check, so a forged beacon can't store
// arbitrary strings.
var eventTargets = map[string]func(string) bool{
	"visit":         oneOf("mobile", "desktop"),
	"tab":           oneOf("schedule", "search", "statistics"),
	"ics_export":    validICSTarget,
	"png_export":    oneOf("schedule", "finals"),
	"finals_view":   noTarget,
	"feedback_open": noTarget,
}

func oneOf(targets ...string) func(string) bool {
	return func(t string) bool { return slices.Contains(targets, t) }
}

func noTarget(t string) bool { return t == "" }

// validICSTarget accepts a non-empty comma-joined subset of icsOptions in
// their listed order.
func validICSTarget(t string) bool {
	last := -1
	for opt := range strings.SplitSeq(t, ",") {
		i := slices.Index(icsOptions, opt)
		if i <= last {
			return false
		}
		last = i
	}
	return true
}

// EventRequest is a client-side event. The session ID is in the body rather
// than the X-Session-ID header because navigator.sendBeacon can't set headers.
type EventRequest struct {
	SessionID string `json:"session_id"`
	Name      string `json:"name"`
	Target    string `json:"target"`
	Term      string `json:"term"`
}

func (r EventRequest) validate() error {
	if !parseSessionID(r.SessionID).Valid {
		return errors.New("invalid session_id")
	}
	validTarget, ok := eventTargets[r.Name]
	if !ok {
		return errors.New("unknown event name")
	}
	if !validTarget(r.Target) {
		return errors.New("invalid target for " + r.Name)
	}
	if r.Term != "" {
		if _, _, err := termcode.Parse(r.Term); err != nil {
			return errors.New("invalid term")
		}
	}
	return nil
}

// RecordEvent stores a client-side event.
func (h *Handlers) RecordEvent(c *gin.Context) {
	// ReadAll and Unmarshal rather than ShouldBindJSON, which decodes only the
	// first JSON value and ignores whatever follows it, past the cap included.
	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, maxEventBodyBytes))
	var req EventRequest
	if err == nil {
		err = json.Unmarshal(body, &req)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event body"})
		return
	}
	if err := req.validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	go func() {
		if err := h.queries.InsertEvent(context.Background(), store.InsertEventParams{
			SessionID: req.SessionID,
			Name:      req.Name,
			Target:    nullStr(req.Target),
			Term:      nullStr(req.Term),
		}); err != nil {
			slog.Warn("Failed to store event", "error", err)
		}
	}()

	c.Status(http.StatusNoContent)
}
