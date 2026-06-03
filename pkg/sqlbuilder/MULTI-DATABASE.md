# Multi-Database Support (MySQL & PostgreSQL)

SQL Builder sekarang mendukung **MySQL** dan **PostgreSQL** dengan automatic placeholder conversion!

## Overview

Package ini secara otomatis mengkonversi placeholder berdasarkan database driver:
- **MySQL**: Menggunakan `?` placeholders
- **PostgreSQL**: Menggunakan `$1, $2, $3, ...` placeholders

## Quick Start

### Set Default Driver (Global)

```go
import "github.com/hanifkf12/hanif_skeleton/pkg/sqlbuilder"

// Set default driver untuk semua query builder
sqlbuilder.SetDefaultDriver(sqlbuilder.DriverPostgreSQL)
// atau
sqlbuilder.SetDefaultDriver(sqlbuilder.DriverMySQL)
```

### Per Query Builder

```go
// MySQL
qb := sqlbuilder.NewQueryBuilderWithDriver(sqlbuilder.DriverMySQL)

// PostgreSQL
qb := sqlbuilder.NewQueryBuilderWithDriver(sqlbuilder.DriverPostgreSQL)
```

### With Model

```go
// MySQL Model
model := sqlbuilder.NewModel(db, &User{})

// PostgreSQL Model
model := sqlbuilder.NewModelWithDriver(db, &User{}, sqlbuilder.DriverPostgreSQL)
```

## Examples

### MySQL Query

```go
qb := sqlbuilder.NewQueryBuilderWithDriver(sqlbuilder.DriverMySQL)
query, args := qb.
    Table("users").
    Select("*").
    Where("age > ?", 18).
    Where("status = ?", "active").
    Build()

// Output:
// Query: SELECT * FROM users WHERE age > ? AND status = ?
// Args: [18, "active"]
```

### PostgreSQL Query

```go
qb := sqlbuilder.NewQueryBuilderWithDriver(sqlbuilder.DriverPostgreSQL)
query, args := qb.
    Table("users").
    Select("*").
    Where("age > ?", 18).
    Where("status = ?", "active").
    Build()

// Output:
// Query: SELECT * FROM users WHERE age > $1 AND status = $2
// Args: [18, "active"]
```

## Feature Support

### SELECT Queries

**MySQL:**
```go
qb := sqlbuilder.NewQueryBuilderWithDriver(sqlbuilder.DriverMySQL)
query, args := qb.
    Table("users").
    Select("*").
    Where("id IN (?, ?, ?)", 1, 2, 3).
    Build()
// Query: SELECT * FROM users WHERE id IN (?, ?, ?)
```

**PostgreSQL:**
```go
qb := sqlbuilder.NewQueryBuilderWithDriver(sqlbuilder.DriverPostgreSQL)
query, args := qb.
    Table("users").
    Select("*").
    WhereIn("id", []interface{}{1, 2, 3}).
    Build()
// Query: SELECT * FROM users WHERE id IN ($1, $2, $3)
```

### INSERT Queries

**MySQL:**
```go
qb := sqlbuilder.NewQueryBuilderWithDriver(sqlbuilder.DriverMySQL)
query, args := qb.
    Table("users").
    Insert(map[string]interface{}{
        "name": "John",
        "email": "john@example.com",
    }).
    Build()
// Query: INSERT INTO users (name, email) VALUES (?, ?)
```

**PostgreSQL:**
```go
qb := sqlbuilder.NewQueryBuilderWithDriver(sqlbuilder.DriverPostgreSQL)
query, args := qb.
    Table("users").
    Insert(map[string]interface{}{
        "name": "John",
        "email": "john@example.com",
    }).
    Build()
// Query: INSERT INTO users (name, email) VALUES ($1, $2)
```

### UPDATE Queries

**MySQL:**
```go
qb := sqlbuilder.NewQueryBuilderWithDriver(sqlbuilder.DriverMySQL)
query, args := qb.
    Table("users").
    Update(map[string]interface{}{
        "name": "John Updated",
    }).
    Where("id = ?", 1).
    Build()
// Query: UPDATE users SET name = ? WHERE id = ?
```

