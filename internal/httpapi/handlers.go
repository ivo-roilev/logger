package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"logger/internal/format"
	"logger/internal/model"
	"logger/internal/sink"
)

// LoggerHandler handles log ingestion over HTTP.
type LoggerHandler struct {
	Sink sink.Sink
}

// NewLoggerHandler constructs a LoggerHandler.
func NewLoggerHandler(s sink.Sink) *LoggerHandler {
	return &LoggerHandler{Sink: s}
}

// PostLog handles POST /logs.
func (h *LoggerHandler) PostLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ct := r.Header.Get("Content-Type")
	if ct == "" || !strings.HasPrefix(strings.ToLower(ct), "application/json") {
		writeJSONError(w, http.StatusBadRequest, "Content-Type must be application/json")
		return
	}

	defer r.Body.Close()

	var payload model.EventPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ev, err := payload.ToEvent()
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	line, err := format.FormatEvent(ev)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to format event")
		return
	}

	if err := h.Sink.WriteLine(r.Context(), line); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to write log")
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

