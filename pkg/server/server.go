package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/development-and-dinosaurs/diplodocs/pkg/builder"
	"github.com/development-and-dinosaurs/diplodocs/pkg/config"
	"github.com/fsnotify/fsnotify"
)

// DevServer orchestrates live file watching and HTTP serving.
type DevServer struct {
	ProjectRoot string
	Port        int
	cfg         *config.Config
	clients     map[chan string]bool
	clientsMu   sync.Mutex
}

// NewDevServer creates an instance of DevServer.
func NewDevServer(projectRoot string, port int) (*DevServer, error) {
	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return nil, err
	}
	return &DevServer{
		ProjectRoot: projectRoot,
		Port:        port,
		cfg:         cfg,
		clients:     make(map[chan string]bool),
	}, nil
}

func (s *DevServer) broadcastReload() {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for ch := range s.clients {
		select {
		case ch <- "reload":
		default:
		}
	}
}

// Start watches project files, runs the initial build, and starts http.ListenAndServe.
func (s *DevServer) Start() error {
	// Initial build in dev mode
	stats, err := builder.Build(s.ProjectRoot, true)
	if err != nil {
		log.Printf("⚠️ Initial build warning: %v", err)
	} else {
		log.Printf("🦕 Built %d pages in %v", stats.PageCount, stats.Duration)
	}

	// Setup file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer watcher.Close()

	docsDir := filepath.Join(s.ProjectRoot, s.cfg.DocsDir)
	_ = filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			_ = watcher.Add(path)
		}
		return nil
	})

	configFile := filepath.Join(s.ProjectRoot, "diplodocs.toml")
	if _, err := os.Stat(configFile); err == nil {
		_ = watcher.Add(configFile)
	}

	// Watcher goroutine with debounce
	go func() {
		var debounceTimer *time.Timer
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Ignore hidden or output directory events
				if strings.Contains(event.Name, s.cfg.OutDir) || strings.HasPrefix(filepath.Base(event.Name), ".") {
					continue
				}

				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					// Add newly created directories to watcher
					if event.Has(fsnotify.Create) {
						if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
							_ = watcher.Add(event.Name)
						}
					}

					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					debounceTimer = time.AfterFunc(100*time.Millisecond, func() {
						log.Printf("🦕 Change detected (%s). Rebuilding...", filepath.Base(event.Name))
						bStats, bErr := builder.Build(s.ProjectRoot, true)
						if bErr != nil {
							log.Printf("❌ Rebuild error: %v", bErr)
						} else {
							log.Printf("⚡ Rebuilt %d pages in %v. Reloading browser.", bStats.PageCount, bStats.Duration)
							s.broadcastReload()
						}
					})
				}
			case wErr, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", wErr)
			}
		}
	}()

	// HTTP Multiplexer
	mux := http.NewServeMux()

	// Live reload SSE endpoint
	mux.HandleFunc("/diplodocs-livereload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		reloadChan := make(chan string, 1)
		s.clientsMu.Lock()
		s.clients[reloadChan] = true
		s.clientsMu.Unlock()

		defer func() {
			s.clientsMu.Lock()
			delete(s.clients, reloadChan)
			s.clientsMu.Unlock()
		}()

		// Send connected message
		fmt.Fprintf(w, "data: connected\n\n")
		flusher.Flush()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-reloadChan:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			}
		}
	})

	// Static file server
	distDir := filepath.Join(s.ProjectRoot, s.cfg.OutDir)
	fileServer := http.FileServer(http.Dir(distDir))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(distDir, filepath.Clean(r.URL.Path))
		// If path doesn't exist, check if appending .html works
		if fi, err := os.Stat(p); err != nil || fi.IsDir() {
			if fi != nil && fi.IsDir() {
				indexPath := filepath.Join(p, "index.html")
				if _, err := os.Stat(indexPath); err == nil {
					fileServer.ServeHTTP(w, r)
					return
				}
			}
			htmlPath := p + ".html"
			if _, err := os.Stat(htmlPath); err == nil {
				r.URL.Path += ".html"
			}
		}
		fileServer.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", s.Port)
	log.Printf("🦕 Diplodocs dev server running at: http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
}
