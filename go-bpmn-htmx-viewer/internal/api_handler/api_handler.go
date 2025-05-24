package api_handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings" // For basic validation example

	"go-bpmn-htmx-viewer/internal/gcs_handler" // Adjust if module path is different
)

type RenderRequest struct {
	GCSURI string `json:"gcs_uri"`
}

type BPMNTemplateData struct {
	BPMNXMLDataJS string // BPMN XML data, escaped for JS string
	Error         string // Error message, if any
}

func CreateRenderBPMNHandler(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var reqPayload RenderRequest
		if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
			log.Printf("Error decoding request body: %v", err)
			// Prepare and send error HTML fragment
			data := BPMNTemplateData{Error: "Invalid request: Could not parse JSON."}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			// Consider a specific error status if HTMX can use it, but usually we send 200 with error in fragment
			// w.WriteHeader(http.StatusBadRequest) 
			tmplErr := templates.ExecuteTemplate(w, "bpmn_fragment.html", data)
			if tmplErr != nil {
				log.Printf("Error executing template to show JSON decoding error: %v", tmplErr)
				// Fallback if template execution itself fails for the error message
				// Avoid writing to w if headers already sent by ExecuteTemplate, though for a critical error like this,
				// it might be acceptable if the client gets a broken response.
				// Consider checking response committed status if a more robust solution is needed.
				// For now, just logging, as the original code also just logged.
			}
			return
		}

		if strings.TrimSpace(reqPayload.GCSURI) == "" {
			log.Println("GCS URI is empty in request")
			data := BPMNTemplateData{Error: "GCS URI cannot be empty."}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			// w.WriteHeader(http.StatusBadRequest)
			tmplErr := templates.ExecuteTemplate(w, "bpmn_fragment.html", data)
			if tmplErr != nil {
				log.Printf("Error executing template to show GCS URI empty error: %v", tmplErr)
			}
			return
		}

		log.Printf("Received request for GCS URI: %s", reqPayload.GCSURI)
		// Pass r.Context() for GCS operations that require it
		bpmnXML, err := gcs_handler.FetchBPMNFromGCS(r.Context(), reqPayload.GCSURI)
		
		var data BPMNTemplateData
		if err != nil {
			log.Printf("Error fetching BPMN from GCS (%s): %v", reqPayload.GCSURI, err)
			// User-friendly error, hide internal details like actual GCS path if sensitive
			errMsg := fmt.Sprintf("Failed to load diagram. Please check the URI and ensure the file exists and is accessible.")
			if strings.Contains(err.Error(), "object doesn't exist") || strings.Contains(err.Error(), "storage: object") { // Basic check
				errMsg = fmt.Sprintf("Failed to load diagram: The specified file at %s was not found or is not accessible.", reqPayload.GCSURI)
			} else if strings.Contains(err.Error(), "invalid GCS URI") {
				errMsg = fmt.Sprintf("Failed to load diagram: The provided GCS URI '%s' is invalid.", reqPayload.GCSURI)
			}
			data = BPMNTemplateData{Error: errMsg}
		} else {
			// Important: Ensure this data is safe to embed in a JS string.
			// Go's html/template does this by default if the context is correct (e.g., within <script>).
			// For direct JS string variable assignment, it should escape backticks, backslashes, etc.
			data = BPMNTemplateData{BPMNXMLDataJS: bpmnXML}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Ensure bpmn_fragment.html exists and is part of the parsed templates
		tmplErr := templates.ExecuteTemplate(w, "bpmn_fragment.html", data)
		if tmplErr != nil {
			log.Printf("Error executing bpmn_fragment template: %v", tmplErr)
			// If ExecuteTemplate fails, it might have already written a partial response
			// or set headers. Avoid writing http.Error if possible.
			// This check is a bit tricky; a common pattern is to check if headers are sent.
			// For simplicity here, just logging. A more robust app might have a way to know.
		}
	}
}
