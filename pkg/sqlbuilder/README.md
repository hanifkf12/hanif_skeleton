# SQL Builder

SQL Builder adalah package untuk membangun query SQL secara dinamis dengan memanfaatkan struct tags dari sqlx. Package ini menyediakan interface yang fluent dan type-safe untuk operasi database.

## Features

- ✅ Query Builder dengan fluent interface
- ✅ Auto mapping dari struct ke SQL menggunakan db tags
- ✅ Support untuk SELECT, INSERT, UPDATE, DELETE
- ✅ WHERE conditions (AND, OR, IN, BETWEEN, NULL, etc.)
- ✅ JOIN support (INNER, LEFT, RIGHT)
- ✅ GROUP BY dan HAVING
- ✅ ORDER BY, LIMIT, OFFSET
- ✅ Pagination helpers
- ✅ Bulk Insert
- ✅ Upsert (ON DUPLICATE KEY UPDATE)
- ✅ Conditional WHERE builder
- ✅ CASE WHEN builder
- ✅ Raw query support

## Installation

Package ini sudah terintegrasi dengan project. Pastikan import path sudah benar:

```go
import "github.com/hanifkf12/hanif_skeleton/pkg/sqlbuilder"
```

## Quick Start

### 1. Define Model dengan DB Tags

```go
type User struct {
    ID        int       `json:"id" db:"id"`
    Name      string    `json:"name" db:"name"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Optional: Implement TableNamer untuk custom table name
func (u User) TableName() string {
    return "users"
}
```

### 2. Basic Query Builder

```go
// SELECT query
qb := sqlbuilder.NewQueryBuilder()
query, args := qb.
    Table("users").
    Select("id", "name", "email").
    Where("status = ?", "active").
    OrderBy("created_at", "DESC").
    Limit(10).
    Build()

// Execute dengan sqlx
var users []User
err := db.Select(ctx, &users, query, args...)
```

### 3. Model Helper (Recommended)

Model helper menyediakan cara yang lebih mudah dengan auto-execution:

```go
// Find all users
var users []User
model := sqlbuilder.NewModel(db, &User{})
err := model.
    Table("users").
    Where("status = ?", "active").
    OrderBy("created_at", "DESC").
    GetAll(ctx, &users)

// Find single user
var user User
err := model.
    Table("users").
    Where("id = ?", 1).
    First(ctx, &user)

// Count
count, err := model.
    Table("users").
    Where("status = ?", "active").
    Count(ctx)
```

## Examples

### SELECT Queries

#### Basic SELECT
```go
var users []User
model := sqlbuilder.NewModel(db, &User{})
err := model.
    Table("users").
    Select("id", "name", "email").
    GetAll(ctx, &users)
```

#### WHERE Conditions
```go
// Simple WHERE
model.Where("age > ?", 18)

// Multiple WHERE (AND)
model.
    Where("age > ?", 18).
    Where("status = ?", "active")

// OR WHERE
model.
    Where("age > ?", 18).
    OrWhere("is_premium = ?", true)

// WHERE IN
userIDs := []interface{}{1, 2, 3, 4, 5}
model.WhereIn("id", userIDs)

// WHERE BETWEEN
model.WhereBetween("age", 18, 30)

// WHERE NULL
model.WhereNull("deleted_at")

// WHERE NOT NULL
model.WhereNotNull("email_verified_at")
```

#### JOINs
```go
var results []struct {
    User
    RoleName string `db:"role_name"`
}

model := sqlbuilder.NewModel(db, nil)
err := model.
    Table("users").
    Select("users.*, roles.name as role_name").
    Join("roles", "users.role_id = roles.id").
    Where("users.status = ?", "active").
    GetAll(ctx, &results)
```

#### GROUP BY dan HAVING
```go
var stats []struct {
    Status string `db:"status"`
    Count  int    `db:"count"`
}

model := sqlbuilder.NewModel(db, nil)
err := model.
    Table("users").
    Select("status", "COUNT(*) as count").
    GroupBy("status").
    Having("COUNT(*) > ?", 10).
    GetAll(ctx, &stats)
```

#### Pagination
```go
var users []User
page := 1
perPage := 20

model := sqlbuilder.NewModel(db, &User{})
result, err := model.
    Table("users").
    Where("status = ?", "active").
    OrderBy("created_at", "DESC").
    GetWithPagination(ctx, &users, page, perPage)

