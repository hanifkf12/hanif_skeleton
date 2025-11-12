# SQL Builder - Cheat Sheet

## Import
```go
import "github.com/hanifkf12/hanif_skeleton/pkg/sqlbuilder"
```

## Quick Reference

### SELECT
```go
// Basic
model := sqlbuilder.NewModel(db, &User{})
model.Table("users").Where("id = ?", 1).First(ctx, &user)

// Multiple records
model.Table("users").GetAll(ctx, &users)

// With conditions
model.Table("users").
    Where("age > ?", 18).
    Where("status = ?", "active").
    OrderBy("name").
    Limit(10).
    GetAll(ctx, &users)
```

### INSERT
```go
// Auto from struct
user := &User{Name: "John", Email: "john@example.com"}
model := sqlbuilder.NewModel(db, user)
model.Table("users").Insert(ctx, user)

// Manual
qb := sqlbuilder.NewQueryBuilder()
qb.Table("users").Insert(map[string]interface{}{
    "name": "John",
    "email": "john@example.com",
})
query, args := qb.Build()
db.Exec(ctx, query, args...)
```

### UPDATE
```go
// Auto from struct
user := &User{ID: 1, Name: "Updated"}
model := sqlbuilder.NewModel(db, user)
model.Table("users").Where("id = ?", user.ID).Update(ctx, user)

// Partial update
model.Table("users").
    Where("id = ?", 1).
    UpdateWithFields(ctx, user, "name", "email")
```

### DELETE
```go
model := sqlbuilder.NewModel(db, nil)
model.Table("users").Where("id = ?", 1).Delete(ctx)
```

### WHERE Conditions

```go
// Simple
.Where("age > ?", 18)

// Multiple (AND)
.Where("age > ?", 18).Where("status = ?", "active")

// OR
.Where("age > ?", 18).OrWhere("is_premium = ?", true)

// IN
ids := []interface{}{1, 2, 3}
.WhereIn("id", ids)

// NOT IN
.WhereNotIn("id", excludeIds)

// BETWEEN
.WhereBetween("age", 18, 30)

// NULL
.WhereNull("deleted_at")

// NOT NULL
.WhereNotNull("email_verified_at")
```

### Dynamic WHERE
```go
cb := sqlbuilder.NewConditionalBuilder()
cb.AddIf(name != "", "name LIKE ?", "%"+name+"%")
cb.AddIf(age > 0, "age = ?", age)

if !cb.IsEmpty() {
    condition, args := cb.Build()
    model.Where(condition, args...)
}
```

### JOIN
```go
model.Table("users").
    Select("users.*, roles.name as role_name").
    Join("roles", "users.role_id = roles.id")

// LEFT JOIN
.LeftJoin("profiles", "users.id = profiles.user_id")

// RIGHT JOIN
.RightJoin("departments", "users.dept_id = departments.id")
```

### ORDER BY
```go
.OrderBy("created_at")              // ASC default
.OrderBy("created_at", "DESC")
.OrderBy("name", "ASC")
```

### LIMIT & OFFSET
```go
.Limit(10)
.Offset(20)

// Or use pagination
.Paginate(page, perPage)
```

### GROUP BY & HAVING
```go
model.Table("users").
    Select("status", "COUNT(*) as count").
    GroupBy("status").
    Having("COUNT(*) > ?", 10)
```

### COUNT
```go
count, err := model.Table("users").
    Where("status = ?", "active").
    Count(ctx)
```

### EXISTS
```go
exists, err := model.Table("users").
    Where("email = ?", email).
    Exists(ctx)
```

### Pagination
```go
result, err := model.Table("users").
    Where("status = ?", "active").
    GetWithPagination(ctx, &users, page, perPage)

// Access pagination info
result.Total      // total records
result.Page       // current page
result.PerPage    // items per page
result.TotalPages // total pages
result.Data       // pointer to users slice
```

### Bulk Insert
```go
bulkInsert := sqlbuilder.NewBulkInsertBuilder("users")
for _, user := range users {
    bulkInsert.AddFromStruct(&user)
}
query, args := bulkInsert.Build()
db.Exec(ctx, query, args...)
```

