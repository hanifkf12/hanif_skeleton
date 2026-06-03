package sqlbuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test PostgreSQL Driver Support

func TestQueryBuilder_PostgreSQL_Select(t *testing.T) {
	qb := NewQueryBuilderWithDriver(DriverPostgreSQL)
	query, args := qb.
		Table("users").
		Select("id", "name", "email").
		Where("status = ?", "active").
		Where("age > ?", 18).
		Build()

	expectedQuery := "SELECT id, name, email FROM users WHERE status = $1 AND age > $2"
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, []interface{}{"active", 18}, args)
}

func TestQueryBuilder_PostgreSQL_WhereIn(t *testing.T) {
	qb := NewQueryBuilderWithDriver(DriverPostgreSQL)
	values := []interface{}{1, 2, 3}
	query, args := qb.
		Table("users").
		Select("*").
		WhereIn("id", values).
		Build()

	expectedQuery := "SELECT * FROM users WHERE id IN ($1, $2, $3)"
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, values, args)
}

func TestQueryBuilder_PostgreSQL_Insert(t *testing.T) {
	qb := NewQueryBuilderWithDriver(DriverPostgreSQL)
	query, args := qb.
		Table("users").
		Insert(map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		}).
		Build()

	assert.Contains(t, query, "INSERT INTO users")
	assert.Contains(t, query, "$1")
	assert.Contains(t, query, "$2")
	assert.Len(t, args, 2)
}

func TestQueryBuilder_PostgreSQL_Update(t *testing.T) {
	qb := NewQueryBuilderWithDriver(DriverPostgreSQL)
	query, args := qb.
		Table("users").
		Update(map[string]interface{}{
			"name": "John Updated",
		}).
		Where("id = ?", 1).
		Build()

	assert.Contains(t, query, "UPDATE users SET")
	assert.Contains(t, query, "name = $1")
	assert.Contains(t, query, "WHERE id = $2")
	assert.Len(t, args, 2)
}

func TestQueryBuilder_PostgreSQL_Delete(t *testing.T) {
	qb := NewQueryBuilderWithDriver(DriverPostgreSQL)
	query, args := qb.
		Table("users").
		Delete().
		Where("id = ?", 1).
		Build()

	expectedQuery := "DELETE FROM users WHERE id = $1"
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, []interface{}{1}, args)
}

func TestQueryBuilder_PostgreSQL_ComplexQuery(t *testing.T) {
	qb := NewQueryBuilderWithDriver(DriverPostgreSQL)
	query, args := qb.
		Table("users").
		Select("users.*, roles.name as role_name").
		Join("roles", "users.role_id = roles.id").
		Where("users.status = ?", "active").
		Where("users.age > ?", 18).
		OrderBy("created_at", "DESC").
		Limit(10).
		Offset(20).
		Build()

	assert.Contains(t, query, "SELECT users.*, roles.name as role_name FROM users")
	assert.Contains(t, query, "INNER JOIN roles ON users.role_id = roles.id")
	assert.Contains(t, query, "WHERE users.status = $1 AND users.age > $2")
	assert.Contains(t, query, "ORDER BY created_at DESC")
	assert.Contains(t, query, "LIMIT 10 OFFSET 20")
	assert.Equal(t, []interface{}{"active", 18}, args)
}

func TestBulkInsertBuilder_PostgreSQL(t *testing.T) {
	bi := NewBulkInsertBuilderWithDriver("users", DriverPostgreSQL)
	bi.Columns("name", "email")

	bi.Values("User 1", "user1@example.com")
	bi.Values("User 2", "user2@example.com")

	query, args := bi.Build()

	expectedQuery := "INSERT INTO users (name, email) VALUES ($1, $2), ($3, $4)"
	assert.Equal(t, expectedQuery, query)
	assert.Len(t, args, 4)
}

