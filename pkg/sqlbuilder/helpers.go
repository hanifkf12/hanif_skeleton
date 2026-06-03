package sqlbuilder

import (
	"fmt"
	"strings"
)

// Raw creates a raw query builder
func Raw(query string, args ...interface{}) *RawQuery {
	return &RawQuery{
		query: query,
		args:  args,
	}
}

// RawQuery represents a raw SQL query
type RawQuery struct {
	query string
	args  []interface{}
}

// Build returns the query and args
func (r *RawQuery) Build() (string, []interface{}) {
	return r.query, r.args
}

// Clause represents a reusable query clause
type Clause struct {
	condition string
	args      []interface{}
}

// NewClause creates a new clause
func NewClause(condition string, args ...interface{}) *Clause {
	return &Clause{
		condition: condition,
		args:      args,
	}
}

// GetCondition returns the condition string
func (c *Clause) GetCondition() string {
	return c.condition
}

// GetArgs returns the arguments
func (c *Clause) GetArgs() []interface{} {
	return c.args
}

// Expression represents a SQL expression
type Expression struct {
	expr string
}

// NewExpression creates a new SQL expression
func NewExpression(expr string) *Expression {
	return &Expression{expr: expr}
}

// String returns the expression as string
func (e *Expression) String() string {
	return e.expr
}

// Common expressions
var (
	ExprNow       = NewExpression("NOW()")
	ExprCurrentTS = NewExpression("CURRENT_TIMESTAMP")
	ExprNull      = NewExpression("NULL")
)

// ConditionalBuilder helps build complex WHERE conditions
type ConditionalBuilder struct {
	conditions []string
	args       []interface{}
	operator   string
}

// NewConditionalBuilder creates a new conditional builder
func NewConditionalBuilder(operator ...string) *ConditionalBuilder {
	op := "AND"
	if len(operator) > 0 {
		op = strings.ToUpper(operator[0])
	}
	return &ConditionalBuilder{
		conditions: []string{},
		args:       []interface{}{},
		operator:   op,
	}
}

// Add adds a condition
func (cb *ConditionalBuilder) Add(condition string, args ...interface{}) *ConditionalBuilder {
	cb.conditions = append(cb.conditions, condition)
	cb.args = append(cb.args, args...)
	return cb
}

// AddIf adds a condition only if the condition is true
func (cb *ConditionalBuilder) AddIf(cond bool, condition string, args ...interface{}) *ConditionalBuilder {
	if cond {
		cb.Add(condition, args...)
	}
	return cb
}

// Build builds the conditional statement
func (cb *ConditionalBuilder) Build() (string, []interface{}) {
	if len(cb.conditions) == 0 {
		return "", nil
	}
	return strings.Join(cb.conditions, fmt.Sprintf(" %s ", cb.operator)), cb.args
}

// IsEmpty checks if the builder has no conditions
func (cb *ConditionalBuilder) IsEmpty() bool {
	return len(cb.conditions) == 0
}

// BulkInsertBuilder helps build bulk insert queries
type BulkInsertBuilder struct {
	table      string
	columns    []string
	values     [][]interface{}
	driverType DriverType
}

// NewBulkInsertBuilder creates a new bulk insert builder with default driver
func NewBulkInsertBuilder(table string) *BulkInsertBuilder {
	return &BulkInsertBuilder{
		table:      table,
		columns:    []string{},
		values:     [][]interface{}{},
		driverType: defaultDriverType,
	}
}

// NewBulkInsertBuilderWithDriver creates a new bulk insert builder with specific driver
func NewBulkInsertBuilderWithDriver(table string, driver DriverType) *BulkInsertBuilder {
	bi := NewBulkInsertBuilder(table)
	bi.driverType = driver
	return bi
}

// SetDriver sets the driver type
func (bi *BulkInsertBuilder) SetDriver(driver DriverType) *BulkInsertBuilder {
	bi.driverType = driver
	return bi
}

