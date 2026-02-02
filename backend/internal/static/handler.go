package static

import (
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

// basePath is stripped from incoming requests to match nginx rewrite behavior.
const basePath = "/schedule-optimizer"

// Handler returns an http.Handler that serves static files from the embedded filesystem.
// It serves pre-compressed .gz files when the client accepts gzip encoding.
func Handler() http.Handler {
	fsys, err := fs.Sub(Dist, "dist")
	if err != nil {
		panic("failed to create sub filesystem: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(fsys))

	// Build set of gzipped files at startup (one-time scan)
	gzipped := make(map[string]bool)
	_ = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if original, found := strings.CutSuffix(path, ".gz"); found {
			gzipped[original] = true
		}
		return nil
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strip base path prefix (matches nginx rewrite behavior)
		urlPath := strings.TrimPrefix(r.URL.Path, basePath)
		if urlPath == "" {
			urlPath = "/"
		}

		// Determine the logical file path (for gzip lookup)
		filePath := strings.TrimPrefix(urlPath, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		// Serve gzipped version if available and accepted
		// Safe: filePath only matches keys from our pre-built gzipped map
		if acceptsGzip(r) && gzipped[filePath] {
			if serveGzipped(w, r, fsys, filePath) {
				return
			}
		}

		// Add cache headers for hashed assets (non-gzip path)
		if strings.HasPrefix(filePath, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}

		// Fall back to http.FileServer for non-gzip requests
		r.URL.Path = urlPath
		fileServer.ServeHTTP(w, r)
	})
}

// acceptsGzip checks if the client accepts gzip encoding.
func acceptsGzip(r *http.Request) bool {
	for enc := range strings.SplitSeq(r.Header.Get("Accept-Encoding"), ",") {
		enc = strings.TrimSpace(enc)
		// Check for "gzip" without q=0 (disabled)
		if enc == "gzip" || strings.HasPrefix(enc, "gzip;") && !strings.Contains(enc, "q=0") {
			return true
		}
	}
	return false
}

// serveGzipped serves a pre-compressed .gz file with appropriate headers.
// filePath must be a key from the gzipped map (safe, not user-controlled).
func serveGzipped(w http.ResponseWriter, r *http.Request, fsys fs.FS, filePath string) bool {
	f, err := fsys.Open(filePath + ".gz")
	if err != nil {
		return false
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return false
	}

	// Set Content-Type based on original file extension (not .gz)
	if ct := mime.TypeByExtension(filepath.Ext(filePath)); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Vary", "Accept-Encoding")

	// Immutable cache for hashed assets
	if strings.HasPrefix(filePath, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}

	// embed.FS files always implement io.ReadSeeker, but check defensively
	rs, ok := f.(io.ReadSeeker)
	if !ok {
		return false
	}
	http.ServeContent(w, r, filePath, stat.ModTime(), rs)
	return true
}
