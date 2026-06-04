package static

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	handler := Handler()

	tests := []struct {
		name           string
		path           string
		acceptGzip     bool
		wantStatus     int
		wantTypePrefix string
		wantGzip       bool
	}{
		{"root", "/", false, http.StatusOK, "text/html", false},
		{"root gzip", "/", true, http.StatusOK, "text/html", true},
		{"index explicit redirects", "/index.html", false, http.StatusMovedPermanently, "", false},
		{"svg logo", "/schopt-logo-dark.svg", false, http.StatusOK, "image/svg+xml", false},

		// Unknown paths should 404
		{"unknown path", "/nonexistent.txt", false, http.StatusNotFound, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.acceptGzip {
				req.Header.Set("Accept-Encoding", "gzip")
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantTypePrefix != "" {
				ct := w.Header().Get("Content-Type")
				if !strings.HasPrefix(ct, tt.wantTypePrefix) {
					t.Errorf("Content-Type = %q, want prefix %q", ct, tt.wantTypePrefix)
				}
			}

			gotGzip := w.Header().Get("Content-Encoding") == "gzip"
			if gotGzip != tt.wantGzip {
				t.Errorf("gzip = %v, want %v", gotGzip, tt.wantGzip)
			}
		})
	}
}

func TestHandlerVaryHeader(t *testing.T) {
	handler := Handler()

	// Gzip responses should include Vary header for caching
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if vary := w.Header().Get("Vary"); vary != "Accept-Encoding" {
		t.Errorf("Vary = %q, want 'Accept-Encoding'", vary)
	}
}

func TestAcceptsGzipEdgeCases(t *testing.T) {
	handler := Handler()

	tests := []struct {
		name           string
		acceptEncoding string
		wantGzip       bool
	}{
		{"simple gzip", "gzip", true},
		{"gzip with others", "gzip, deflate, br", true},
		{"gzip disabled", "gzip;q=0", false},
		{"gzip disabled with spaces", "gzip; q=0", false},
		{"empty", "", false},
		{"only deflate", "deflate", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			gotGzip := w.Header().Get("Content-Encoding") == "gzip"
			if gotGzip != tt.wantGzip {
				t.Errorf("gzip = %v, want %v", gotGzip, tt.wantGzip)
			}
		})
	}
}