### Upsert (MySQL)
```go
upsert := sqlbuilder.NewUpsertBuilder("users").
    Insert(map[string]interface{}{
        "id": 1,
        "name": "John",
    }).
    Update(map[string]interface{}{
        "name": "John Updated",
    })
query, args := upsert.Build()
db.Exec(ctx, query, args...)
```

### CASE WHEN
```go
caseExpr := sqlbuilder.NewCaseBuilder().
    When("age < 18", "'minor'").
    When("age >= 18", "'adult'").
    Else("'senior'").
    Build()
```

### Helper Functions
```go
// Find by ID
sqlbuilder.FindByID(ctx, db, "users", 1, &user)

// Find all
sqlbuilder.FindAll(ctx, db, "users", &users)

// Delete by ID
sqlbuilder.DeleteByID(ctx, db, "users", 1)

// Create
sqlbuilder.CreateRecord(ctx, db, "users", &user)

// Update
sqlbuilder.UpdateRecord(ctx, db, "users", 1, &user)
```

### Struct Utilities
```go
// To map
data := sqlbuilder.StructToMap(&user, false)

// Skip zero values
data := sqlbuilder.StructToMap(&user, true)

// Exclude fields
data := sqlbuilder.StructToMapExclude(&user, "id", "created_at")

// Include only
data := sqlbuilder.StructToMapInclude(&user, "name", "email")

// Get columns
columns := sqlbuilder.GetColumns(&user)

// Exclude columns
columns := sqlbuilder.GetColumnsExclude(&user, "password")

// Build SELECT
selectCols := sqlbuilder.BuildSelectColumns(&user, "u")
// Result: "u.id, u.name, u.email, ..."
```

### Debug Query
```go
model := sqlbuilder.NewModel(db, &User{})
model.Table("users").Where("id = ?", 1)

query, args := model.ToSQL()
fmt.Printf("Query: %s\nArgs: %v\n", query, args)
```

### Raw Query
```go
raw := sqlbuilder.Raw("SELECT * FROM users WHERE id = ?", 1)
query, args := raw.Build()
db.Get(ctx, &user, query, args...)
```

## Common Patterns

### Repository Method Template
```go
func (r *repo) GetByID(ctx context.Context, id string) (*Entity, error) {
    ctx, span := telemetry.StartSpan(ctx, "Repository.GetByID")
    defer span.End()

    var entity Entity
    model := sqlbuilder.NewModel(r.db, &entity)
    err := model.
        Table("table_name").
        Where("id = ?", id).
        First(ctx, &entity)
    
    if err != nil {
        return nil, err
    }
    return &entity, nil
}
```

### Search with Filters
```go
func (r *repo) Search(ctx context.Context, filters Filters) ([]Entity, error) {
    var entities []Entity
    
    cb := sqlbuilder.NewConditionalBuilder()
    cb.AddIf(filters.Name != "", "name LIKE ?", "%"+filters.Name+"%")
    cb.AddIf(filters.Status != "", "status = ?", filters.Status)
    cb.AddIf(filters.MinPrice > 0, "price >= ?", filters.MinPrice)
    
    model := sqlbuilder.NewModel(r.db, &Entity{})
    model.Table("table_name")
    
    if !cb.IsEmpty() {
        condition, args := cb.Build()
        model.Where(condition, args...)
    }
    
    return entities, model.GetAll(ctx, &entities)
}
```

### Paginated List
```go
func (r *repo) GetPaginated(ctx context.Context, page, perPage int) (*sqlbuilder.PaginationResult, error) {
    var entities []Entity
    
    model := sqlbuilder.NewModel(r.db, &Entity{})
    return model.
        Table("table_name").
        OrderBy("created_at", "DESC").
        GetWithPagination(ctx, &entities, page, perPage)
}
```

## Tips
- Model helper auto-excludes `id`, `created_at`, `updated_at` on Insert/Update
- Use `First()` for single record, `GetAll()` for multiple
- Use `ConditionalBuilder` for dynamic filters
- Chain methods for readable code
- Use `ToSQL()` for debugging
- All queries are parameterized (safe from SQL injection)

