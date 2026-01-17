package schema

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

// SimpleScenario is the original users table scenario
type SimpleScenario struct {
	maxID int64
	ids   []string
	mu    sync.RWMutex
}

func NewSimpleScenario() *SimpleScenario {
	return &SimpleScenario{
		maxID: 100000,
		ids:   make([]string, 0, 10000),
	}
}

func (s *SimpleScenario) Name() string        { return "simple" }
func (s *SimpleScenario) Description() string { return "Simple users table (username, email)" }
func (s *SimpleScenario) TableName() string   { return "users" }
func (s *SimpleScenario) MaxID() int64        { return s.maxID }

func (s *SimpleScenario) Initialize(ctx context.Context, conn *pgx.Conn) error {
	// Pre-load some existing IDs to support UUIDs or string IDs
	rows, err := conn.Query(ctx, "SELECT id::text FROM users LIMIT 10000")
	if err != nil {
		return nil // Ignore error if table doesn't exist
	}
	defer rows.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			s.ids = append(s.ids, id)
		}
	}
	return nil
}

func (s *SimpleScenario) ExecuteRead(ctx context.Context, conn *pgx.Conn) error {
	// Try to use cached ID first (for UUID support), fallback to random int
	var id interface{}
	
	s.mu.RLock()
	if len(s.ids) > 0 {
		id = s.ids[rand.Intn(len(s.ids))]
	} else {
		id = rand.Int63n(s.maxID) + 1
	}
	s.mu.RUnlock()

	var userID string
	var username, email string
	var createdAt time.Time
	
	return conn.QueryRow(ctx,
		"SELECT id::text, username, email, created_at FROM users WHERE id = $1",
		id,
	).Scan(&userID, &username, &email, &createdAt)
}

func (s *SimpleScenario) ExecuteWrite(ctx context.Context, conn *pgx.Conn) error {
	randNum := rand.Int63()
	username := fmt.Sprintf("user_%d", randNum)
	email := fmt.Sprintf("user_%d@example.com", randNum)

	var newID string
	err := conn.QueryRow(ctx,
		"INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id::text",
		username, email,
	).Scan(&newID)

	if err == nil && newID != "" {
		s.mu.Lock()
		if len(s.ids) < 10000 {
			s.ids = append(s.ids, newID)
		} else {
			s.ids[rand.Intn(len(s.ids))] = newID
		}
		s.mu.Unlock()
	}

	return err
}

// JSONBScenario uses a table with a JSONB payload column
type JSONBScenario struct {
	maxID int64
	ids   []string
	mu    sync.RWMutex
}

func NewJSONBScenario() *JSONBScenario {
	return &JSONBScenario{
		maxID: 100000,
		ids:   make([]string, 0, 10000),
	}
}

func (s *JSONBScenario) Name() string        { return "jsonb" }
func (s *JSONBScenario) Description() string { return "Table with JSONB payload column" }
func (s *JSONBScenario) TableName() string   { return "jsonb_data" }
func (s *JSONBScenario) MaxID() int64        { return s.maxID }

func (s *JSONBScenario) Initialize(ctx context.Context, conn *pgx.Conn) error {
	// Pre-load some existing IDs
	rows, err := conn.Query(ctx, "SELECT id::text FROM jsonb_data LIMIT 10000")
	if err != nil {
		return nil
	}
	defer rows.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			s.ids = append(s.ids, id)
		}
	}
	return nil
}

func (s *JSONBScenario) ExecuteRead(ctx context.Context, conn *pgx.Conn) error {
	var id interface{}
	
	s.mu.RLock()
	if len(s.ids) > 0 {
		id = s.ids[rand.Intn(len(s.ids))]
	} else {
		id = rand.Int63n(s.maxID) + 1
	}
	s.mu.RUnlock()

	var dataID int64
	var payload string
	var createdAt time.Time
	return conn.QueryRow(ctx,
		"SELECT id, payload, created_at FROM jsonb_data WHERE id = $1",
		id,
	).Scan(&dataID, &payload, &createdAt)
}

func (s *JSONBScenario) ExecuteWrite(ctx context.Context, conn *pgx.Conn) error {
	payload := generateJSON()

	var newID string
	err := conn.QueryRow(ctx,
		"INSERT INTO jsonb_data (payload) VALUES ($1::jsonb) RETURNING id::text",
		payload,
	).Scan(&newID)

	if err == nil && newID != "" {
		s.mu.Lock()
		if len(s.ids) < 10000 {
			s.ids = append(s.ids, newID)
		} else {
			s.ids[rand.Intn(len(s.ids))] = newID
		}
		s.mu.Unlock()
	}
	return err
}

// WideScenario uses a table with many columns
type WideScenario struct {
	maxID int64
	ids   []string
	mu    sync.RWMutex
}

func NewWideScenario() *WideScenario {
	return &WideScenario{
		maxID: 100000,
		ids:   make([]string, 0, 10000),
	}
}

func (s *WideScenario) Name() string        { return "wide" }
func (s *WideScenario) Description() string { return "Wide table with 20+ columns" }
func (s *WideScenario) TableName() string   { return "wide_data" }
func (s *WideScenario) MaxID() int64        { return s.maxID }

func (s *WideScenario) Initialize(ctx context.Context, conn *pgx.Conn) error {
	// Pre-load some existing IDs
	rows, err := conn.Query(ctx, "SELECT id::text FROM wide_data LIMIT 10000")
	if err != nil {
		return nil
	}
	defer rows.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			s.ids = append(s.ids, id)
		}
	}
	return nil
}

