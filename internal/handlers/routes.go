package handlers

import (
	"net/http"
)

func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("HEAD /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Add a specific handler for favicon.ico to prevent it from being routed to the index page
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Only handle paths that weren't matched by more specific handlers
		if r.URL.Path != "/" {
			NotFoundHandler(w, r)
			return
		}
		// If somehow a request to "/" gets here, redirect to index handler
		IndexHandler(w, r)
	})
	mux.HandleFunc("GET /about", AboutHandler)
	mux.HandleFunc("GET /events", EventsHandler)
	mux.HandleFunc("GET /events/{id}", EventDetailHandler)
	mux.HandleFunc("GET /players", PlayersHandler)
	mux.HandleFunc("GET /players/{id}", PlayerDetailHandler)
	mux.HandleFunc("GET /leaderboards", LeaderboardsHandler)
	mux.HandleFunc("GET /decklists/{id}", DecklistHandler)

	mux.HandleFunc("GET /_/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return mux
}