**PostgreSQL:**
```go
qb := sqlbuilder.NewQueryBuilderWithDriver(sqlbuilder.DriverPostgreSQL)
query, args := qb.
    Table("users").
    Update(map[string]interface{}{
        "name": "John Updated",
    }).
    Where("id = ?", 1).
    Build()
// Query: UPDATE users SET name = $1 WHERE id = $2
```

### Bulk Insert

**MySQL:**
```go
bi := sqlbuilder.NewBulkInsertBuilderWithDriver("users", sqlbuilder.DriverMySQL)
bi.Columns("name", "email")
bi.Values("User 1", "user1@example.com")
bi.Values("User 2", "user2@example.com")

query, args := bi.Build()
// Query: INSERT INTO users (name, email) VALUES (?, ?), (?, ?)
```

**PostgreSQL:**
```go
bi := sqlbuilder.NewBulkInsertBuilderWithDriver("users", sqlbuilder.DriverPostgreSQL)
bi.Columns("name", "email")
bi.Values("User 1", "user1@example.com")
bi.Values("User 2", "user2@example.com")

query, args := bi.Build()
// Query: INSERT INTO users (name, email) VALUES ($1, $2), ($3, $4)
```

### Upsert

**MySQL (ON DUPLICATE KEY UPDATE):**
```go
ub := sqlbuilder.NewUpsertBuilderWithDriver("users", sqlbuilder.DriverMySQL)
ub.Insert(map[string]interface{}{
    "id": 1,
    "name": "John",
    "email": "john@example.com",
})
ub.Update(map[string]interface{}{
    "name": "John Updated",
})

query, args := ub.Build()
// Query: INSERT INTO users (id, name, email) VALUES (?, ?, ?) 
//        ON DUPLICATE KEY UPDATE name = ?
```

**PostgreSQL (ON CONFLICT DO UPDATE):**
```go
ub := sqlbuilder.NewUpsertBuilderWithDriver("users", sqlbuilder.DriverPostgreSQL)
ub.OnConflict("email") // Specify conflict column
ub.Insert(map[string]interface{}{
    "name": "John",
    "email": "john@example.com",
})
ub.Update(map[string]interface{}{
    "name": "John Updated",
})

query, args := ub.Build()
// Query: INSERT INTO users (name, email) VALUES ($1, $2) 
//        ON CONFLICT (email) DO UPDATE SET name = $3
```

## Switching Drivers

### Switch at Runtime

```go
qb := sqlbuilder.NewQueryBuilder() // Default: MySQL

// Build MySQL query
query1, _ := qb.Table("users").Where("id = ?", 1).Build()
// Query: SELECT * FROM users WHERE id = ?

// Switch to PostgreSQL
qb.SetDriver(sqlbuilder.DriverPostgreSQL)

// Build PostgreSQL query
query2, _ := qb.Table("users").Where("id = ?", 1).Build()
// Query: SELECT * FROM users WHERE id = $1
```

### Model with Driver

```go
// Create model with specific driver
model := sqlbuilder.NewModelWithDriver(db, &User{}, sqlbuilder.DriverPostgreSQL)

// Or switch driver on existing model
model.SetDriver(sqlbuilder.DriverMySQL)
```

## Repository Pattern Integration

### Multi-Database Repository

```go
type UserRepository struct {
    db     databasex.Database
    driver sqlbuilder.DriverType
}

func NewUserRepository(db databasex.Database, driver sqlbuilder.DriverType) *UserRepository {
    return &UserRepository{
        db:     db,
        driver: driver,
    }
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*User, error) {
    var user User
    
    model := sqlbuilder.NewModelWithDriver(r.db, &user, r.driver)
    err := model.
        Table("users").
        Where("id = ?", id).
        First(ctx, &user)
    
    return &user, err
}
```

### Using in Application

```go
// For MySQL
mysqlRepo := NewUserRepository(mysqlDB, sqlbuilder.DriverMySQL)

// For PostgreSQL
postgresRepo := NewUserRepository(postgresDB, sqlbuilder.DriverPostgreSQL)
```

## Configuration Examples

### Option 1: Global Default

