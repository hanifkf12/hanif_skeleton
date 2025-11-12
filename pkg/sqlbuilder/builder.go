package sqlbuilder

import (
	"fmt"
	"reflect"
	"strings"
)

// QueryBuilder is the main builder for SQL queries
type QueryBuilder struct {
	table      string
	columns    []string
	where      []whereClause
	orderBy    []string
	limit      int
	offset     int
	joins      []joinClause
	groupBy    []string
	having     []whereClause
	args       []interface{}
	queryType  QueryType
	updateData map[string]interface{}
	insertData map[string]interface{}
}

type QueryType int

const (
	QueryTypeSelect QueryType = iota
	QueryTypeInsert
	QueryTypeUpdate
	QueryTypeDelete
)

type whereClause struct {
	condition string
	args      []interface{}
	operator  string // AND, OR
}

type joinClause struct {
	joinType  string // INNER, LEFT, RIGHT, FULL
	table     string
	condition string
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		columns:    []string{},
		where:      []whereClause{},
		orderBy:    []string{},
		joins:      []joinClause{},
		groupBy:    []string{},
		having:     []whereClause{},
		args:       []interface{}{},
		updateData: make(map[string]interface{}),
		insertData: make(map[string]interface{}),
	}
}

// Table sets the table name
func (qb *QueryBuilder) Table(table string) *QueryBuilder {
	qb.table = table
	return qb
}

// Select sets the columns to select
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.queryType = QueryTypeSelect
	qb.columns = columns
	return qb
}

// Where adds a WHERE clause
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.where = append(qb.where, whereClause{
		condition: condition,
		args:      args,
		operator:  "AND",
	})
	return qb
}

// OrWhere adds an OR WHERE clause
func (qb *QueryBuilder) OrWhere(condition string, args ...interface{}) *QueryBuilder {
	qb.where = append(qb.where, whereClause{
		condition: condition,
		args:      args,
		operator:  "OR",
	})
	return qb
}

// WhereIn adds a WHERE IN clause
func (qb *QueryBuilder) WhereIn(column string, values []interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
	}
	condition := fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", "))
	return qb.Where(condition, values...)
}

// WhereNotIn adds a WHERE NOT IN clause
func (qb *QueryBuilder) WhereNotIn(column string, values []interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
	}
	condition := fmt.Sprintf("%s NOT IN (%s)", column, strings.Join(placeholders, ", "))
	return qb.Where(condition, values...)
}

// WhereBetween adds a WHERE BETWEEN clause
func (qb *QueryBuilder) WhereBetween(column string, start, end interface{}) *QueryBuilder {
	condition := fmt.Sprintf("%s BETWEEN ? AND ?", column)
	return qb.Where(condition, start, end)
}

// WhereNull adds a WHERE IS NULL clause
func (qb *QueryBuilder) WhereNull(column string) *QueryBuilder {
	condition := fmt.Sprintf("%s IS NULL", column)
	return qb.Where(condition)
}

// WhereNotNull adds a WHERE IS NOT NULL clause
func (qb *QueryBuilder) WhereNotNull(column string) *QueryBuilder {
	condition := fmt.Sprintf("%s IS NOT NULL", column)
	return qb.Where(condition)
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder) OrderBy(column string, direction ...string) *QueryBuilder {
	dir := "ASC"
	if len(direction) > 0 {
		dir = strings.ToUpper(direction[0])
	}
	qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", column, dir))
	return qb
}

// GroupBy adds a GROUP BY clause
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groupBy = append(qb.groupBy, columns...)
	return qb
}

// Having adds a HAVING clause
func (qb *QueryBuilder) Having(condition string, args ...interface{}) *QueryBuilder {
	qb.having = append(qb.having, whereClause{
		condition: condition,
		args:      args,
		operator:  "AND",
	})
	return qb
}

// Limit sets the LIMIT
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset sets the OFFSET
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Join adds a JOIN clause
func (qb *QueryBuilder) Join(table, condition string) *QueryBuilder {
	qb.joins = append(qb.joins, joinClause{
		joinType:  "INNER",
		table:     table,
		condition: condition,
	})
	return qb
}

// LeftJoin adds a LEFT JOIN clause
func (qb *QueryBuilder) LeftJoin(table, condition string) *QueryBuilder {
	qb.joins = append(qb.joins, joinClause{
		joinType:  "LEFT",
		table:     table,
		condition: condition,
	})
	return qb
}

// RightJoin adds a RIGHT JOIN clause
func (qb *QueryBuilder) RightJoin(table, condition string) *QueryBuilder {
	qb.joins = append(qb.joins, joinClause{
		joinType:  "RIGHT",
		table:     table,
		condition: condition,
	})
	return qb
}