func TestUpsertBuilder_MySQL(t *testing.T) {
	ub := NewUpsertBuilderWithDriver("users", DriverMySQL)
	ub.Insert(map[string]interface{}{
		"id":    1,
		"name":  "John",
		"email": "john@example.com",
	})
	ub.Update(map[string]interface{}{
		"name": "John Updated",
	})

	query, args := ub.Build()

	assert.Contains(t, query, "INSERT INTO users")
	assert.Contains(t, query, "VALUES (?, ?, ?)")
	assert.Contains(t, query, "ON DUPLICATE KEY UPDATE")
	assert.Contains(t, query, "name = ?")
	assert.Len(t, args, 4)
}

func TestUpsertBuilder_PostgreSQL(t *testing.T) {
	ub := NewUpsertBuilderWithDriver("users", DriverPostgreSQL)
	ub.OnConflict("email")
	ub.Insert(map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
	})
	ub.Update(map[string]interface{}{
		"name": "John Updated",
	})

	query, args := ub.Build()

	assert.Contains(t, query, "INSERT INTO users")
	assert.Contains(t, query, "VALUES ($1, $2)")
	assert.Contains(t, query, "ON CONFLICT (email) DO UPDATE SET")
	assert.Contains(t, query, "name = $3")
	assert.Len(t, args, 3)
}

func TestSetDefaultDriver(t *testing.T) {
	// Test changing default driver
	SetDefaultDriver(DriverPostgreSQL)

	qb := NewQueryBuilder()
	query, args := qb.
		Table("users").
		Select("*").
		Where("id = ?", 1).
		Build()

	assert.Contains(t, query, "$1")
	assert.Equal(t, []interface{}{1}, args)

	// Reset to MySQL
	SetDefaultDriver(DriverMySQL)

	qb2 := NewQueryBuilder()
	query2, args2 := qb2.
		Table("users").
		Select("*").
		Where("id = ?", 1).
		Build()

	assert.Contains(t, query2, "?")
	assert.Equal(t, []interface{}{1}, args2)
}

func TestQueryBuilder_SwitchDriver(t *testing.T) {
	qb := NewQueryBuilder()

	// Start with MySQL (default)
	query1, _ := qb.
		Table("users").
		Select("*").
		Where("id = ?", 1).
		Build()

	assert.Contains(t, query1, "?")

	// Switch to PostgreSQL
	qb.SetDriver(DriverPostgreSQL)
	query2, _ := qb.
		Table("users").
		Select("*").
		Where("id = ?", 1).
		Build()

	assert.Contains(t, query2, "$1")
}

func TestModel_WithDriver(t *testing.T) {
	// Create model with PostgreSQL driver
	model := NewModelWithDriver(nil, &TestUser{}, DriverPostgreSQL)

	model.Table("users").Where("id = ?", 1)
	query, args := model.ToSQL()

	assert.Contains(t, query, "$1")
	assert.Equal(t, []interface{}{1}, args)
}

func TestQueryBuilder_PostgreSQL_GroupByHaving(t *testing.T) {
	qb := NewQueryBuilderWithDriver(DriverPostgreSQL)
	query, args := qb.
		Table("users").
		Select("status", "COUNT(*) as count").
		GroupBy("status").
		Having("COUNT(*) > ?", 10).
		Build()

	expectedQuery := "SELECT status, COUNT(*) as count FROM users GROUP BY status HAVING COUNT(*) > $1"
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, []interface{}{10}, args)
}

func TestQueryBuilder_PostgreSQL_MultiplePlaceholders(t *testing.T) {
	qb := NewQueryBuilderWithDriver(DriverPostgreSQL)
	query, args := qb.
		Table("users").
		Select("*").
		Where("age > ?", 18).
		Where("status = ?", "active").
		Where("created_at > ?", "2024-01-01").
		WhereBetween("score", 50, 100).
		Build()

	// Should have $1, $2, $3, $4, $5
	assert.Contains(t, query, "$1")
	assert.Contains(t, query, "$2")
	assert.Contains(t, query, "$3")
	assert.Contains(t, query, "$4")
	assert.Contains(t, query, "$5")
	assert.Len(t, args, 5)
}
