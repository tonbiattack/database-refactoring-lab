package order

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func NewHandler(service *Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/orders/")
		parts := strings.Split(path, "/")
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil || id < 1 {
			writeError(w, http.StatusBadRequest, "invalid order id")
			return
		}
		if len(parts) == 1 && r.Method == http.MethodGet {
			order, err := service.Get(r.Context(), id)
			if err != nil {
				writeError(w, http.StatusNotFound, "order not found")
				return
			}
			writeJSON(w, http.StatusOK, order)
			return
		}
		if len(parts) == 2 && parts[1] == "status" && r.Method == http.MethodPut {
			var request struct {
				Status string `json:"status"`
			}
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&request); err != nil || request.Status == "" {
				writeError(w, http.StatusBadRequest, "status is required")
				return
			}
			if err := service.UpdateStatus(r.Context(), id, request.Status); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": request.Status})
			return
		}
		writeError(w, http.StatusNotFound, "not found")
	})
	return mux
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