// Columns sets the columns
func (bi *BulkInsertBuilder) Columns(columns ...string) *BulkInsertBuilder {
	bi.columns = columns
	return bi
}

// Values adds a row of values
func (bi *BulkInsertBuilder) Values(values ...interface{}) *BulkInsertBuilder {
	bi.values = append(bi.values, values)
	return bi
}

// AddRow adds a row from a map
func (bi *BulkInsertBuilder) AddRow(data map[string]interface{}) *BulkInsertBuilder {
	if len(bi.columns) == 0 {
		// Extract columns from first row
		for col := range data {
			bi.columns = append(bi.columns, col)
		}
	}

	row := make([]interface{}, len(bi.columns))
	for i, col := range bi.columns {
		row[i] = data[col]
	}
	bi.values = append(bi.values, row)
	return bi
}

// AddFromStruct adds a row from a struct
func (bi *BulkInsertBuilder) AddFromStruct(model interface{}) *BulkInsertBuilder {
	data := StructToMap(model, true)
	return bi.AddRow(data)
}

// Build builds the bulk insert query
func (bi *BulkInsertBuilder) Build() (string, []interface{}) {
	if len(bi.columns) == 0 || len(bi.values) == 0 {
		return "", nil
	}

	var query strings.Builder
	args := []interface{}{}

	query.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES ",
		bi.table,
		strings.Join(bi.columns, ", ")))

	valuePlaceholders := []string{}
	for _, row := range bi.values {
		placeholders := make([]string, len(row))
		for i := range row {
			placeholders[i] = "?"
		}
		valuePlaceholders = append(valuePlaceholders, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
		args = append(args, row...)
	}

	query.WriteString(strings.Join(valuePlaceholders, ", "))

	// Convert placeholders based on driver type
	finalQuery := bi.convertPlaceholders(query.String())

	return finalQuery, args
}

// convertPlaceholders converts placeholders based on driver type
func (bi *BulkInsertBuilder) convertPlaceholders(query string) string {
	if bi.driverType == DriverMySQL {
		return query
	}

	// PostgreSQL: convert ? to $1, $2, $3, ...
	placeholderIndex := 1
	result := strings.Builder{}

	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			result.WriteString(fmt.Sprintf("$%d", placeholderIndex))
			placeholderIndex++
		} else {
			result.WriteByte(query[i])
		}
	}

	return result.String()
}

// UpsertBuilder helps build INSERT ... ON DUPLICATE KEY UPDATE queries (MySQL)
// or INSERT ... ON CONFLICT ... DO UPDATE (PostgreSQL)
type UpsertBuilder struct {
	table          string
	insertData     map[string]interface{}
	updateData     map[string]interface{}
	driverType     DriverType
	conflictColumn string // For PostgreSQL ON CONFLICT
}

// NewUpsertBuilder creates a new upsert builder with default driver
func NewUpsertBuilder(table string) *UpsertBuilder {
	return &UpsertBuilder{
		table:      table,
		insertData: make(map[string]interface{}),
		updateData: make(map[string]interface{}),
		driverType: defaultDriverType,
	}
}

// NewUpsertBuilderWithDriver creates a new upsert builder with specific driver
func NewUpsertBuilderWithDriver(table string, driver DriverType) *UpsertBuilder {
	ub := NewUpsertBuilder(table)
	ub.driverType = driver
	return ub
}

// SetDriver sets the driver type
func (ub *UpsertBuilder) SetDriver(driver DriverType) *UpsertBuilder {
	ub.driverType = driver
	return ub
}

// OnConflict sets the conflict column for PostgreSQL (e.g., "id" or "(email)")
func (ub *UpsertBuilder) OnConflict(column string) *UpsertBuilder {
	ub.conflictColumn = column
	return ub
}

