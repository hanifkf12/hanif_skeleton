package sqlbuilder

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hanifkf12/hanif_skeleton/pkg/databasex"
)

// Model is a helper struct that combines QueryBuilder with database operations
type Model struct {
	db      databasex.Database
	builder *QueryBuilder
	model   interface{}
}

// NewModel creates a new Model instance
func NewModel(db databasex.Database, model interface{}) *Model {
	return &Model{
		db:      db,
		builder: NewQueryBuilder(),
		model:   model,
	}
}

// Table sets the table name
func (m *Model) Table(table string) *Model {
	m.builder.Table(table)
	return m
}

// Select sets the columns to select
func (m *Model) Select(columns ...string) *Model {
	m.builder.Select(columns...)
	return m
}

// Where adds a WHERE clause
func (m *Model) Where(condition string, args ...interface{}) *Model {
	m.builder.Where(condition, args...)
	return m
}

// OrWhere adds an OR WHERE clause
func (m *Model) OrWhere(condition string, args ...interface{}) *Model {
	m.builder.OrWhere(condition, args...)
	return m
}

// WhereIn adds a WHERE IN clause
func (m *Model) WhereIn(column string, values []interface{}) *Model {
	m.builder.WhereIn(column, values)
	return m
}

// WhereNotIn adds a WHERE NOT IN clause
func (m *Model) WhereNotIn(column string, values []interface{}) *Model {
	m.builder.WhereNotIn(column, values)
	return m
}

// WhereBetween adds a WHERE BETWEEN clause
func (m *Model) WhereBetween(column string, start, end interface{}) *Model {
	m.builder.WhereBetween(column, start, end)
	return m
}

// WhereNull adds a WHERE IS NULL clause
func (m *Model) WhereNull(column string) *Model {
	m.builder.WhereNull(column)
	return m
}

// WhereNotNull adds a WHERE IS NOT NULL clause
func (m *Model) WhereNotNull(column string) *Model {
	m.builder.WhereNotNull(column)
	return m
}

// OrderBy adds an ORDER BY clause
func (m *Model) OrderBy(column string, direction ...string) *Model {
	m.builder.OrderBy(column, direction...)
	return m
}

// GroupBy adds a GROUP BY clause
func (m *Model) GroupBy(columns ...string) *Model {
	m.builder.GroupBy(columns...)
	return m
}

// Having adds a HAVING clause
func (m *Model) Having(condition string, args ...interface{}) *Model {
	m.builder.Having(condition, args...)
	return m
}

// Limit sets the LIMIT
func (m *Model) Limit(limit int) *Model {
	m.builder.Limit(limit)
	return m
}

// Offset sets the OFFSET
func (m *Model) Offset(offset int) *Model {
	m.builder.Offset(offset)
	return m
}

// Join adds a JOIN clause
func (m *Model) Join(table, condition string) *Model {
	m.builder.Join(table, condition)
	return m
}

// LeftJoin adds a LEFT JOIN clause
func (m *Model) LeftJoin(table, condition string) *Model {
	m.builder.LeftJoin(table, condition)
	return m
}

// RightJoin adds a RIGHT JOIN clause
func (m *Model) RightJoin(table, condition string) *Model {
	m.builder.RightJoin(table, condition)
	return m
}

// Get executes the query and returns a single result
func (m *Model) Get(ctx context.Context, dest interface{}) error {
	query, args := m.builder.Build()
	return m.db.Get(ctx, dest, query, args...)
}

// GetAll executes the query and returns all results
func (m *Model) GetAll(ctx context.Context, dest interface{}) error {
	query, args := m.builder.Build()
	return m.db.Select(ctx, dest, query, args...)
}

// First executes the query and returns the first result
func (m *Model) First(ctx context.Context, dest interface{}) error {
	m.builder.Limit(1)
	query, args := m.builder.Build()
	return m.db.Get(ctx, dest, query, args...)
}

