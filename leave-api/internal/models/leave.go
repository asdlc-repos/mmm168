package models

import "time"

// LeaveRequest represents a leave request in the system.
type LeaveRequest struct {
	ID         string    `json:"id"`
	EmployeeID string    `json:"employeeId"`
	ManagerID  string    `json:"managerId"`
	StartDate  string    `json:"startDate"`
	EndDate    string    `json:"endDate"`
	LeaveType  string    `json:"leaveType"`
	Status     string    `json:"status"`
	Comment    string    `json:"comment,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// BalanceResponse is the response for the GET /balance endpoint.
type BalanceResponse struct {
	EmployeeID string             `json:"employeeId"`
	Balances   map[string]float64 `json:"balances"`
}

// CreateLeaveRequestInput is the request body for POST /requests.
type CreateLeaveRequestInput struct {
	EmployeeID string `json:"employeeId"`
	StartDate  string `json:"startDate"`
	EndDate    string `json:"endDate"`
	LeaveType  string `json:"leaveType"`
}

// UpdateLeaveRequestInput is the request body for PATCH /requests/{id}.
type UpdateLeaveRequestInput struct {
	Status    string `json:"status"`
	ManagerID string `json:"managerId"`
	Comment   string `json:"comment,omitempty"`
}

// ErrorResponse is a generic error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Valid leave types.
var ValidLeaveTypes = map[string]bool{
	"Annual": true,
	"Sick":   true,
	"Casual": true,
}

// Valid statuses for update.
var ValidUpdateStatuses = map[string]bool{
	"approved": true,
	"rejected": true,
}