// Insert sets the data to insert
func (ub *UpsertBuilder) Insert(data map[string]interface{}) *UpsertBuilder {
	ub.insertData = data
	return ub
}

// Update sets the data to update on duplicate
func (ub *UpsertBuilder) Update(data map[string]interface{}) *UpsertBuilder {
	ub.updateData = data
	return ub
}

// Build builds the upsert query (MySQL or PostgreSQL syntax)
func (ub *UpsertBuilder) Build() (string, []interface{}) {
	if ub.driverType == DriverPostgreSQL {
		return ub.buildPostgreSQL()
	}
	return ub.buildMySQL()
}

// buildMySQL builds MySQL upsert query (ON DUPLICATE KEY UPDATE)
func (ub *UpsertBuilder) buildMySQL() (string, []interface{}) {
	var query strings.Builder
	args := []interface{}{}

	// Build INSERT part
	columns := []string{}
	placeholders := []string{}

	for col, val := range ub.insertData {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		args = append(args, val)
	}

	query.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		ub.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", ")))

	// Build ON DUPLICATE KEY UPDATE part
	if len(ub.updateData) > 0 {
		query.WriteString(" ON DUPLICATE KEY UPDATE ")

		updateClauses := []string{}
		for col, val := range ub.updateData {
			updateClauses = append(updateClauses, fmt.Sprintf("%s = ?", col))
			args = append(args, val)
		}

		query.WriteString(strings.Join(updateClauses, ", "))
	}

	return query.String(), args
}

// buildPostgreSQL builds PostgreSQL upsert query (ON CONFLICT ... DO UPDATE)
func (ub *UpsertBuilder) buildPostgreSQL() (string, []interface{}) {
	var query strings.Builder
	args := []interface{}{}

	// Build INSERT part
	columns := []string{}
	placeholders := []string{}
	placeholderIndex := 1

	for col, val := range ub.insertData {
		columns = append(columns, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", placeholderIndex))
		args = append(args, val)
		placeholderIndex++
	}

	query.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		ub.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", ")))

	// Build ON CONFLICT DO UPDATE part
	if len(ub.updateData) > 0 {
		conflictCol := ub.conflictColumn
		if conflictCol == "" {
			conflictCol = "id" // default conflict column
		}

		query.WriteString(fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET ", conflictCol))

		updateClauses := []string{}
		for col, val := range ub.updateData {
			updateClauses = append(updateClauses, fmt.Sprintf("%s = $%d", col, placeholderIndex))
			args = append(args, val)
			placeholderIndex++
		}

		query.WriteString(strings.Join(updateClauses, ", "))
	}

	return query.String(), args
}

// CaseBuilder helps build CASE WHEN expressions
type CaseBuilder struct {
	cases     []caseWhen
	elseValue interface{}
	hasElse   bool
}

type caseWhen struct {
	condition string
	value     interface{}
}

// NewCaseBuilder creates a new CASE builder
func NewCaseBuilder() *CaseBuilder {
	return &CaseBuilder{
		cases:   []caseWhen{},
		hasElse: false,
	}
}

// When adds a WHEN clause
func (cb *CaseBuilder) When(condition string, value interface{}) *CaseBuilder {
	cb.cases = append(cb.cases, caseWhen{
		condition: condition,
		value:     value,
	})
	return cb
}

// Else sets the ELSE clause
func (cb *CaseBuilder) Else(value interface{}) *CaseBuilder {
	cb.elseValue = value
	cb.hasElse = true
	return cb
}

// Build builds the CASE expression
func (cb *CaseBuilder) Build() string {
	var query strings.Builder

	query.WriteString("CASE")

	for _, c := range cb.cases {
		query.WriteString(fmt.Sprintf(" WHEN %s THEN %v", c.condition, c.value))
	}

	if cb.hasElse {
		query.WriteString(fmt.Sprintf(" ELSE %v", cb.elseValue))
	}

	query.WriteString(" END")

	return query.String()
}
