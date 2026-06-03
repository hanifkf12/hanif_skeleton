package sqlbuilder

import (
	"strings"
)

// DriverDetector helps detect database driver type from various sources
type DriverDetector struct{}

// NewDriverDetector creates a new driver detector
func NewDriverDetector() *DriverDetector {
	return &DriverDetector{}
}

// DetectFromDSN detects driver type from database connection string
func (d *DriverDetector) DetectFromDSN(dsn string) DriverType {
	dsnLower := strings.ToLower(dsn)

	// PostgreSQL patterns
	if strings.Contains(dsnLower, "postgres") ||
		strings.Contains(dsnLower, "postgresql") ||
		strings.HasPrefix(dsnLower, "postgres://") ||
		strings.Contains(dsnLower, "host=") && strings.Contains(dsnLower, "port=5432") {
		return DriverPostgreSQL
	}

	// MySQL patterns (default)
	return DriverMySQL
}

// DetectFromDriverName detects driver type from database driver name
func (d *DriverDetector) DetectFromDriverName(driverName string) DriverType {
	driverLower := strings.ToLower(driverName)

	switch driverLower {
	case "postgres", "postgresql", "pgx", "pq":
		return DriverPostgreSQL
	case "mysql", "mariadb":
		return DriverMySQL
	default:
		return DriverMySQL // default to MySQL
	}
}

// DetectFromString detects driver type from any string identifier
func (d *DriverDetector) DetectFromString(s string) DriverType {
	sLower := strings.ToLower(s)

	if strings.Contains(sLower, "postgres") ||
		strings.Contains(sLower, "pg") ||
		strings.Contains(sLower, "psql") {
		return DriverPostgreSQL
	}

	return DriverMySQL
}

// AutoDetectDriver is a convenience function for auto-detection
func AutoDetectDriver(identifier string) DriverType {
	detector := NewDriverDetector()

	// Try DSN detection first
	if strings.Contains(identifier, "://") || strings.Contains(identifier, "host=") {
		return detector.DetectFromDSN(identifier)
	}

	// Check for common PostgreSQL indicators
	identifierLower := strings.ToLower(identifier)
	if strings.Contains(identifierLower, "pg") ||
		strings.Contains(identifierLower, "postgres") ||
		strings.Contains(identifierLower, "psql") {
		return DriverPostgreSQL
	}

	// Try driver name detection for known driver names
	if len(identifier) < 20 { // likely a driver name
		knownDrivers := []string{"mysql", "mariadb", "postgres", "postgresql", "pgx", "pq"}
		for _, driver := range knownDrivers {
			if strings.EqualFold(identifier, driver) {
				return detector.DetectFromDriverName(identifier)
			}
		}
	}

	// Fallback to MySQL
	return DriverMySQL
}

// SetDefaultDriverFromEnv sets default driver based on environment variable
func SetDefaultDriverFromEnv(envValue string) {
	driver := AutoDetectDriver(envValue)
	SetDefaultDriver(driver)
}

// GetDriverName returns the string name of the driver
func GetDriverName(driver DriverType) string {
	switch driver {
	case DriverPostgreSQL:
		return "PostgreSQL"
	case DriverMySQL:
		return "MySQL"
	default:
		return "Unknown"
	}
}

// IsPostgreSQL checks if driver is PostgreSQL
func IsPostgreSQL(driver DriverType) bool {
	return driver == DriverPostgreSQL
}

// IsMySQL checks if driver is MySQL
func IsMySQL(driver DriverType) bool {
	return driver == DriverMySQL
}
