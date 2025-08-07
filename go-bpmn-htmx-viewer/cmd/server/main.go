package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath" // For ParseGlob if templates are nested

	"go-bpmn-htmx-viewer/internal/api_handler" // Added import
)

var (
	templates *template.Template
)

func init() {
	// Load templates
	var err error
	// Corrected path for ParseGlob assuming execution from project root
	// or WORKDIR is set to project root in Docker.
	templatesPattern := filepath.Join("templates", "*.html")
	templates, err = template.ParseGlob(templatesPattern)
	if err != nil {
		log.Fatalf("Error parsing templates: %v. Pattern used: %s", err, templatesPattern)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Strict path checking for root only.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Printf("Error executing index template: %v", err)
		// Send a generic error message to the client
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Placeholder renderBPMNHandler is removed.

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Static file server
	// The path "./static/" is relative to the working directory.
	// Ensure the Docker WORKDIR is the project root.
	staticDir := "./static/"
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", serveIndex)
	// Use the new handler from api_handler, passing the loaded templates
	http.HandleFunc("/render-bpmn", api_handler.CreateRenderBPMNHandler(templates))

	log.Printf("Server starting on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
