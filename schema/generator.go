package schema

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
)

// GenerateValue generates a random value appropriate for the given PostgreSQL column type.
// It uses column name hints to generate more realistic data.
func GenerateValue(colType string, colName string) interface{} {
	colType = strings.ToLower(colType)
	colName = strings.ToLower(colName)

	// Check for specific column name patterns first for realistic data
	switch {
	case strings.Contains(colName, "email"):
		return gofakeit.Email()
	case strings.Contains(colName, "username") || strings.Contains(colName, "user_name"):
		return gofakeit.Username()
	case strings.Contains(colName, "first_name") || strings.Contains(colName, "firstname"):
		return gofakeit.FirstName()
	case strings.Contains(colName, "last_name") || strings.Contains(colName, "lastname"):
		return gofakeit.LastName()
	case strings.Contains(colName, "full_name") || strings.Contains(colName, "fullname") || colName == "name":
		return gofakeit.Name()
	case strings.Contains(colName, "phone") || strings.Contains(colName, "cell") || strings.Contains(colName, "mobile"):
		return gofakeit.Phone()
	case strings.Contains(colName, "city"):
		return gofakeit.City()
	case strings.Contains(colName, "country"):
		return gofakeit.Country()
	case strings.Contains(colName, "state") || strings.Contains(colName, "province"):
		return gofakeit.State()
	case strings.Contains(colName, "zip") || strings.Contains(colName, "postal"):
		return gofakeit.Zip()
	case strings.Contains(colName, "address"):
		return gofakeit.Address().Address
	case strings.Contains(colName, "company") || strings.Contains(colName, "org"):
		return gofakeit.Company()
	case strings.Contains(colName, "job") || strings.Contains(colName, "title"):
		return gofakeit.JobTitle()
	case strings.Contains(colName, "bio") || strings.Contains(colName, "description"):
		return gofakeit.Sentence(10)
	case strings.Contains(colName, "url") || strings.Contains(colName, "link") || strings.Contains(colName, "website"):
		return gofakeit.URL()
	case strings.Contains(colName, "ipv4"):
		return gofakeit.IPv4Address()
	case strings.Contains(colName, "ipv6"):
		return gofakeit.IPv6Address()
	case strings.Contains(colName, "user_agent"):
		return gofakeit.UserAgent()
	}

	// Generate based on column type if no specific name match
	switch {
	case strings.HasPrefix(colType, "varchar"), strings.HasPrefix(colType, "character varying"),
		colType == "text", strings.HasPrefix(colType, "char"):
		if strings.Contains(colType, "(") {
			// Extract length if possible, otherwise default
			return gofakeit.Sentence(3)
		}
		return gofakeit.Sentence(5)

	case colType == "integer", colType == "int", colType == "int4":
		return int32(gofakeit.Number(0, 1000000))

	case colType == "bigint", colType == "int8":
		return int64(gofakeit.Number(0, 1000000000))

	case colType == "smallint", colType == "int2":
		return int16(gofakeit.Number(0, 32000))

	case colType == "boolean", colType == "bool":
		return gofakeit.Bool()

	case colType == "real", colType == "float4":
		return float32(gofakeit.Float64Range(0, 1000))

	case colType == "double precision", colType == "float8":
		return gofakeit.Float64Range(0, 10000)

	case strings.HasPrefix(colType, "numeric"), strings.HasPrefix(colType, "decimal"):
		return gofakeit.Float64Range(0, 100000)

	case colType == "uuid":
		return uuid.New().String()

	case colType == "timestamp", colType == "timestamp without time zone":
		return gofakeit.Date()

	case colType == "timestamptz", colType == "timestamp with time zone":
		return gofakeit.Date()

	case colType == "date":
		return gofakeit.Date().Format("2006-01-02")

	case colType == "time", colType == "time without time zone":
		return gofakeit.Date().Format("15:04:05")

	case colType == "timetz", colType == "time with time zone":
		return gofakeit.Date().Format("15:04:05-07:00")

	case colType == "jsonb", colType == "json":
		return generateJSON() // Keep existing simple JSON generator or use gofakeit structure

	case colType == "bytea":
		return []byte(gofakeit.Sentence(5))

	case strings.HasPrefix(colType, "interval"):
		return fmt.Sprintf("%d hours", gofakeit.Number(1, 24))

	default:
		// Default to string for unknown types
		return gofakeit.Sentence(5)
	}
}

func generateJSON() string {
	person := gofakeit.Person()
	data := map[string]interface{}{
		"id":        gofakeit.UUID(),
		"timestamp": time.Now().Unix(),
		"name":      person.FirstName + " " + person.LastName,
		"active":    gofakeit.Bool(),
		"score":     gofakeit.Float64Range(0, 100),
		"tags":      []string{gofakeit.Word(), gofakeit.Word()},
		"metadata": map[string]interface{}{
			"version": "1.0",
			"source":  "generated",
			"job":     person.Job.Title,
		},
	}
	bytes, _ := json.Marshal(data)
	return string(bytes)
}

func generateBytes(length int) []byte {
	b := make([]byte, length)
	rand.Read(b)
	return b
}

func generateString(maxLen int) string {
	return gofakeit.Sentence(maxLen/10 + 1) // Approx words to match length
}
