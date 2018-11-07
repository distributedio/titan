package logbunny

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// NewHTTPHandler will generate a http handler which implement the http httphandler
func NewHTTPHandler(log Logger) *HTTPHandler {
	return &HTTPHandler{logger: log}
}

// defined the http server request & respones
type httpPayload struct {
	Level string `json:"level"`
}

// defined the http erro respones
type errorResponse struct {
	Error string `json:"error"`
}

// HTTPHandler implement the htttp handler
type HTTPHandler struct {
	logger Logger
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.getLevel(w, r)
	case "PUT":
		h.putLevel(w, r)
	default:
		h.error(w, "Only GET and PUT are supported.", http.StatusMethodNotAllowed)
	}
}

// getlevel & setlevel used for serve http , get the level and trans into string
func (h *HTTPHandler) getLevel(w http.ResponseWriter, r *http.Request) {
	lv := h.logger.Level()
	var current string
	switch lv {
	case DebugLevel:
		current = "debug"
	case InfoLevel:
		current = "info"
	case WarnLevel:
		current = "warn"
	case ErrorLevel:
		current = "error"
	case PanicLevel:
		current = "panic"
	case FatalLevel:
		current = "fatal"
	default:
		current = ""
	}
	json.NewEncoder(w).Encode(httpPayload{Level: current})
}

func (h *HTTPHandler) putLevel(w http.ResponseWriter, r *http.Request) {
	var p httpPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		h.error(
			w,
			fmt.Sprintf("Request body must be well-formed JSON: %v", err),
			http.StatusBadRequest,
		)
		return
	}
	if p.Level == "" {
		h.error(w, "Must specify a logging level.", http.StatusBadRequest)
		return
	}

	p.Level = strings.ToLower(p.Level)

	switch p.Level {
	case "debug":
		h.logger.SetLevel(DebugLevel)
	case "info":
		h.logger.SetLevel(InfoLevel)
	case "warn":
		h.logger.SetLevel(WarnLevel)
	case "error":
		h.logger.SetLevel(ErrorLevel)
	case "panic":
		h.logger.SetLevel(PanicLevel)
	case "fatal":
		h.logger.SetLevel(FatalLevel)
	default:
		h.error(w, "Missmatch the given logging level.", http.StatusBadRequest)
	}

	json.NewEncoder(w).Encode(p)
}

func (h *HTTPHandler) error(w http.ResponseWriter, msg string, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: msg})
}