```go
func init() {
    // Set based on environment variable
    if os.Getenv("DB_DRIVER") == "postgres" {
        sqlbuilder.SetDefaultDriver(sqlbuilder.DriverPostgreSQL)
    } else {
        sqlbuilder.SetDefaultDriver(sqlbuilder.DriverMySQL)
    }
}
```

### Option 2: Configuration Struct

```go
type Config struct {
    DatabaseDriver string
}

func (c *Config) GetDriverType() sqlbuilder.DriverType {
    if c.DatabaseDriver == "postgres" || c.DatabaseDriver == "postgresql" {
        return sqlbuilder.DriverPostgreSQL
    }
    return sqlbuilder.DriverMySQL
}

// Usage
config := &Config{DatabaseDriver: "postgres"}
model := sqlbuilder.NewModelWithDriver(db, &User{}, config.GetDriverType())
```

### Option 3: Auto-detect from Connection String

```go
func DetectDriver(dsn string) sqlbuilder.DriverType {
    if strings.Contains(dsn, "postgres") {
        return sqlbuilder.DriverPostgreSQL
    }
    return sqlbuilder.DriverMySQL
}

// Usage
driver := DetectDriver(config.DatabaseDSN)
sqlbuilder.SetDefaultDriver(driver)
```

## Testing

### Test Coverage

✅ 43 tests total (all passing)
- 29 MySQL tests (original)
- 14 PostgreSQL tests (new)

### Running Tests

```bash
go test -v ./pkg/sqlbuilder/...
```

### Test Examples

```go
func TestPostgreSQLQuery(t *testing.T) {
    qb := sqlbuilder.NewQueryBuilderWithDriver(sqlbuilder.DriverPostgreSQL)
    query, args := qb.
        Table("users").
        Where("age > ?", 18).
        Build()
    
    assert.Contains(t, query, "$1")
    assert.Equal(t, []interface{}{18}, args)
}
```

## Driver Constants

```go
const (
    DriverMySQL      DriverType = 0  // MySQL/MariaDB
    DriverPostgreSQL DriverType = 1  // PostgreSQL
)
```

## Best Practices

1. **Set driver early**: Set default driver di init() atau main()
2. **Consistent usage**: Gunakan driver yang sama throughout your application
3. **Repository pattern**: Pass driver type to repositories
4. **Testing**: Test dengan kedua driver jika aplikasi support multi-database
5. **PostgreSQL ON CONFLICT**: Selalu specify conflict column dengan `OnConflict()`

## Migration from Single Database

### Before (MySQL only):
```go
model := sqlbuilder.NewModel(db, &User{})
```

### After (Multi-database):
```go
// MySQL
model := sqlbuilder.NewModel(db, &User{})
// atau explicit
model := sqlbuilder.NewModelWithDriver(db, &User{}, sqlbuilder.DriverMySQL)

// PostgreSQL
model := sqlbuilder.NewModelWithDriver(db, &User{}, sqlbuilder.DriverPostgreSQL)
```

## Compatibility Notes

- ✅ **MySQL/MariaDB**: Full support dengan `?` placeholders
- ✅ **PostgreSQL**: Full support dengan `$n` placeholders
- ✅ **Upsert**: Different syntax handled automatically
  - MySQL: `ON DUPLICATE KEY UPDATE`
  - PostgreSQL: `ON CONFLICT ... DO UPDATE`

## Performance

- Zero overhead untuk placeholder conversion
- Conversion dilakukan saat Build() dipanggil
- No runtime reflection untuk driver detection
- All conversions are string-based (very fast)

## Limitations

1. **Mixed drivers**: Satu QueryBuilder = satu driver type
2. **Manual placeholders**: Jika Anda menulis raw SQL dengan placeholder manual, pastikan sesuai dengan driver
3. **PostgreSQL ON CONFLICT**: Harus specify conflict column explicitly

## Future Enhancements

Possible future additions:
- SQLite support
- MSSQL support
- Automatic driver detection from database connection
- Named parameters support

---

**Documentation**: [README.md](README.md) | [QUICKSTART.md](QUICKSTART.md) | [CHEATSHEET.md](CHEATSHEET.md)

