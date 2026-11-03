package api

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"schedule-optimizer/internal/cache"
	"schedule-optimizer/internal/generator"
	"schedule-optimizer/internal/search"
	"schedule-optimizer/internal/stats/grades"
	"schedule-optimizer/internal/store"

	"github.com/gin-gonic/gin"
)

const maxTermsReturned = 8 // ~2 years of quarters

// Handlers contains all HTTP handler dependencies.
type Handlers struct {
	db        *sql.DB
	cache     *cache.ScheduleCache
	generator *generator.Service
	queries   *store.Queries
	search    *search.Service
	grades    *grades.Service
}

// Helper functions for converting sql.Null types

func fromNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func fromNullInt64(ni sql.NullInt64) int {
	if ni.Valid {
		return int(ni.Int64)
	}
	return 0
}

func fromNullInt64Raw(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}

func fromNullInt64ToBool(ni sql.NullInt64) bool {
	return ni.Valid && ni.Int64 != 0
}

// Analytics helper: matches UUID v4 format (with or without dashes) or 32-char hex
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$`)

func getSessionID(c *gin.Context) sql.NullString {
	id := c.GetHeader("X-Session-ID")
	if id == "" || !uuidPattern.MatchString(strings.ToLower(id)) {
		return sql.NullString{}
	}
	return sql.NullString{String: id, Valid: true}
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullInt64(n int) sql.NullInt64 {
	if n == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(n), Valid: true}
}

func nullIntPtr(p *int) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*p), Valid: true}
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// NewHandlers creates a new Handlers instance with all dependencies.
func NewHandlers(db *sql.DB, cache *cache.ScheduleCache, generator *generator.Service, queries *store.Queries, searchSvc *search.Service, gradesSvc *grades.Service) *Handlers {
	return &Handlers{
		db:        db,
		cache:     cache,
		generator: generator,
		queries:   queries,
		search:    searchSvc,
		grades:    gradesSvc,
	}
}

// validateTerm checks if a term exists and sends an error response if not.
// Returns true if the term is valid, false if an error response was sent.
func (h *Handlers) validateTerm(c *gin.Context, term string) bool {
	_, err := h.queries.GetTermByCode(c.Request.Context(), term)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Term not found: " + term})
		return false
	}
	if err != nil {
		slog.Error("Failed to validate term", "term", term, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate term"})
		return false
	}
	return true
}

// Health returns server health status.
func (h *Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