// result contains:
// - result.Data (pointer to users slice)
// - result.Total (total records)
// - result.Page (current page)
// - result.PerPage
// - result.TotalPages
```

### INSERT Queries

#### Auto Insert from Struct
```go
user := User{
    Name:  "John Doe",
    Email: "john@example.com",
}

model := sqlbuilder.NewModel(db, &user)
result, err := model.
    Table("users").
    Insert(ctx, &user)

// Auto exclude id, created_at, updated_at
```

#### Insert with Specific Fields
```go
model := sqlbuilder.NewModel(db, &user)
result, err := model.
    Table("users").
    InsertWithFields(ctx, &user, "name", "email")
```

#### Manual Insert
```go
qb := sqlbuilder.NewQueryBuilder()
query, args := qb.
    Table("users").
    Insert(map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
    }).
    Build()

result, err := db.Exec(ctx, query, args...)
```

#### Bulk Insert
```go
users := []User{
    {Name: "User 1", Email: "user1@example.com"},
    {Name: "User 2", Email: "user2@example.com"},
    {Name: "User 3", Email: "user3@example.com"},
}

bulkInsert := sqlbuilder.NewBulkInsertBuilder("users").
    Columns("name", "email")

for _, user := range users {
    bulkInsert.AddFromStruct(&user)
}

query, args := bulkInsert.Build()
result, err := db.Exec(ctx, query, args...)
```

### UPDATE Queries

#### Auto Update from Struct
```go
user := User{
    ID:    1,
    Name:  "John Updated",
    Email: "john.updated@example.com",
}

model := sqlbuilder.NewModel(db, &user)
result, err := model.
    Table("users").
    Where("id = ?", user.ID).
    Update(ctx, &user)

// Auto exclude id, created_at, updated_at
```

#### Update with Specific Fields
```go
model := sqlbuilder.NewModel(db, &user)
result, err := model.
    Table("users").
    Where("id = ?", user.ID).
    UpdateWithFields(ctx, &user, "name", "email")
```

#### Manual Update
```go
qb := sqlbuilder.NewQueryBuilder()
query, args := qb.
    Table("users").
    Update(map[string]interface{}{
        "name":       "John Updated",
        "updated_at": time.Now(),
    }).
    Where("id = ?", 1).
    Build()

result, err := db.Exec(ctx, query, args...)
```

### DELETE Queries

```go
// Delete with Model
model := sqlbuilder.NewModel(db, nil)
result, err := model.
    Table("users").
    Where("id = ?", 1).
    Delete(ctx)

// Delete with conditions
result, err := model.
    Table("users").
    Where("status = ?", "inactive").
    Where("created_at < ?", time.Now().AddDate(0, -6, 0)).
    Delete(ctx)
```

## Advanced Features

### Conditional WHERE Builder

Berguna untuk membangun WHERE clause yang kompleks berdasarkan kondisi:

```go
cb := sqlbuilder.NewConditionalBuilder()

// Add condition hanya jika name tidak kosong
if name != "" {
    cb.Add("name LIKE ?", "%"+name+"%")
}

// Add condition hanya jika age valid
if age > 0 {
    cb.Add("age = ?", age)
}

// Atau menggunakan AddIf
cb.
    AddIf(name != "", "name LIKE ?", "%"+name+"%").
    AddIf(age > 0, "age = ?", age).
    AddIf(status != "", "status = ?", status)

if !cb.IsEmpty() {
    condition, args := cb.Build()
    model.Where(condition, args...)
}
```

### CASE WHEN Builder

```go
caseExpr := sqlbuilder.NewCaseBuilder().
    When("age < 18", "'minor'").
    When("age >= 18 AND age < 65", "'adult'").
    Else("'senior'").
    Build()

query := fmt.Sprintf("SELECT name, %s as age_group FROM users", caseExpr)
```

### Upsert (MySQL)

```go
upsert := sqlbuilder.NewUpsertBuilder("users").
    Insert(map[string]interface{}{
        "id":    1,
        "name":  "John",
        "email": "john@example.com",
    }).
    Update(map[string]interface{}{
        "name":       "John Updated",
        "updated_at": time.Now(),
    })

query, args := upsert.Build()
result, err := db.Exec(ctx, query, args...)
```

### Helper Functions

```go
// Find by ID
var user User
err := sqlbuilder.FindByID(ctx, db, "users", 1, &user)

// Find all
var users []User
err := sqlbuilder.FindAll(ctx, db, "users", &users)

// Delete by ID
result, err := sqlbuilder.DeleteByID(ctx, db, "users", 1)