// Exec executes the query (for INSERT, UPDATE, DELETE)
func (m *Model) Exec(ctx context.Context) (sql.Result, error) {
	query, args := m.builder.Build()
	return m.db.Exec(ctx, query, args...)
}

// Count returns the count of rows
func (m *Model) Count(ctx context.Context) (int64, error) {
	// Store original columns
	originalCols := m.builder.columns

	// Set count query
	m.builder.Select("COUNT(*) as count")

	query, args := m.builder.Build()

	// Restore original columns
	m.builder.columns = originalCols

	var count int64
	err := m.db.Get(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Exists checks if any rows exist
func (m *Model) Exists(ctx context.Context) (bool, error) {
	count, err := m.Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Insert inserts a record using struct
func (m *Model) Insert(ctx context.Context, model interface{}) (sql.Result, error) {
	data := StructToMapExclude(model, "id", "created_at", "updated_at")
	m.builder.Insert(data)
	return m.Exec(ctx)
}

// InsertWithFields inserts a record with specific fields
func (m *Model) InsertWithFields(ctx context.Context, model interface{}, fields ...string) (sql.Result, error) {
	data := StructToMapInclude(model, fields...)
	m.builder.Insert(data)
	return m.Exec(ctx)
}

// Update updates records using struct
func (m *Model) Update(ctx context.Context, model interface{}) (sql.Result, error) {
	data := StructToMapExclude(model, "id", "created_at", "updated_at")
	m.builder.Update(data)
	return m.Exec(ctx)
}

// UpdateWithFields updates records with specific fields
func (m *Model) UpdateWithFields(ctx context.Context, model interface{}, fields ...string) (sql.Result, error) {
	data := StructToMapInclude(model, fields...)
	m.builder.Update(data)
	return m.Exec(ctx)
}

// Delete deletes records
func (m *Model) Delete(ctx context.Context) (sql.Result, error) {
	m.builder.Delete()
	return m.Exec(ctx)
}

// ToSQL returns the SQL query and args without executing
func (m *Model) ToSQL() (string, []interface{}) {
	return m.builder.Build()
}

// FindByID is a helper to find a record by ID
func FindByID(ctx context.Context, db databasex.Database, table string, id interface{}, dest interface{}) error {
	model := NewModel(db, dest)
	return model.Table(table).Where("id = ?", id).First(ctx, dest)
}

// FindAll is a helper to find all records
func FindAll(ctx context.Context, db databasex.Database, table string, dest interface{}) error {
	model := NewModel(db, dest)
	return model.Table(table).Select("*").GetAll(ctx, dest)
}

// DeleteByID is a helper to delete a record by ID
func DeleteByID(ctx context.Context, db databasex.Database, table string, id interface{}) (sql.Result, error) {
	model := NewModel(db, nil)
	return model.Table(table).Where("id = ?", id).Delete(ctx)
}

// CreateRecord is a helper to create a record from struct
func CreateRecord(ctx context.Context, db databasex.Database, table string, model interface{}) (sql.Result, error) {
	m := NewModel(db, model)
	return m.Table(table).Insert(ctx, model)
}

// UpdateRecord is a helper to update a record from struct
func UpdateRecord(ctx context.Context, db databasex.Database, table string, id interface{}, model interface{}) (sql.Result, error) {
	m := NewModel(db, model)
	return m.Table(table).Where("id = ?", id).Update(ctx, model)
}

// Paginate adds pagination to query
func (m *Model) Paginate(page, perPage int) *Model {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	offset := (page - 1) * perPage
	m.builder.Limit(perPage).Offset(offset)
	return m
}

// PaginationResult holds pagination data
type PaginationResult struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

// GetWithPagination executes query with pagination info
func (m *Model) GetWithPagination(ctx context.Context, dest interface{}, page, perPage int) (*PaginationResult, error) {
	// Get total count first
	total, err := m.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get count: %w", err)
	}

	// Apply pagination
	m.Paginate(page, perPage)

	// Get data
	err = m.GetAll(ctx, dest)
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	return &PaginationResult{
		Data:       dest,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}