func (s *WideScenario) ExecuteRead(ctx context.Context, conn *pgx.Conn) error {
	var id interface{}
	
	s.mu.RLock()
	if len(s.ids) > 0 {
		id = s.ids[rand.Intn(len(s.ids))]
	} else {
		id = rand.Int63n(s.maxID) + 1
	}
	s.mu.RUnlock()

	// Read all columns
	var dataID int64
	var cols [20]string
	var ints [5]int32
	var createdAt time.Time

	return conn.QueryRow(ctx,
		`SELECT id,
			col_01, col_02, col_03, col_04, col_05,
			col_06, col_07, col_08, col_09, col_10,
			col_11, col_12, col_13, col_14, col_15,
			col_16, col_17, col_18, col_19, col_20,
			int_01, int_02, int_03, int_04, int_05,
			created_at
		FROM wide_data WHERE id = $1`,
		id,
	).Scan(&dataID,
		&cols[0], &cols[1], &cols[2], &cols[3], &cols[4],
		&cols[5], &cols[6], &cols[7], &cols[8], &cols[9],
		&cols[10], &cols[11], &cols[12], &cols[13], &cols[14],
		&cols[15], &cols[16], &cols[17], &cols[18], &cols[19],
		&ints[0], &ints[1], &ints[2], &ints[3], &ints[4],
		&createdAt,
	)
}

func (s *WideScenario) ExecuteWrite(ctx context.Context, conn *pgx.Conn) error {
	// Generate values for all 25 data columns
	args := make([]interface{}, 25)
	for i := 0; i < 20; i++ {
		args[i] = generateString(50)
	}
	for i := 20; i < 25; i++ {
		args[i] = rand.Int31()
	}

	var newID string
	err := conn.QueryRow(ctx,
		`INSERT INTO wide_data (
			col_01, col_02, col_03, col_04, col_05,
			col_06, col_07, col_08, col_09, col_10,
			col_11, col_12, col_13, col_14, col_15,
			col_16, col_17, col_18, col_19, col_20,
			int_01, int_02, int_03, int_04, int_05
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25
		) RETURNING id::text`,
		args...,
	).Scan(&newID)

	if err == nil && newID != "" {
		s.mu.Lock()
		if len(s.ids) < 10000 {
			s.ids = append(s.ids, newID)
		} else {
			s.ids[rand.Intn(len(s.ids))] = newID
		}
		s.mu.Unlock()
	}
	return err
}

// FKScenario uses tables with foreign key relationships
type FKScenario struct {
	maxCategoryID int64
	ids           []string
	mu            sync.RWMutex
}

func NewFKScenario() *FKScenario {
	return &FKScenario{
		maxCategoryID: 100,
		ids:           make([]string, 0, 10000),
	}
}

func (s *FKScenario) Name() string        { return "fk" }
func (s *FKScenario) Description() string { return "Tables with foreign key lookup" }
func (s *FKScenario) TableName() string   { return "items" }
func (s *FKScenario) MaxID() int64        { return 0 } // Not used for UUIDs

func (s *FKScenario) Initialize(ctx context.Context, conn *pgx.Conn) error {
	// Pre-load some existing IDs
	rows, err := conn.Query(ctx, "SELECT id::text FROM items LIMIT 10000")
	if err != nil {
		return nil // Ignore error if table doesn't exist or is empty
	}
	defer rows.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			s.ids = append(s.ids, id)
		}
	}
	return nil
}

func (s *FKScenario) ExecuteRead(ctx context.Context, conn *pgx.Conn) error {
	var id string

	s.mu.RLock()
	if len(s.ids) > 0 {
		id = s.ids[rand.Intn(len(s.ids))]
	}
	s.mu.RUnlock()

	// If we don't have any IDs yet, we can't efficiently read
	if id == "" {
		return nil
	}

	// Join query to read item with its category
	var itemID, categoryID string // IDs are potentially UUIDs or strings
	var itemName, categoryName string
	var createdAt time.Time

	return conn.QueryRow(ctx,
		`SELECT i.id::text, i.name, i.created_at, c.id::text, c.name
		FROM items i
		JOIN categories c ON i.category_id = c.id
		WHERE i.id = $1`,
		id,
	).Scan(&itemID, &itemName, &createdAt, &categoryID, &categoryName)
}

func (s *FKScenario) ExecuteWrite(ctx context.Context, conn *pgx.Conn) error {
	// Pick a random category
	categoryID := rand.Int63n(s.maxCategoryID) + 1
	name := fmt.Sprintf("item_%d", rand.Int63())

	var newID string
	err := conn.QueryRow(ctx,
		"INSERT INTO items (category_id, name) VALUES ($1, $2) RETURNING id::text",
		categoryID, name,
	).Scan(&newID)

	if err == nil && newID != "" {
		s.mu.Lock()
		// Keep cache strict size
		if len(s.ids) < 10000 {
			s.ids = append(s.ids, newID)
		} else {
			// Random replacement to keep cache fresh
			s.ids[rand.Intn(len(s.ids))] = newID
		}
		s.mu.Unlock()
	}

	return err
}

// Helper to build a column list for queries
func buildColumnList(columns []string, prefix string) string {
	if prefix == "" {
		return strings.Join(columns, ", ")
	}
	prefixed := make([]string, len(columns))
	for i, col := range columns {
		prefixed[i] = prefix + "." + col
	}
	return strings.Join(prefixed, ", ")
}

// Helper to build placeholder list ($1, $2, $3, ...)
func buildPlaceholders(count int) string {
	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	return strings.Join(placeholders, ", ")
}
