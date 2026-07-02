package main

import (
	"cybernexus/api"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Serve static files from the root directory
	fs := http.FileServer(http.Dir("./"))
	http.Handle("/", fs)

	// Route all API and WebSocket requests to the Vercel Go handler
	http.HandleFunc("/api/", handler.Handler)
	http.HandleFunc("/ws/", handler.Handler)

	port := ":8080"
	fmt.Printf("CYBERNEXUS Server started locally on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
