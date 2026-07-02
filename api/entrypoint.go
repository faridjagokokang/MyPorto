package handler

import (
	"cybernexus/backend"
	"encoding/json"
	"net/http"
	"sync"
)

var (
	mux     *http.ServeMux
	once    sync.Once
	initErr error
)

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		initErr = backend.InitDB() // Initialize DB and migrations once
		if initErr != nil {
			return
		}

		mux = http.NewServeMux()
		mux.HandleFunc("/api/articles", backend.HandleArticles)
		mux.HandleFunc("/api/users", backend.HandleUsers)
		mux.HandleFunc("/api/login", backend.HandleLogin)
		mux.HandleFunc("/api/register", backend.HandleRegister)
		mux.HandleFunc("/api/logs", backend.HandleGetLogs)
		mux.HandleFunc("/api/terminal/execute", backend.HandleTerminalExecute)
		mux.HandleFunc("/ws/logs", backend.HandleWebSocket)
		mux.HandleFunc("/api/analytics", backend.HandleAnalytics)
		mux.HandleFunc("/api/leaderboard", backend.HandleLeaderboard)
		mux.HandleFunc("/api/contact", backend.HandleContact)
		mux.HandleFunc("/api/projects", backend.HandleProjects)
		mux.HandleFunc("/api/contacts", backend.HandleContacts)
		mux.HandleFunc("/api/login-logs", backend.HandleLoginLogs)
	})

	if initErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Database initialization failed: " + initErr.Error(),
			"note":  "Please check your DATABASE_URL environment variable and connection configuration.",
		})
		
		// Reset once to allow retrying on next request
		once = sync.Once{}
		return
	}

	mux.ServeHTTP(w, r)
}
