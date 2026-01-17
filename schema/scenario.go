package schema

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
)

// Scenario defines the interface for different database schema scenarios.
// Each scenario encapsulates the read/write operations for a specific table structure.
type Scenario interface {
	// Name returns the unique identifier for this scenario
	Name() string

	// Description returns a human-readable description
	Description() string

	// TableName returns the primary table name used by this scenario
	TableName() string

	// MaxID returns the maximum ID for read operations (for seeded data)
	MaxID() int64

	// ExecuteRead performs a read operation using this scenario's query
	ExecuteRead(ctx context.Context, conn *pgx.Conn) error

	// ExecuteWrite performs a write operation using this scenario's query
	ExecuteWrite(ctx context.Context, conn *pgx.Conn) error

	// Initialize performs any necessary setup (e.g., table introspection for custom scenarios)
	Initialize(ctx context.Context, conn *pgx.Conn) error
}

// ScenarioInfo contains metadata about a scenario for API responses
type ScenarioInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	TableName   string `json:"table_name"`
}

// Registry holds all available scenarios
type Registry struct {
	mu        sync.RWMutex
	scenarios map[string]Scenario
}

// NewRegistry creates a new scenario registry with builtin scenarios
func NewRegistry() *Registry {
	r := &Registry{
		scenarios: make(map[string]Scenario),
	}

	// Register builtin scenarios
	r.Register(NewSimpleScenario())
	r.Register(NewJSONBScenario())
	r.Register(NewWideScenario())
	r.Register(NewFKScenario())

	return r
}

// Register adds a scenario to the registry
func (r *Registry) Register(s Scenario) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scenarios[s.Name()] = s
}

// Get retrieves a scenario by name
func (r *Registry) Get(name string) (Scenario, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.scenarios[name]
	return s, ok
}

// List returns info about all registered scenarios
func (r *Registry) List() []ScenarioInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]ScenarioInfo, 0, len(r.scenarios))
	// Return in a consistent order
	order := []string{"simple", "jsonb", "wide", "fk"}
	for _, name := range order {
		if s, ok := r.scenarios[name]; ok {
			list = append(list, ScenarioInfo{
				Name:        s.Name(),
				Description: s.Description(),
				TableName:   s.TableName(),
			})
		}
	}
	// Add any others not in the predefined order
	for name, s := range r.scenarios {
		found := false
		for _, o := range order {
			if o == name {
				found = true
				break
			}
		}
		if !found {
			list = append(list, ScenarioInfo{
				Name:        s.Name(),
				Description: s.Description(),
				TableName:   s.TableName(),
			})
		}
	}
	return list
}

// CreateCustomScenario creates a custom scenario for a specific table
func (r *Registry) CreateCustomScenario(tableName string) *CustomScenario {
	return NewCustomScenario(tableName)
}