// Create record
user := User{Name: "John", Email: "john@example.com"}
result, err := sqlbuilder.CreateRecord(ctx, db, "users", &user)

// Update record
result, err := sqlbuilder.UpdateRecord(ctx, db, "users", 1, &user)
```

### Struct Utilities

```go
// Convert struct to map
user := User{Name: "John", Email: "john@example.com"}
data := sqlbuilder.StructToMap(&user, false)
// Returns: map[string]interface{}{"name": "John", "email": "john@example.com", ...}

// Skip zero values
data := sqlbuilder.StructToMap(&user, true)

// Exclude specific fields
data := sqlbuilder.StructToMapExclude(&user, "id", "created_at", "updated_at")

// Include only specific fields
data := sqlbuilder.StructToMapInclude(&user, "name", "email")

// Get all columns from struct
columns := sqlbuilder.GetColumns(&user)
// Returns: []string{"id", "name", "email", "created_at", "updated_at"}

// Get columns excluding some
columns := sqlbuilder.GetColumnsExclude(&user, "created_at", "updated_at")

// Build SELECT column list
selectCols := sqlbuilder.BuildSelectColumns(&user, "u")
// Returns: "u.id, u.name, u.email, u.created_at, u.updated_at"
```

### Raw Query

```go
rawQuery := sqlbuilder.Raw("SELECT * FROM users WHERE id = ?", 1)
query, args := rawQuery.Build()

var user User
err := db.Get(ctx, &user, query, args...)
```

### Debug Query

```go
// Get SQL query without executing
model := sqlbuilder.NewModel(db, &User{})
model.Table("users").Where("id = ?", 1)

query, args := model.ToSQL()
fmt.Printf("Query: %s\nArgs: %v\n", query, args)
```

## Integration dengan Repository Pattern

```go
type campaignRepository struct {
    db databasex.Database
}

func (r *campaignRepository) GetByID(ctx context.Context, id string) (*entity.Campaign, error) {
    var campaign entity.Campaign
    
    model := sqlbuilder.NewModel(r.db, &campaign)
    err := model.
        Table("campaigns").
        Where("id = ?", id).
        First(ctx, &campaign)
    
    if err != nil {
        return nil, err
    }
    
    return &campaign, nil
}

func (r *campaignRepository) GetAll(ctx context.Context, filter CampaignFilter) ([]entity.Campaign, error) {
    var campaigns []entity.Campaign
    
    model := sqlbuilder.NewModel(r.db, &entity.Campaign{})
    model.Table("campaigns")
    
    // Dynamic filtering
    if filter.Status != "" {
        model.Where("status = ?", filter.Status)
    }
    
    if filter.MinDonation > 0 {
        model.Where("target_donation >= ?", filter.MinDonation)
    }
    
    if !filter.StartDate.IsZero() {
        model.Where("end_date >= ?", filter.StartDate)
    }
    
    err := model.
        OrderBy("created_at", "DESC").
        GetAll(ctx, &campaigns)
    
    return campaigns, err
}

func (r *campaignRepository) Create(ctx context.Context, campaign *entity.Campaign) error {
    campaign.ID = uuid.New().String()
    
    model := sqlbuilder.NewModel(r.db, campaign)
    _, err := model.
        Table("campaigns").
        Insert(ctx, campaign)
    
    return err
}

func (r *campaignRepository) Update(ctx context.Context, campaign *entity.Campaign) error {
    model := sqlbuilder.NewModel(r.db, campaign)
    _, err := model.
        Table("campaigns").
        Where("id = ?", campaign.ID).
        Update(ctx, campaign)
    
    return err
}
```

## Best Practices

1. **Gunakan Model Helper untuk operasi umum**: Lebih mudah dan mengurangi boilerplate
2. **Gunakan struct tags `db`**: Pastikan semua field yang perlu di-map ke database memiliki tag `db`
3. **Implement TableNamer**: Untuk custom table name, implement interface TableNamer
4. **Use Conditional Builder**: Untuk dynamic WHERE berdasarkan input user
5. **Pagination**: Gunakan `GetWithPagination` untuk list dengan pagination
6. **Transaction**: Gunakan dengan `db.Transact()` untuk operasi yang memerlukan transaction

## Notes

- Package ini menggunakan `?` sebagai placeholder (MySQL style)
- Auto-exclude `id`, `created_at`, `updated_at` pada Insert/Update (dapat di-override dengan `InsertWithFields`/`UpdateWithFields`)
- Semua method chainable untuk fluent interface
- Compatible dengan sqlx dan databasex package yang sudah ada

## License

MIT

