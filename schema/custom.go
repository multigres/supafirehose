package schema

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
)

// ColumnInfo holds metadata about a table column
type ColumnInfo struct {
	Name       string
	DataType   string
	IsNullable bool
	HasDefault bool
	IsSerial   bool // SERIAL/BIGSERIAL columns (auto-generated)
}

// CustomScenario dynamically introspects a table and generates appropriate queries
type CustomScenario struct {
	tableName   string
	schemaName  string
	initialized bool
	mu          sync.RWMutex

	// Discovered column information
	columns       []ColumnInfo
	insertColumns []ColumnInfo // Columns we can insert into (excludes serials)
	primaryKey    string
	primaryKeyType string  // "uuid", "integer", etc.

	// Pre-built queries
	insertQuery string
	selectQuery string

	// For read operations
	maxID int64
	ids   []string // Cache for UUIDs or non-sequential IDs
}

// NewCustomScenario creates a new custom scenario for the given table
func NewCustomScenario(tableName string) *CustomScenario {
	schemaName := "public"
	tblName := tableName
	
	if tableName != "" && strings.Contains(tableName, ".") {
		parts := strings.SplitN(tableName, ".", 2)
		schemaName = parts[0]
		tblName = parts[1]
	}

	return &CustomScenario{
		tableName:  tblName,
		schemaName: schemaName,
		maxID:      100000,
		ids:        make([]string, 0, 10000),
	}
}

func (s *CustomScenario) Name() string {
	if s.tableName == "" {
		return "custom:auto"
	}
	return "custom:" + s.tableName
}

func (s *CustomScenario) Description() string {
	if s.tableName == "" {
		return "Custom table: <auto-discover>"
	}
	return fmt.Sprintf("Custom table: %s", s.tableName)
}

func (s *CustomScenario) TableName() string {
	return s.tableName
}

func (s *CustomScenario) MaxID() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.maxID
}

// Initialize introspects the table structure and builds queries
func (s *CustomScenario) Initialize(ctx context.Context, conn *pgx.Conn) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return nil
	}

	// Auto-discover table if not specified
	if s.tableName == "" {
		err := conn.QueryRow(ctx, `
			SELECT table_schema, table_name 
			FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_type = 'BASE TABLE'
			ORDER BY table_name LIMIT 1
		`).Scan(&s.schemaName, &s.tableName)
		
		if err != nil {
			return fmt.Errorf("failed to auto-discover table: %w", err)
		}
	}

	// Get column information from information_schema
	// Fix: use COALESCE for is_serial to avoid NULL scanning error
	rows, err := conn.Query(ctx, `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable = 'YES' as is_nullable,
			c.column_default IS NOT NULL as has_default,
			COALESCE((c.column_default LIKE 'nextval%'), false) as is_serial
		FROM information_schema.columns c
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position
	`, s.schemaName, s.tableName)
	if err != nil {
		return fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	s.columns = nil
	s.insertColumns = nil

	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Name, &col.DataType, &col.IsNullable, &col.HasDefault, &col.IsSerial); err != nil {
			return fmt.Errorf("failed to scan column info: %w", err)
		}
		s.columns = append(s.columns, col)

		// Include columns that are not serial/auto-generated for inserts
		if !col.IsSerial {
			s.insertColumns = append(s.insertColumns, col)
		}
	}

	if len(s.columns) == 0 {
		return fmt.Errorf("table %s.%s not found or has no columns", s.schemaName, s.tableName)
	}

	// Get primary key column and type
	var pkType string
	err = conn.QueryRow(ctx, `
		SELECT a.attname, format_type(a.atttypid, a.atttypmod) as data_type
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = $1::regclass AND i.indisprimary
		LIMIT 1
	`, fmt.Sprintf("%s.%s", s.schemaName, s.tableName)).Scan(&s.primaryKey, &pkType)
	
	if err != nil {
		// No primary key found, use first column
		s.primaryKey = s.columns[0].Name
		s.primaryKeyType = s.columns[0].DataType
	} else {
		s.primaryKeyType = pkType
	}

	// Determine strategy based on PK type (int vs uuid/other)
	// We default to integer strategy ONLY if it strongly looks like an integer/serial.
	// otherwise we default to the safer ID caching strategy (which handles UUIDs, text, sparseness)
	lowerType := strings.ToLower(s.primaryKeyType)
	isIntegerPK := (strings.Contains(lowerType, "int") || strings.Contains(lowerType, "serial")) && 
	               !strings.Contains(lowerType, "uuid") // Explicitly exclude UUID if it somehow matches "int" (unlikely but safe)

	if isIntegerPK {
		// Integer strategy: Get max ID
		var maxID *int64
		err = conn.QueryRow(ctx, fmt.Sprintf(
			"SELECT MAX(%s) FROM %s.%s",
			s.primaryKey, s.schemaName, s.tableName,
		)).Scan(&maxID)
		if err == nil && maxID != nil {
			s.maxID = *maxID
		}
		if s.maxID < 1 {
			s.maxID = 1 // At least 1
		}
	} else {
		// UUID/String strategy: Cache IDs
		// We explicitly cast to text to ensure scanning works for UUID/VARCHAR/etc
		idRows, err := conn.Query(ctx, fmt.Sprintf(
			"SELECT %s::text FROM %s.%s LIMIT 10000",
			s.primaryKey, s.schemaName, s.tableName,
		))
		if err == nil {
			defer idRows.Close()
			for idRows.Next() {
				var id string
				if err := idRows.Scan(&id); err == nil {
					s.ids = append(s.ids, id)
				}
			}
		}
	}

	// Build INSERT query
	if len(s.insertColumns) > 0 {
		colNames := make([]string, len(s.insertColumns))
		placeholders := make([]string, len(s.insertColumns))
		for i, col := range s.insertColumns {
			colNames[i] = col.Name
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		// Always return as text to be safe
		s.insertQuery = fmt.Sprintf(
			"INSERT INTO %s.%s (%s) VALUES (%s) RETURNING %s::text",
			s.schemaName, s.tableName,
			strings.Join(colNames, ", "),
			strings.Join(placeholders, ", "),
			s.primaryKey,
		)
	}

	// Build SELECT query
	colNames := make([]string, len(s.columns))
	for i, col := range s.columns {
		colNames[i] = col.Name
	}
	// Always cast parameter to text in application logic (handled by driver usually, but good to be consistent)
	s.selectQuery = fmt.Sprintf(
		"SELECT %s FROM %s.%s WHERE %s = $1",
		strings.Join(colNames, ", "),
		s.schemaName, s.tableName,
		s.primaryKey,
	)

	s.initialized = true
	return nil
}

