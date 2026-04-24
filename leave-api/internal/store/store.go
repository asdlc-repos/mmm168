package store

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"leave-api/internal/models"
)

// Store is the in-memory data store for leave requests and balances.
type Store struct {
	mu       sync.RWMutex
	requests []*models.LeaveRequest
	balances map[string]map[string]float64
	// employeeManager maps employee IDs to their manager IDs
	employeeManager map[string]string
}

// New creates a new Store seeded with sample data.
func New() *Store {
	s := &Store{
		requests:        make([]*models.LeaveRequest, 0),
		balances:        make(map[string]map[string]float64),
		employeeManager: make(map[string]string),
	}
	s.seed()
	return s
}

func defaultBalance() map[string]float64 {
	return map[string]float64{
		"Annual": 15,
		"Sick":   10,
		"Casual": 5,
	}
}

func (s *Store) seed() {
	// Seed employee → manager mapping
	s.employeeManager["emp1"] = "mgr1"
	s.employeeManager["emp2"] = "mgr1"
	s.employeeManager["emp3"] = "mgr2"

	// Seed default balances for each employee
	for _, emp := range []string{"emp1", "emp2", "emp3"} {
		s.balances[emp] = defaultBalance()
	}

	now := time.Now().UTC()
	// Seed 3 pending requests
	s.requests = append(s.requests,
		&models.LeaveRequest{
			ID:         uuid.New().String(),
			EmployeeID: "emp1",
			ManagerID:  "mgr1",
			StartDate:  "2026-05-01",
			EndDate:    "2026-05-03",
			LeaveType:  "Annual",
			Status:     "pending",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		&models.LeaveRequest{
			ID:         uuid.New().String(),
			EmployeeID: "emp2",
			ManagerID:  "mgr1",
			StartDate:  "2026-05-10",
			EndDate:    "2026-05-10",
			LeaveType:  "Sick",
			Status:     "pending",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		&models.LeaveRequest{
			ID:         uuid.New().String(),
			EmployeeID: "emp3",
			ManagerID:  "mgr2",
			StartDate:  "2026-05-15",
			EndDate:    "2026-05-16",
			LeaveType:  "Casual",
			Status:     "pending",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	)
}

// AddRequest adds a new leave request to the store.
func (s *Store) AddRequest(req *models.LeaveRequest) {
	// Assign manager based on employee
	if mgr, ok := s.employeeManager[req.EmployeeID]; ok {
		req.ManagerID = mgr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests = append(s.requests, req)
}

// GetRequests returns filtered leave requests.
func (s *Store) GetRequests(employeeID, managerID, status string) []*models.LeaveRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.LeaveRequest, 0)
	for _, r := range s.requests {
		if employeeID != "" && r.EmployeeID != employeeID {
			continue
		}
		if managerID != "" && r.ManagerID != managerID {
			continue
		}
		if status != "" && r.Status != status {
			continue
		}
		result = append(result, r)
	}
	return result
}

// GetRequestByID returns a single leave request by ID.
func (s *Store) GetRequestByID(id string) *models.LeaveRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, r := range s.requests {
		if r.ID == id {
			return r
		}
	}
	return nil
}

// UpdateRequest updates the status/managerId/comment of a leave request.
func (s *Store) UpdateRequest(id, status, managerID, comment string) *models.LeaveRequest {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range s.requests {
		if r.ID == id {
			r.Status = status
			r.ManagerID = managerID
			if comment != "" {
				r.Comment = comment
			}
			r.UpdatedAt = time.Now().UTC()
			return r
		}
	}
	return nil
}

// GetBalance returns the leave balance for an employee.
func (s *Store) GetBalance(employeeID string) (map[string]float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bal, ok := s.balances[employeeID]
	if !ok {
		// Return default balances for unknown employees
		return defaultBalance(), false
	}
	// Return a copy
	copy := make(map[string]float64)
	for k, v := range bal {
		copy[k] = v
	}
	return copy, true
}
