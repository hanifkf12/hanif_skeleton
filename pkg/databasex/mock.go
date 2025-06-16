package databasex

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
)

type mockData struct {
	rows []map[string]interface{}
}

type mockDB struct {
	data         map[string]*mockData
	inTransaction bool
}

func NewMockDB() Database {
	return &mockDB{
		data:         make(map[string]*mockData),
		inTransaction: false,
	}
}

func (m *mockDB) QueryX(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// Not implemented for mock
	return nil, errors.New("not implemented")
}

func (m *mockDB) QueryRowX(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// Not implemented for mock
	return nil
}

func (m *mockDB) Get(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if data, ok := m.data[query]; ok && len(data.rows) > 0 {
		return mapToStruct(data.rows[0], dst)
	}
	return sql.ErrNoRows
}

func (m *mockDB) Select(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if data, ok := m.data[query]; ok {
		return mapToSlice(data.rows, dst)
	}
	return nil
}

func (m *mockDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// Store the query and args for mock data
	if _, ok := m.data[query]; !ok {
		m.data[query] = &mockData{rows: make([]map[string]interface{}, 0)}
	}
	return nil, nil
}

func (m *mockDB) Transact(ctx context.Context, iso sql.IsolationLevel, txFunc func(database Database) error) (err error) {
	m.inTransaction = true
	defer func() { m.inTransaction = false }()
	return txFunc(m)
}

func (m *mockDB) InTransaction() bool {
	return m.inTransaction
}

// Helper functions to convert between maps and structs
func mapToStruct(data map[string]interface{}, dst interface{}) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr {
		return errors.New("destination must be a pointer")
	}
	v = v.Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if value, ok := data[field.Name]; ok {
			v.Field(i).Set(reflect.ValueOf(value))
		}
	}
	return nil
}

func mapToSlice(data []map[string]interface{}, dst interface{}) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return errors.New("destination must be a pointer to slice")
	}

	slice := reflect.MakeSlice(v.Elem().Type(), len(data), len(data))
	for i, item := range data {
		elemPtr := reflect.New(slice.Type().Elem())
		if err := mapToStruct(item, elemPtr.Interface()); err != nil {
			return err
		}
		slice.Index(i).Set(elemPtr.Elem())
	}

	v.Elem().Set(slice)
	return nil
}

// Mock data manipulation methods
func (m *mockDB) SetMockData(query string, data []map[string]interface{}) {
	m.data[query] = &mockData{rows: data}
}