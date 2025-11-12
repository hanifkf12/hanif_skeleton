package sqlbuilder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestUser struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Age       int       `db:"age"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (u TestUser) TableName() string {
	return "users"
}

func TestQueryBuilder_Select(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("users").
		Select("id", "name", "email").
		Where("status = ?", "active").
		OrderBy("created_at", "DESC").
		Limit(10).
		Build()

	expectedQuery := "SELECT id, name, email FROM users WHERE status = ? ORDER BY created_at DESC LIMIT 10"
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, []interface{}{"active"}, args)
}

func TestQueryBuilder_SelectAll(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("users").
		Select("*").
		Build()

	expectedQuery := "SELECT * FROM users"
	assert.Equal(t, expectedQuery, query)
	assert.Empty(t, args)
}

func TestQueryBuilder_WhereMultiple(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("users").
		Select("*").
		Where("age > ?", 18).
		Where("status = ?", "active").
		Build()

	expectedQuery := "SELECT * FROM users WHERE age > ? AND status = ?"
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, []interface{}{18, "active"}, args)
}

func TestQueryBuilder_OrWhere(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("users").
		Select("*").
		Where("age > ?", 18).
		OrWhere("is_premium = ?", true).
		Build()

	expectedQuery := "SELECT * FROM users WHERE age > ? OR is_premium = ?"
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, []interface{}{18, true}, args)
}

func TestQueryBuilder_WhereIn(t *testing.T) {
	qb := NewQueryBuilder()
	values := []interface{}{1, 2, 3, 4, 5}
	query, args := qb.
		Table("users").
		Select("*").
		WhereIn("id", values).
		Build()

	expectedQuery := "SELECT * FROM users WHERE id IN (?, ?, ?, ?, ?)"
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, values, args)
}

func TestQueryBuilder_WhereBetween(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("users").
		Select("*").
		WhereBetween("age", 18, 30).
		Build()

	expectedQuery := "SELECT * FROM users WHERE age BETWEEN ? AND ?"
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, []interface{}{18, 30}, args)
}

func TestQueryBuilder_WhereNull(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("users").
		Select("*").
		WhereNull("deleted_at").
		Build()

	expectedQuery := "SELECT * FROM users WHERE deleted_at IS NULL"
	assert.Equal(t, expectedQuery, query)
	assert.Empty(t, args)
}

func TestQueryBuilder_Join(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("users").
		Select("users.*, roles.name as role_name").
		Join("roles", "users.role_id = roles.id").
		Where("users.status = ?", "active").
		Build()

	expectedQuery := "SELECT users.*, roles.name as role_name FROM users INNER JOIN roles ON users.role_id = roles.id WHERE users.status = ?"
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, []interface{}{"active"}, args)
}

func TestQueryBuilder_LeftJoin(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("users").
		Select("*").
		LeftJoin("profiles", "users.id = profiles.user_id").
		Build()

	expectedQuery := "SELECT * FROM users LEFT JOIN profiles ON users.id = profiles.user_id"
	assert.Equal(t, expectedQuery, query)
	assert.Empty(t, args)
}

func TestQueryBuilder_GroupByHaving(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("users").
		Select("status", "COUNT(*) as count").
		GroupBy("status").
		Having("COUNT(*) > ?", 10).
		Build()

	expectedQuery := "SELECT status, COUNT(*) as count FROM users GROUP BY status HAVING COUNT(*) > ?"
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, []interface{}{10}, args)
}

func TestQueryBuilder_LimitOffset(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("users").
		Select("*").
		Limit(20).
		Offset(40).
		Build()

	expectedQuery := "SELECT * FROM users LIMIT 20 OFFSET 40"
	assert.Equal(t, expectedQuery, query)
	assert.Empty(t, args)
}

func TestQueryBuilder_Insert(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("users").
		Insert(map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   25,
		}).
		Build()

	assert.Contains(t, query, "INSERT INTO users")
	assert.Contains(t, query, "name")
	assert.Contains(t, query, "email")
	assert.Contains(t, query, "age")
	assert.Len(t, args, 3)
}

func TestQueryBuilder_Update(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("users").
		Update(map[string]interface{}{
			"name":  "John Updated",
			"email": "john.updated@example.com",
		}).
		Where("id = ?", 1).
		Build()

	assert.Contains(t, query, "UPDATE users SET")
	assert.Contains(t, query, "name = ?")
	assert.Contains(t, query, "email = ?")
	assert.Contains(t, query, "WHERE id = ?")
	assert.Len(t, args, 3)
}

func TestQueryBuilder_Delete(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.
		Table("users").
		Delete().
		Where("id = ?", 1).
		Build()

	expectedQuery := "DELETE FROM users WHERE id = ?"
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, []interface{}{1}, args)
}

func TestStructToMap(t *testing.T) {
	user := TestUser{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}

	data := StructToMap(&user, false)

	assert.Equal(t, 1, data["id"])
	assert.Equal(t, "John Doe", data["name"])
	assert.Equal(t, "john@example.com", data["email"])
	assert.Equal(t, 25, data["age"])
}

func TestStructToMap_SkipZero(t *testing.T) {
	user := TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	data := StructToMap(&user, true)

	assert.NotContains(t, data, "id")
	assert.NotContains(t, data, "age")
	assert.Contains(t, data, "name")
	assert.Contains(t, data, "email")
}

func TestStructToMapExclude(t *testing.T) {
	user := TestUser{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	data := StructToMapExclude(&user, "id", "created_at", "updated_at")

	assert.NotContains(t, data, "id")
	assert.NotContains(t, data, "created_at")
	assert.NotContains(t, data, "updated_at")
	assert.Contains(t, data, "name")
	assert.Contains(t, data, "email")
}

func TestStructToMapInclude(t *testing.T) {
	user := TestUser{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}

	data := StructToMapInclude(&user, "name", "email")

	assert.Equal(t, 2, len(data))
	assert.Contains(t, data, "name")
	assert.Contains(t, data, "email")
	assert.NotContains(t, data, "id")
	assert.NotContains(t, data, "age")
}

func TestGetColumns(t *testing.T) {
	user := TestUser{}
	columns := GetColumns(&user)

	assert.Contains(t, columns, "id")
	assert.Contains(t, columns, "name")
	assert.Contains(t, columns, "email")
	assert.Contains(t, columns, "age")
}

func TestGetColumnsExclude(t *testing.T) {
	user := TestUser{}
	columns := GetColumnsExclude(&user, "created_at", "updated_at")

	assert.Contains(t, columns, "id")
	assert.Contains(t, columns, "name")
	assert.Contains(t, columns, "email")
	assert.NotContains(t, columns, "created_at")
	assert.NotContains(t, columns, "updated_at")
}

func TestBuildSelectColumns(t *testing.T) {
	user := TestUser{}
	cols := BuildSelectColumns(&user, "u")

	assert.Contains(t, cols, "u.id")
	assert.Contains(t, cols, "u.name")
	assert.Contains(t, cols, "u.email")
}

func TestConditionalBuilder(t *testing.T) {
	cb := NewConditionalBuilder()
	cb.Add("age > ?", 18)
	cb.Add("status = ?", "active")

	condition, args := cb.Build()

	assert.Equal(t, "age > ? AND status = ?", condition)
	assert.Equal(t, []interface{}{18, "active"}, args)
}

func TestConditionalBuilder_AddIf(t *testing.T) {
	name := "John"
	age := 0
	status := "active"

	cb := NewConditionalBuilder()
	cb.AddIf(name != "", "name = ?", name)
	cb.AddIf(age > 0, "age = ?", age)
	cb.AddIf(status != "", "status = ?", status)

	condition, args := cb.Build()

	assert.Equal(t, "name = ? AND status = ?", condition)
	assert.Equal(t, []interface{}{"John", "active"}, args)
	assert.Len(t, args, 2)
}

func TestBulkInsertBuilder(t *testing.T) {
	bi := NewBulkInsertBuilder("users")
	bi.Columns("name", "email", "age")

	bi.Values("User 1", "user1@example.com", 25)
	bi.Values("User 2", "user2@example.com", 30)
	bi.Values("User 3", "user3@example.com", 28)

	query, args := bi.Build()

	expectedQuery := "INSERT INTO users (name, email, age) VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?)"
	assert.Equal(t, expectedQuery, query)
	assert.Len(t, args, 9)
}

func TestBulkInsertBuilder_AddRow(t *testing.T) {
	bi := NewBulkInsertBuilder("users")

	bi.AddRow(map[string]interface{}{
		"name":  "User 1",
		"email": "user1@example.com",
	})

	bi.AddRow(map[string]interface{}{
		"name":  "User 2",
		"email": "user2@example.com",
	})

	query, args := bi.Build()

	assert.Contains(t, query, "INSERT INTO users")
	assert.Len(t, args, 4)
}

func TestUpsertBuilder(t *testing.T) {
	ub := NewUpsertBuilder("users")
	ub.Insert(map[string]interface{}{
		"id":    1,
		"name":  "John",
		"email": "john@example.com",
	})
	ub.Update(map[string]interface{}{
		"name":  "John Updated",
		"email": "john.updated@example.com",
	})

	query, args := ub.Build()

	assert.Contains(t, query, "INSERT INTO users")
	assert.Contains(t, query, "ON DUPLICATE KEY UPDATE")
	assert.Len(t, args, 5)
}

func TestCaseBuilder(t *testing.T) {
	cb := NewCaseBuilder()
	result := cb.
		When("age < 18", "'minor'").
		When("age >= 18 AND age < 65", "'adult'").
		Else("'senior'").
		Build()

	expected := "CASE WHEN age < 18 THEN 'minor' WHEN age >= 18 AND age < 65 THEN 'adult' ELSE 'senior' END"
	assert.Equal(t, expected, result)
}

func TestGetTableName(t *testing.T) {
	user := TestUser{}
	tableName := GetTableName(&user)

	assert.Equal(t, "users", tableName)
}

func TestRawQuery(t *testing.T) {
	raw := Raw("SELECT * FROM users WHERE id = ?", 1)
	query, args := raw.Build()

	assert.Equal(t, "SELECT * FROM users WHERE id = ?", query)
	assert.Equal(t, []interface{}{1}, args)
}
