# SQL Builder - Quick Start Guide

## Installation

Package sudah tersedia di `pkg/sqlbuilder`. Import dengan:

```go
import "github.com/hanifkf12/hanif_skeleton/pkg/sqlbuilder"
```

## Basic Usage

### 1. Simple SELECT

```go
// Menggunakan Model helper (recommended)
var users []User
model := sqlbuilder.NewModel(db, &User{})
err := model.
    Table("users").
    Where("status = ?", "active").
    GetAll(ctx, &users)
```

### 2. SELECT dengan Kondisi Kompleks

```go
var campaigns []Campaign
model := sqlbuilder.NewModel(db, &Campaign{})
err := model.
    Table("campaigns").
    Where("target_donation >= ?", 1000000).
    Where("end_date > ?", time.Now()).
    OrderBy("created_at", "DESC").
    Limit(10).
    GetAll(ctx, &campaigns)
```

### 3. Dynamic WHERE (Berdasarkan Input User)

```go
// Filter yang optional
name := "John"
age := 0  // tidak diisi
status := "active"

// Gunakan ConditionalBuilder
cb := sqlbuilder.NewConditionalBuilder()
cb.AddIf(name != "", "name LIKE ?", "%"+name+"%")
cb.AddIf(age > 0, "age = ?", age)  // akan di-skip karena age = 0
cb.AddIf(status != "", "status = ?", status)

model := sqlbuilder.NewModel(db, &User{})
model.Table("users")

if !cb.IsEmpty() {
    condition, args := cb.Build()
    model.Where(condition, args...)
}

err := model.GetAll(ctx, &users)
// Query: SELECT * FROM users WHERE name LIKE ? AND status = ?
// Args: ["%John%", "active"]
```

### 4. INSERT (Auto dari Struct)

```go
campaign := &entity.Campaign{
    Name:           "Campaign Baru",
    TargetDonation: 5000000,
    EndDate:        time.Now().AddDate(0, 1, 0),
}

model := sqlbuilder.NewModel(db, campaign)
_, err := model.
    Table("campaigns").
    Insert(ctx, campaign)

// Auto exclude: id, created_at, updated_at
```

### 5. UPDATE (Auto dari Struct)

```go
campaign := &entity.Campaign{
    ID:             "uuid-here",
    Name:           "Campaign Updated",
    TargetDonation: 7000000,
}

model := sqlbuilder.NewModel(db, campaign)
_, err := model.
    Table("campaigns").
    Where("id = ?", campaign.ID).
    Update(ctx, campaign)
```

### 6. UPDATE Partial (Hanya Field Tertentu)

```go
campaign := &entity.Campaign{
    ID:   "uuid-here",
    Name: "New Name",
}

model := sqlbuilder.NewModel(db, campaign)
_, err := model.
    Table("campaigns").
    Where("id = ?", campaign.ID).
    UpdateWithFields(ctx, campaign, "name")
```

### 7. DELETE

```go
model := sqlbuilder.NewModel(db, nil)
_, err := model.
    Table("campaigns").
    Where("id = ?", campaignID).
    Delete(ctx)
```

### 8. Pagination

```go
var campaigns []Campaign
page := 1
perPage := 20

model := sqlbuilder.NewModel(db, &Campaign{})
result, err := model.
    Table("campaigns").
    Where("status = ?", "active").
    OrderBy("created_at", "DESC").
    GetWithPagination(ctx, &campaigns, page, perPage)

// result.Data -> slice campaigns
// result.Total -> total records
// result.Page -> current page
// result.PerPage -> items per page
// result.TotalPages -> total pages
```

### 9. COUNT

```go
model := sqlbuilder.NewModel(db, nil)
count, err := model.
    Table("campaigns").
    Where("status = ?", "active").
    Count(ctx)
```

### 10. WHERE IN

```go
ids := []string{"id1", "id2", "id3"}
idsInterface := make([]interface{}, len(ids))
for i, id := range ids {
    idsInterface[i] = id
}

var campaigns []Campaign
model := sqlbuilder.NewModel(db, &Campaign{})
err := model.
    Table("campaigns").
    WhereIn("id", idsInterface).
    GetAll(ctx, &campaigns)
```

