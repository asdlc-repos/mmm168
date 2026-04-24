package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"leave-api/internal/models"
	"leave-api/internal/store"
)

// Handler holds the store and implements all HTTP handlers.
type Handler struct {
	store *store.Store
}

// New creates a new Handler.
func New(s *store.Store) *Handler {
	return &Handler{store: s}
}

// corsMiddleware sets CORS headers on all responses.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs each incoming request.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.String(), r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// RegisterRoutes sets up all routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) http.Handler {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/requests", h.handleRequests)
	mux.HandleFunc("/requests/", h.handleRequestByID)
	mux.HandleFunc("/balance", h.handleBalance)

	return loggingMiddleware(corsMiddleware(mux))
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error encoding JSON response: %v", err)
	}
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, models.ErrorResponse{Error: msg})
}

// handleHealth handles GET /health.
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleRequests routes GET and POST /requests.
func (h *Handler) handleRequests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listRequests(w, r)
	case http.MethodPost:
		h.createRequest(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// listRequests handles GET /requests with optional query filters.
func (h *Handler) listRequests(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	employeeID := q.Get("employee")
	managerID := q.Get("manager")
	status := q.Get("status")

	// Validate status if provided
	if status != "" && status != "pending" && status != "approved" && status != "rejected" {
		writeError(w, http.StatusBadRequest, "invalid status filter; must be pending, approved, or rejected")
		return
	}

	requests := h.store.GetRequests(employeeID, managerID, status)
	writeJSON(w, http.StatusOK, requests)
}

// createRequest handles POST /requests.
func (h *Handler) createRequest(w http.ResponseWriter, r *http.Request) {
	var input models.CreateLeaveRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Validate required fields
	if input.EmployeeID == "" {
		writeError(w, http.StatusBadRequest, "employeeId is required")
		return
	}
	if input.StartDate == "" {
		writeError(w, http.StatusBadRequest, "startDate is required")
		return
	}
	if input.EndDate == "" {
		writeError(w, http.StatusBadRequest, "endDate is required")
		return
	}
	if input.LeaveType == "" {
		writeError(w, http.StatusBadRequest, "leaveType is required")
		return
	}
	if !models.ValidLeaveTypes[input.LeaveType] {
		writeError(w, http.StatusBadRequest, "leaveType must be one of: Annual, Sick, Casual")
		return
	}

	// Validate date format and range
	const dateLayout = "2006-01-02"
	start, err := time.Parse(dateLayout, input.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "startDate must be in YYYY-MM-DD format")
		return
	}
	end, err := time.Parse(dateLayout, input.EndDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "endDate must be in YYYY-MM-DD format")
		return
	}
	if end.Before(start) {
		writeError(w, http.StatusBadRequest, "endDate must be on or after startDate")
		return
	}

	now := time.Now().UTC()
	req := &models.LeaveRequest{
		ID:         uuid.New().String(),
		EmployeeID: input.EmployeeID,
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
		LeaveType:  input.LeaveType,
		Status:     "pending",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	h.store.AddRequest(req)
	writeJSON(w, http.StatusCreated, req)
}

// handleRequestByID routes PATCH /requests/{id}.
func (h *Handler) handleRequestByID(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the path
	id := strings.TrimPrefix(r.URL.Path, "/requests/")
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "request ID is required")
		return
	}

	switch r.Method {
	case http.MethodPatch:
		h.updateRequest(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// updateRequest handles PATCH /requests/{id}.
func (h *Handler) updateRequest(w http.ResponseWriter, r *http.Request, id string) {
	// Check the request exists
	existing := h.store.GetRequestByID(id)
	if existing == nil {
		writeError(w, http.StatusNotFound, "request not found")
		return
	}

	var input models.UpdateLeaveRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Validate required fields
	if input.Status == "" {
		writeError(w, http.StatusBadRequest, "status is required")
		return
	}
	if !models.ValidUpdateStatuses[input.Status] {
		writeError(w, http.StatusBadRequest, "status must be approved or rejected")
		return
	}
	if input.ManagerID == "" {
		writeError(w, http.StatusBadRequest, "managerId is required")
		return
	}

	updated := h.store.UpdateRequest(id, input.Status, input.ManagerID, input.Comment)
	if updated == nil {
		writeError(w, http.StatusNotFound, "request not found")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// handleBalance handles GET /balance.
func (h *Handler) handleBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	employeeID := r.URL.Query().Get("employeeId")
	if employeeID == "" {
		writeError(w, http.StatusBadRequest, "employeeId query parameter is required")
		return
	}

	balances, _ := h.store.GetBalance(employeeID)
	writeJSON(w, http.StatusOK, models.BalanceResponse{
		EmployeeID: employeeID,
		Balances:   balances,
	})
}
