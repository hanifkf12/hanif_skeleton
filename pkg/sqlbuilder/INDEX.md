# SQL Builder Package - Documentation Index

Selamat datang di dokumentasi SQL Builder! Package ini menyediakan query builder yang powerful dan mudah digunakan untuk operasi database dengan sqlx.

## 📚 Dokumentasi

### 1. [QUICKSTART.md](QUICKSTART.md) - **Mulai dari sini!** ⭐
Panduan quick start untuk segera menggunakan SQL Builder:
- Installation
- Basic usage examples (SELECT, INSERT, UPDATE, DELETE)
- Dynamic WHERE
- Pagination
- Bulk operations
- Migration guide dari repository lama
- Tips & best practices

### 2. [CHEATSHEET.md](CHEATSHEET.md) - **Reference cepat** 📋
Cheat sheet untuk reference cepat semua method dan pattern:
- Quick reference untuk semua operasi
- WHERE conditions
- JOIN, GROUP BY, HAVING
- Common patterns
- Repository method templates

### 3. [README.md](README.md) - **Dokumentasi lengkap** 📖
Dokumentasi lengkap dengan penjelasan detail:
- Semua features explained
- Advanced examples
- Integration dengan repository pattern
- Best practices

### 4. [SUMMARY.md](SUMMARY.md) - **Overview package** 📊
Summary lengkap tentang package:
- File structure
- Features implemented
- Testing results
- Migration status
- Key advantages

## 💻 Source Code

### Core Files
- **builder.go** - Core query builder dengan fluent interface
- **mapper.go** - Struct to map conversion utilities
- **model.go** - Model helper dengan database execution
- **helpers.go** - Advanced helpers (bulk, upsert, case, dll)

### Examples
- **example_repository.go** - Contoh lengkap repository implementation
- **campaign_v2_example.go** - Migration example dari old code

### Tests
- **builder_test.go** - 29 unit tests (all passing ✅)

## 🚀 Quick Start

```go
import "github.com/hanifkf12/hanif_skeleton/pkg/sqlbuilder"

// SELECT
var users []User
model := sqlbuilder.NewModel(db, &User{})
err := model.
    Table("users").
    Where("status = ?", "active").
    GetAll(ctx, &users)

// INSERT
user := &User{Name: "John"}
model.Table("users").Insert(ctx, user)

// UPDATE
model.Table("users").
    Where("id = ?", user.ID).
    Update(ctx, user)

// DELETE
model.Table("users").
    Where("id = ?", 1).
    Delete(ctx)
```

## 📖 Rekomendasi Urutan Baca

1. **Pemula**: QUICKSTART.md → CHEATSHEET.md
2. **Advanced**: README.md → example_repository.go
3. **Migration**: campaign_v2_example.go → Update repository Anda
4. **Reference**: CHEATSHEET.md (bookmark untuk daily use)

## ✨ Features Highlight

- ✅ Fluent interface yang chainable
- ✅ Auto mapping dari struct dengan db tags
- ✅ Dynamic query building
- ✅ Type-safe operations
- ✅ Built-in pagination
- ✅ Bulk operations
- ✅ Transaction ready
- ✅ SQL injection protected
- ✅ Comprehensive tests

## 🎯 Common Use Cases

### 1. Simple CRUD
```go
// Lihat: QUICKSTART.md section "Basic Usage"
```

### 2. Dynamic Filtering
```go
// Lihat: QUICKSTART.md section "Dynamic WHERE"
```

### 3. Pagination
```go
// Lihat: QUICKSTART.md section "Pagination"
```

### 4. Complex Queries with JOIN
```go
// Lihat: README.md section "JOINs"
```

### 5. Bulk Operations
```go
// Lihat: QUICKSTART.md section "Bulk Insert"
```

## 🧪 Testing

Semua tests passing:
```bash
go test -v ./pkg/sqlbuilder/...
# PASS - 29 tests
```

## 📦 Integration

Package sudah terintegrasi dengan:
- ✅ `pkg/databasex` - Database interface
- ✅ `pkg/telemetry` - OpenTelemetry tracing
- ✅ sqlx - SQL operations
- ✅ Repository pattern yang ada

## 🔧 Repository yang Sudah Updated

1. ✅ `internal/repository/campaign/campaign.go`
2. ✅ `internal/repository/user/user.go`

Lihat files tersebut untuk contoh real-world usage.

## 💡 Pro Tips

1. Gunakan **Model helper** untuk operasi standard
2. Gunakan **ConditionalBuilder** untuk dynamic WHERE
3. Gunakan **GetWithPagination()** untuk list dengan pagination
4. Use **ToSQL()** untuk debug queries
5. Check **CHEATSHEET.md** untuk quick reference

## 🆘 Need Help?

1. Check **QUICKSTART.md** untuk getting started
2. Check **CHEATSHEET.md** untuk syntax reference
3. Check **example_repository.go** untuk real examples
4. Check **README.md** untuk detailed explanation

## 📝 Migration Guide

Untuk migrate repository lama ke SQL Builder:

**Before:**
```go
query := `SELECT * FROM users WHERE id = ?`
err := db.Get(ctx, &user, query, id)
```

**After:**
```go
model := sqlbuilder.NewModel(db, &user)
err := model.Table("users").Where("id = ?", id).First(ctx, &user)
```

Lihat lengkapnya di QUICKSTART.md section "Migrasi dari Repository Lama".

---

Happy coding with SQL Builder! 🚀

**Last Updated**: November 12, 2025
**Version**: 1.0.0
**Status**: Production Ready ✅

