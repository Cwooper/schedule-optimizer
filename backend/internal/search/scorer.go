package search

import (
	"math"
	"strings"
	"time"

	"schedule-optimizer/internal/jobs"
)

// Scorer calculates relevance scores for search results.
type Scorer interface {
	// Score returns a relevance score for the given section and search request.
	Score(section *sectionRow, req *SearchRequest) float64
	// Name returns the scorer's name for debugging/logging.
	Name() string
}

// RecencyScorer scores sections based on term recency.
// More recent terms score higher using exponential decay.
// Only applies when searching across multiple terms (year scope or all-time).
type RecencyScorer struct {
	decayRate float64 // How quickly score decays per term (default 0.3)
	maxScore  float64 // Maximum score for current term
}

// NewRecencyScorer creates a new recency scorer with default settings.
func NewRecencyScorer() *RecencyScorer {
	return &RecencyScorer{
		decayRate: 0.3,
		maxScore:  50.0,
	}
}

// Score implements Scorer.
// Returns 0 for single-term searches since recency is meaningless when all results are from the same term.
func (s *RecencyScorer) Score(section *sectionRow, req *SearchRequest) float64 {
	// Skip recency scoring for single-term searches
	if req.Term != "" {
		return 0
	}
	currentTerm := jobs.CurrentTermCode(time.Now())
	distance := termDistance(section.Term, currentTerm)
	return s.maxScore * math.Exp(-float64(distance)*s.decayRate)
}

// Name implements Scorer.
func (s *RecencyScorer) Name() string {
	return "recency"
}

// MatchQualityScorer scores sections based on how well they match the search query.
// Exact matches score higher than partial matches.
type MatchQualityScorer struct {
	exactMatchBonus  float64
	prefixMatchBonus float64
}

// NewMatchQualityScorer creates a new match quality scorer with default settings.
func NewMatchQualityScorer() *MatchQualityScorer {
	return &MatchQualityScorer{
		exactMatchBonus:  30.0,
		prefixMatchBonus: 15.0,
	}
}

// Score implements Scorer.
func (s *MatchQualityScorer) Score(section *sectionRow, req *SearchRequest) float64 {
	var score float64

	// Course number match quality
	if req.CourseNumber != "" {
		query := strings.ToUpper(strings.TrimSuffix(strings.TrimSuffix(req.CourseNumber, "*"), "%"))
		if section.CourseNumber == query {
			score += s.exactMatchBonus
		} else if strings.HasPrefix(section.CourseNumber, query) {
			score += s.prefixMatchBonus
		}
	}

	// Subject match quality (if specified)
	if req.Subject != "" && strings.EqualFold(section.Subject, req.Subject) {
		score += s.exactMatchBonus
	}

	// Title match quality - boost for exact word matches
	if req.Title != "" {
		titleLower := strings.ToLower(section.Title)
		queryLower := strings.ToLower(req.Title)
		tokens := strings.FieldsFunc(queryLower, func(r rune) bool {
			return r == ' ' || r == '-'
		})
		for _, token := range tokens {
			// Check for word boundary match vs substring match
			if strings.Contains(" "+titleLower+" ", " "+token+" ") {
				score += s.prefixMatchBonus // Full word match
			} else if strings.Contains(titleLower, token) {
				score += s.prefixMatchBonus / 2 // Partial match
			}
		}
	}

	return score
}

// Name implements Scorer.
func (s *MatchQualityScorer) Name() string {
	return "match_quality"
}

// termDistance calculates the number of terms between two term codes.
func termDistance(term1, term2 string) int {
	y1, q1, err1 := jobs.ParseTermCode(term1)
	y2, q2, err2 := jobs.ParseTermCode(term2)

	if err1 != nil || err2 != nil {
		return 100 // Large distance for invalid terms
	}

	// Convert quarter to index (Winter=0, Spring=1, Summer=2, Fall=3)
	quarterIndex := func(q int) int {
		switch q {
		case jobs.QuarterWinter:
			return 0
		case jobs.QuarterSpring:
			return 1
		case jobs.QuarterSummer:
			return 2
		case jobs.QuarterFall:
			return 3
		}
		return 0
	}

	idx1 := y1*4 + quarterIndex(q1)
	idx2 := y2*4 + quarterIndex(q2)

	diff := idx2 - idx1
	if diff < 0 {
		diff = -diff
	}
	return diff
}
