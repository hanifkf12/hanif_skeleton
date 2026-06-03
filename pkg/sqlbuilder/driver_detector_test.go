package sqlbuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDriverDetector_DetectFromDSN(t *testing.T) {
	detector := NewDriverDetector()

	tests := []struct {
		name     string
		dsn      string
		expected DriverType
	}{
		{
			name:     "PostgreSQL with postgres://",
			dsn:      "postgres://user:pass@localhost:5432/dbname",
			expected: DriverPostgreSQL,
		},
		{
			name:     "PostgreSQL with postgresql://",
			dsn:      "postgresql://user:pass@localhost:5432/dbname",
			expected: DriverPostgreSQL,
		},
		{
			name:     "PostgreSQL with host=",
			dsn:      "host=localhost port=5432 user=postgres password=secret dbname=mydb",
			expected: DriverPostgreSQL,
		},
		{
			name:     "MySQL DSN",
			dsn:      "user:password@tcp(localhost:3306)/dbname",
			expected: DriverMySQL,
		},
		{
			name:     "MySQL with mysql://",
			dsn:      "mysql://user:pass@localhost:3306/dbname",
			expected: DriverMySQL,
		},
		{
			name:     "Empty DSN defaults to MySQL",
			dsn:      "",
			expected: DriverMySQL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectFromDSN(tt.dsn)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDriverDetector_DetectFromDriverName(t *testing.T) {
	detector := NewDriverDetector()

	tests := []struct {
		name       string
		driverName string
		expected   DriverType
	}{
		{
			name:       "postgres",
			driverName: "postgres",
			expected:   DriverPostgreSQL,
		},
		{
			name:       "postgresql",
			driverName: "postgresql",
			expected:   DriverPostgreSQL,
		},
		{
			name:       "pgx",
			driverName: "pgx",
			expected:   DriverPostgreSQL,
		},
		{
			name:       "pq",
			driverName: "pq",
			expected:   DriverPostgreSQL,
		},
		{
			name:       "mysql",
			driverName: "mysql",
			expected:   DriverMySQL,
		},
		{
			name:       "mariadb",
			driverName: "mariadb",
			expected:   DriverMySQL,
		},
		{
			name:       "unknown defaults to MySQL",
			driverName: "unknown",
			expected:   DriverMySQL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectFromDriverName(tt.driverName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDriverDetector_DetectFromString(t *testing.T) {
	detector := NewDriverDetector()

	tests := []struct {
		name     string
		input    string
		expected DriverType
	}{
		{
			name:     "contains postgres",
			input:    "my_postgres_db",
			expected: DriverPostgreSQL,
		},
		{
			name:     "contains pg",
			input:    "pg_connection",
			expected: DriverPostgreSQL,
		},
		{
			name:     "contains psql",
			input:    "psql_client",
			expected: DriverPostgreSQL,
		},
		{
			name:     "contains mysql",
			input:    "mysql_db",
			expected: DriverMySQL,
		},
		{
			name:     "generic string defaults to MySQL",
			input:    "database",
			expected: DriverMySQL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectFromString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAutoDetectDriver(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		expected   DriverType
	}{
		{
			name:       "PostgreSQL DSN",
			identifier: "postgres://localhost:5432/db",
			expected:   DriverPostgreSQL,
		},
		{
			name:       "MySQL DSN",
			identifier: "user:pass@tcp(localhost:3306)/db",
			expected:   DriverMySQL,
		},
		{
			name:       "Driver name: postgres",
			identifier: "postgres",
			expected:   DriverPostgreSQL,
		},
		{
			name:       "Driver name: mysql",
			identifier: "mysql",
			expected:   DriverMySQL,
		},
		{
			name:       "String: pg_database",
			identifier: "pg_database",
			expected:   DriverPostgreSQL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AutoDetectDriver(tt.identifier)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDriverName(t *testing.T) {
	assert.Equal(t, "PostgreSQL", GetDriverName(DriverPostgreSQL))
	assert.Equal(t, "MySQL", GetDriverName(DriverMySQL))
	assert.Equal(t, "Unknown", GetDriverName(DriverType(999)))
}

func TestIsPostgreSQL(t *testing.T) {
	assert.True(t, IsPostgreSQL(DriverPostgreSQL))
	assert.False(t, IsPostgreSQL(DriverMySQL))
}

func TestIsMySQL(t *testing.T) {
	assert.True(t, IsMySQL(DriverMySQL))
	assert.False(t, IsMySQL(DriverPostgreSQL))
}

func TestSetDefaultDriverFromEnv(t *testing.T) {
	// Test PostgreSQL
	SetDefaultDriverFromEnv("postgres")
	qb := NewQueryBuilder()
	query, _ := qb.Table("users").Where("id = ?", 1).Build()
	assert.Contains(t, query, "$1")

	// Test MySQL
	SetDefaultDriverFromEnv("mysql")
	qb2 := NewQueryBuilder()
	query2, _ := qb2.Table("users").Where("id = ?", 1).Build()
	assert.Contains(t, query2, "?")
}