func (s *CustomScenario) ExecuteRead(ctx context.Context, conn *pgx.Conn) error {
	s.mu.RLock()
	if !s.initialized {
		s.mu.RUnlock()
		if err := s.Initialize(ctx, conn); err != nil {
			return err
		}
		s.mu.RLock()
	}
	query := s.selectQuery
	maxID := s.maxID
	numCols := len(s.columns)
	cachedIDs := len(s.ids)
	s.mu.RUnlock()

	if query == "" {
		return fmt.Errorf("custom scenario not initialized")
	}

	// Determine ID to read
	var id interface{}
	if cachedIDs > 0 {
		s.mu.RLock()
		if len(s.ids) > 0 {
			id = s.ids[rand.Intn(len(s.ids))]
		}
		s.mu.RUnlock()
	} else if maxID > 0 {
		id = rand.Int63n(maxID) + 1
	}

	// If we still don't have an ID (empty table or failed init), we can't read
	if id == nil {
		return nil // Behave like a no-op instead of erroring
	}

	// Create scan destinations for all columns
	destinations := make([]interface{}, numCols)
	values := make([]interface{}, numCols)
	for i := range destinations {
		destinations[i] = &values[i]
	}

	return conn.QueryRow(ctx, query, id).Scan(destinations...)
}

func (s *CustomScenario) ExecuteWrite(ctx context.Context, conn *pgx.Conn) error {
	s.mu.RLock()
	if !s.initialized {
		s.mu.RUnlock()
		if err := s.Initialize(ctx, conn); err != nil {
			return err
		}
		s.mu.RLock()
	}
	query := s.insertQuery
	insertColumns := s.insertColumns
	s.mu.RUnlock()

	if query == "" || len(insertColumns) == 0 {
		return fmt.Errorf("custom scenario not initialized or no insertable columns")
	}

	// Generate values for each insert column
	args := make([]interface{}, len(insertColumns))
	for i, col := range insertColumns {
		args[i] = GenerateValue(col.DataType, col.Name)
	}

	var newID string
	err := conn.QueryRow(ctx, query, args...).Scan(&newID)

	// If successful and we have a new ID, cache it if using ID cache strategy
	if err == nil && newID != "" {
		s.mu.Lock()
		if len(s.ids) > 0 || s.maxID == 0 { // Use cache if we have IDs or maxID is 0 (meaning not int strategy)
			if len(s.ids) < 10000 {
				s.ids = append(s.ids, newID)
			} else {
				s.ids[rand.Intn(len(s.ids))] = newID
			}
		}
		s.mu.Unlock()
	}

	return err
}

// IsInitialized returns whether the scenario has been initialized
func (s *CustomScenario) IsInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.initialized
}

// GetColumns returns the discovered columns (for debugging/info)
func (s *CustomScenario) GetColumns() []ColumnInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]ColumnInfo, len(s.columns))
	copy(result, s.columns)
	return result
}