// Insert prepares an INSERT query
func (qb *QueryBuilder) Insert(data map[string]interface{}) *QueryBuilder {
	qb.queryType = QueryTypeInsert
	qb.insertData = data
	return qb
}

// Update prepares an UPDATE query
func (qb *QueryBuilder) Update(data map[string]interface{}) *QueryBuilder {
	qb.queryType = QueryTypeUpdate
	qb.updateData = data
	return qb
}

// Delete prepares a DELETE query
func (qb *QueryBuilder) Delete() *QueryBuilder {
	qb.queryType = QueryTypeDelete
	return qb
}

// Build builds the SQL query and returns query string and args
func (qb *QueryBuilder) Build() (string, []interface{}) {
	switch qb.queryType {
	case QueryTypeSelect:
		return qb.buildSelect()
	case QueryTypeInsert:
		return qb.buildInsert()
	case QueryTypeUpdate:
		return qb.buildUpdate()
	case QueryTypeDelete:
		return qb.buildDelete()
	default:
		return qb.buildSelect()
	}
}

func (qb *QueryBuilder) buildSelect() (string, []interface{}) {
	var query strings.Builder
	args := []interface{}{}

	// SELECT
	query.WriteString("SELECT ")
	if len(qb.columns) == 0 {
		query.WriteString("*")
	} else {
		query.WriteString(strings.Join(qb.columns, ", "))
	}

	// FROM
	query.WriteString(" FROM ")
	query.WriteString(qb.table)

	// JOINS
	for _, join := range qb.joins {
		query.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", join.joinType, join.table, join.condition))
	}

	// WHERE
	if len(qb.where) > 0 {
		query.WriteString(" WHERE ")
		for i, w := range qb.where {
			if i > 0 {
				query.WriteString(fmt.Sprintf(" %s ", w.operator))
			}
			query.WriteString(w.condition)
			args = append(args, w.args...)
		}
	}

	// GROUP BY
	if len(qb.groupBy) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(qb.groupBy, ", "))
	}

	// HAVING
	if len(qb.having) > 0 {
		query.WriteString(" HAVING ")
		for i, h := range qb.having {
			if i > 0 {
				query.WriteString(fmt.Sprintf(" %s ", h.operator))
			}
			query.WriteString(h.condition)
			args = append(args, h.args...)
		}
	}

	// ORDER BY
	if len(qb.orderBy) > 0 {
		query.WriteString(" ORDER BY ")
		query.WriteString(strings.Join(qb.orderBy, ", "))
	}

	// LIMIT
	if qb.limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}

	// OFFSET
	if qb.offset > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	return query.String(), args
}

func (qb *QueryBuilder) buildInsert() (string, []interface{}) {
	var query strings.Builder
	args := []interface{}{}

	query.WriteString("INSERT INTO ")
	query.WriteString(qb.table)

	columns := []string{}
	placeholders := []string{}

	for col, val := range qb.insertData {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		args = append(args, val)
	}

	query.WriteString(fmt.Sprintf(" (%s) VALUES (%s)",
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", ")))

	return query.String(), args
}

func (qb *QueryBuilder) buildUpdate() (string, []interface{}) {
	var query strings.Builder
	args := []interface{}{}

	query.WriteString("UPDATE ")
	query.WriteString(qb.table)
	query.WriteString(" SET ")

	setClauses := []string{}
	for col, val := range qb.updateData {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
		args = append(args, val)
	}

	query.WriteString(strings.Join(setClauses, ", "))

	// WHERE
	if len(qb.where) > 0 {
		query.WriteString(" WHERE ")
		for i, w := range qb.where {
			if i > 0 {
				query.WriteString(fmt.Sprintf(" %s ", w.operator))
			}
			query.WriteString(w.condition)
			args = append(args, w.args...)
		}
	}

	return query.String(), args
}

func (qb *QueryBuilder) buildDelete() (string, []interface{}) {
	var query strings.Builder
	args := []interface{}{}

	query.WriteString("DELETE FROM ")
	query.WriteString(qb.table)

	// WHERE
	if len(qb.where) > 0 {
		query.WriteString(" WHERE ")
		for i, w := range qb.where {
			if i > 0 {
				query.WriteString(fmt.Sprintf(" %s ", w.operator))
			}
			query.WriteString(w.condition)
			args = append(args, w.args...)
		}
	}

	return query.String(), args
}

// GetTableName extracts table name from struct tag
func GetTableName(model interface{}) string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check for TableName method
	if tn, ok := model.(TableNamer); ok {
		return tn.TableName()
	}

	// Default: lowercase struct name + s
	return strings.ToLower(t.Name()) + "s"
}

// TableNamer interface for models that define their own table name
type TableNamer interface {
	TableName() string
}