### 11. JOIN

```go
type CampaignWithUser struct {
    Campaign
    UserName string `db:"user_name"`
}

var results []CampaignWithUser
model := sqlbuilder.NewModel(db, nil)
err := model.
    Table("campaigns").
    Select("campaigns.*, users.name as user_name").
    Join("users", "campaigns.user_id = users.id").
    Where("campaigns.status = ?", "active").
    GetAll(ctx, &results)
```

### 12. Bulk Insert

```go
campaigns := []Campaign{
    {Name: "Campaign 1", TargetDonation: 1000000},
    {Name: "Campaign 2", TargetDonation: 2000000},
    {Name: "Campaign 3", TargetDonation: 3000000},
}

bulkInsert := sqlbuilder.NewBulkInsertBuilder("campaigns")
for _, c := range campaigns {
    bulkInsert.AddFromStruct(&c)
}

query, args := bulkInsert.Build()
_, err := db.Exec(ctx, query, args...)
```

## Migrasi dari Repository Lama

### Sebelum (Manual Query):

```go
func (c *campaignRepository) GetByID(ctx context.Context, id string) (*entity.Campaign, error) {
    query := `SELECT * FROM campaigns WHERE id = ?`
    var campaign entity.Campaign
    err := c.db.Get(ctx, &campaign, query, id)
    if err != nil {
        return nil, err
    }
    return &campaign, nil
}
```

### Sesudah (SQL Builder):

```go
func (c *campaignRepository) GetByID(ctx context.Context, id string) (*entity.Campaign, error) {
    var campaign entity.Campaign
    model := sqlbuilder.NewModel(c.db, &campaign)
    err := model.
        Table("campaigns").
        Where("id = ?", id).
        First(ctx, &campaign)
    
    if err != nil {
        return nil, err
    }
    return &campaign, nil
}
```

## Tips & Best Practices

1. **Gunakan Model Helper** untuk operasi umum (lebih simple)
2. **Gunakan QueryBuilder** langsung jika perlu custom query yang kompleks
3. **ConditionalBuilder** sangat berguna untuk dynamic filtering
4. **Selalu gunakan parameterized queries** (sudah otomatis dengan builder ini)
5. **Untuk pagination**, gunakan `GetWithPagination()` daripada manual LIMIT/OFFSET
6. **Auto-exclude fields**: Insert/Update otomatis exclude `id`, `created_at`, `updated_at`

## Debug Query

Untuk melihat query yang akan dijalankan:

```go
model := sqlbuilder.NewModel(db, &User{})
model.Table("users").Where("id = ?", 1)

query, args := model.ToSQL()
fmt.Printf("Query: %s\nArgs: %v\n", query, args)
```

## Helper Functions

```go
// Find by ID
var user User
err := sqlbuilder.FindByID(ctx, db, "users", 1, &user)

// Delete by ID
result, err := sqlbuilder.DeleteByID(ctx, db, "users", 1)

// Create record
user := User{Name: "John"}
result, err := sqlbuilder.CreateRecord(ctx, db, "users", &user)
```

## Advanced Features

### CASE WHEN

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
        "name": "John Updated",
    })

query, args := upsert.Build()
_, err := db.Exec(ctx, query, args...)
```

## Struct Utilities

```go
user := User{ID: 1, Name: "John", Email: "john@example.com"}

// Convert to map
data := sqlbuilder.StructToMap(&user, false)

// Skip zero values
data := sqlbuilder.StructToMap(&user, true)

// Exclude fields
data := sqlbuilder.StructToMapExclude(&user, "id", "created_at")

// Include only specific fields
data := sqlbuilder.StructToMapInclude(&user, "name", "email")

// Get all columns
columns := sqlbuilder.GetColumns(&user)
```

## Contoh Lengkap di Repository

Lihat file:
- `pkg/sqlbuilder/example_repository.go` - Contoh lengkap campaign repository
- `internal/repository/campaign/campaign_v2_example.go` - Migration example
- `pkg/sqlbuilder/README.md` - Dokumentasi lengkap

## Testing

Package sudah dilengkapi dengan unit tests:

```bash
go test -v ./pkg/sqlbuilder/...
```

Semua test sudah passing ✅

