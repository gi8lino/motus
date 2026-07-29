package routes

import (
	"bytes"
	"compress/gzip"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	immutableCacheControl = "public, max-age=31536000, immutable"
	staticCacheControl    = "public, max-age=86400"
)

type staticFileHandler struct {
	files      fs.FS
	fileServer http.Handler
	compressed sync.Map
}

func newStaticFileHandler(files fs.FS) http.Handler {
	return &staticFileHandler{
		files:      files,
		fileServer: http.FileServer(http.FS(files)),
	}
}

func (h *staticFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/assets/") {
		w.Header().Set("Cache-Control", immutableCacheControl)
	} else {
		w.Header().Set("Cache-Control", staticCacheControl)
	}

	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if !isCompressibleAsset(name) {
		h.fileServer.ServeHTTP(w, r)
		return
	}
	w.Header().Add("Vary", "Accept-Encoding")
	if !acceptsGzip(r.Header.Get("Accept-Encoding")) {
		h.fileServer.ServeHTTP(w, r)
		return
	}

	data, err := fs.ReadFile(h.files, name)
	if err != nil || len(data) < 1024 {
		h.fileServer.ServeHTTP(w, r)
		return
	}

	compressed, err := h.gzipFile(name, data)
	if err != nil {
		h.fileServer.ServeHTTP(w, r)
		return
	}

	if contentType := mime.TypeByExtension(path.Ext(name)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Content-Encoding", "gzip")
	if info, statErr := fs.Stat(h.files, name); statErr == nil {
		http.ServeContent(
			w,
			r,
			path.Base(name),
			info.ModTime(),
			bytes.NewReader(compressed),
		)
		return
	}
	http.ServeContent(
		w,
		r,
		path.Base(name),
		time.Time{},
		bytes.NewReader(compressed),
	)
}

func (h *staticFileHandler) gzipFile(name string, data []byte) ([]byte, error) {
	if cached, ok := h.compressed.Load(name); ok {
		return cached.([]byte), nil
	}

	var output bytes.Buffer
	writer := gzip.NewWriter(&output)
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	compressed := output.Bytes()
	h.compressed.Store(name, compressed)
	return compressed, nil
}

func isCompressibleAsset(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".css", ".html", ".js", ".json", ".svg", ".txt":
		return true
	default:
		return false
	}
}

func acceptsGzip(value string) bool {
	wildcard := false
	for _, entry := range strings.Split(value, ",") {
		parts := strings.Split(strings.TrimSpace(entry), ";")
		encoding := strings.ToLower(strings.TrimSpace(parts[0]))
		quality := 1.0
		for _, parameter := range parts[1:] {
			keyValue := strings.SplitN(strings.TrimSpace(parameter), "=", 2)
			if len(keyValue) != 2 || !strings.EqualFold(keyValue[0], "q") {
				continue
			}
			if parsed, err := strconv.ParseFloat(keyValue[1], 64); err == nil {
				quality = parsed
			}
		}
		if encoding == "gzip" {
			return quality > 0
		}
		if encoding == "*" && quality > 0 {
			wildcard = true
		}
	}
	return wildcard
}
