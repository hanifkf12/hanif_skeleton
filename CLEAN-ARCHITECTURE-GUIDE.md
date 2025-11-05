# Clean Architecture Analysis & Guide

## 📋 Apakah Project Ini Sudah Clean Architecture?

**Jawaban: YA**, project ini sudah mengikuti prinsip **Clean Architecture** dengan baik. Berikut analisisnya:

### ✅ Prinsip yang Sudah Diterapkan

#### 1. **Separation of Concerns**
```
internal/
├── entity/          # Domain entities (bisnis rules)
├── repository/      # Data access layer (interface + implementation)
├── usecase/         # Business logic layer
├── handler/         # Presentation layer adapter
└── router/          # Infrastructure routing
```

#### 2. **Dependency Rule**
- **Entity** tidak bergantung pada layer lain ✅
- **Repository** contract (interface) di layer domain ✅
- **UseCase** hanya bergantung pada entity dan repository interface ✅
- **Handler/Router** (outer layer) bergantung pada usecase ✅
- **Dependency Injection** dari outer ke inner layer ✅

#### 3. **Interface Abstraction**
```go
// Repository Contract (abstraksi)
type UserRepository interface {
    GetUsers(ctx context.Context) ([]entity.User, error)
    CreateUser(ctx context.Context, user entity.CreateUserRequest) (int64, error)
}

// UseCase Contract (abstraksi)
type UseCase interface {
    Serve(data appctx.Data) appctx.Response
}

// PubSub Consumer Contract (abstraksi)
type PubSubConsumer interface {
    Consume(data appctx.PubSubData) appctx.PubSubResponse
}
```

#### 4. **Testability**
- Semua layer menggunakan interface ✅
- Easy to mock dependencies ✅
- Business logic terpisah dari framework ✅

### 🔶 Area Improvement (Opsional)

1. **Logger/Telemetry dalam UseCase**
   - Saat ini: UseCase langsung import `pkg/logger` dan `pkg/telemetry`
   - Ideal: Inject logger interface ke usecase atau handle di middleware
   - **Status**: Acceptable trade-off untuk observability

2. **appctx.Data berisi framework (Fiber)**
   - Saat ini: `appctx.Data` membungkus `*fiber.Ctx`
   - Ideal: Pure domain types tanpa framework
   - **Status**: Pragmatic approach yang umum di Go

3. **Error Handling**
   - Tambahkan domain-level error types (`ErrNotFound`, `ErrInvalidInput`, dll)
   - Pisahkan business errors dari infrastructure errors

### 📊 Architecture Score

| Aspek | Score | Keterangan |
|-------|-------|------------|
| **Layer Separation** | ⭐⭐⭐⭐⭐ | Excellent - Jelas terpisah |
| **Dependency Direction** | ⭐⭐⭐⭐⭐ | Excellent - Inner tidak depend outer |
| **Interface Abstraction** | ⭐⭐⭐⭐⭐ | Excellent - Repository & UseCase abstrak |
| **Testability** | ⭐⭐⭐⭐⭐ | Excellent - Mudah di-mock |
| **Framework Independence** | ⭐⭐⭐⭐ | Good - Core logic tidak terikat framework |

**Overall: 4.8/5** - Very Good Clean Architecture Implementation

---

## 📚 Guide: Menambahkan Endpoint/Feature Baru

### Struktur Layer

```
┌─────────────────────────────────────────┐
│         External (HTTP/PubSub)          │
├─────────────────────────────────────────┤
│  Handler (internal/handler/)            │  ← Adapter
├─────────────────────────────────────────┤
│  Router (internal/router/)              │  ← Infrastructure
├─────────────────────────────────────────┤
│  UseCase (internal/usecase/)            │  ← Business Logic
├─────────────────────────────────────────┤
│  Repository (internal/repository/)      │  ← Data Access
├─────────────────────────────────────────┤
│  Entity (internal/entity/)              │  ← Domain
└─────────────────────────────────────────┘
```

---

## 🔧 Guide: Menambahkan HTTP Endpoint Baru

### Contoh: GET /products/:id

#### Step 1: Buat Entity (Domain Model)

File: `internal/entity/product.go`

```go
package entity

import "time"

type Product struct {
    ID          string    `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Description string    `json:"description" db:"description"`
    Price       float64   `json:"price" db:"price"`
    Stock       int       `json:"stock" db:"stock"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
```

#### Step 2: Tambah Repository Interface

File: `internal/repository/contract.go`

```go
type ProductRepository interface {
    GetByID(ctx context.Context, id string) (*entity.Product, error)
    GetAll(ctx context.Context) ([]entity.Product, error)
    Create(ctx context.Context, product *entity.Product) error
    Update(ctx context.Context, product *entity.Product) error
    Delete(ctx context.Context, id string) error
}
```

#### Step 3: Implementasi Repository

File: `internal/repository/product/product.go`

```go
package product

import (
    "context"
    "github.com/hanifkf12/hanif_skeleton/internal/entity"
    "github.com/hanifkf12/hanif_skeleton/internal/repository"
    "github.com/hanifkf12/hanif_skeleton/pkg/databasex"
    "github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type productRepository struct {
    db databasex.Database
}

func (r *productRepository) GetByID(ctx context.Context, id string) (*entity.Product, error) {
    ctx, span := telemetry.StartSpan(ctx, "productRepository.GetByID")
    defer span.End()

    query := `SELECT id, name, description, price, stock, created_at, updated_at 
              FROM products WHERE id = ?`
    
    var product entity.Product
    err := r.db.Get(ctx, &product, query, id)
    if err != nil {
        return nil, err
    }
    
    return &product, nil
}

func (r *productRepository) GetAll(ctx context.Context) ([]entity.Product, error) {
    ctx, span := telemetry.StartSpan(ctx, "productRepository.GetAll")
    defer span.End()

    query := `SELECT id, name, description, price, stock, created_at, updated_at 
              FROM products ORDER BY created_at DESC`
    
    var products []entity.Product
    err := r.db.Select(ctx, &products, query)
    if err != nil {
        return nil, err
    }
    
    return products, nil
}

// Implement other methods...

func NewProductRepository(db databasex.Database) repository.ProductRepository {
    return &productRepository{db: db}
}
```

#### Step 4: Buat UseCase

File: `internal/usecase/get_product.go`

```go
package usecase

import (
    "github.com/gofiber/fiber/v2"
    "github.com/hanifkf12/hanif_skeleton/internal/appctx"
    "github.com/hanifkf12/hanif_skeleton/internal/repository"
    "github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
    "github.com/hanifkf12/hanif_skeleton/pkg/logger"
    "github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type getProduct struct {
    productRepo repository.ProductRepository
}

func NewGetProduct(repo repository.ProductRepository) contract.UseCase {
    return &getProduct{productRepo: repo}
}

func (u *getProduct) Serve(data appctx.Data) appctx.Response {
    ctx := data.FiberCtx.UserContext()
    ctx, span := telemetry.StartSpan(ctx, "getProduct.Serve")
    defer span.End()

    lf := logger.NewFields("GetProduct").WithTrace(ctx)

    // Get ID from URL params
    id := data.FiberCtx.Params("id")
    if id == "" {
        logger.Error("Product ID is required", lf)
        return *appctx.NewResponse().
            WithCode(fiber.StatusBadRequest).
            WithErrors("Product ID is required")
    }

    lf.Append(logger.Any("product_id", id))

    // Get product from repository
    product, err := u.productRepo.GetByID(ctx, id)
    if err != nil {
        telemetry.SpanError(ctx, err)
        lf.Append(logger.Any("error", err.Error()))
        logger.Error("Failed to get product", lf)
        
        // Handle not found vs other errors
        return *appctx.NewResponse().
            WithCode(fiber.StatusNotFound).
            WithErrors("Product not found")
    }

    logger.Info("Product retrieved successfully", lf)
    return *appctx.NewResponse().WithCode(fiber.StatusOK).WithData(product)
}
```

#### Step 5: Register di Router

File: `internal/router/router.go`

```go
func (rtr *router) Route() {
    db := bootstrap.RegistryDatabase(rtr.cfg, false)
    
    // ... existing repositories ...
    productRepository := product.NewProductRepository(db)

    // Product routes
    getProductUseCase := usecase.NewGetProduct(productRepository)
    rtr.fiber.Get("/products/:id", rtr.handle(
        handler.HttpRequest,
        getProductUseCase,
    ))

    listProductsUseCase := usecase.NewListProducts(productRepository)
    rtr.fiber.Get("/products", rtr.handle(
        handler.HttpRequest,
        listProductsUseCase,
    ))

    // ... other routes ...
}
```

#### Step 6: Migration (Opsional)

File: `database/migration/20240325000000_create_table_products.sql`

```sql
-- +migrate Up
CREATE TABLE products (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE products;
```

---

## 🔔 Guide: Menambahkan Pub/Sub Consumer

### Contoh: Product Created Event

#### Step 1: Buat Consumer UseCase

File: `internal/usecase/product_created_consumer.go`

```go
package usecase

import (
    "encoding/json"
    "github.com/hanifkf12/hanif_skeleton/internal/appctx"
    "github.com/hanifkf12/hanif_skeleton/internal/entity"
    "github.com/hanifkf12/hanif_skeleton/internal/repository"
    "github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
    "github.com/hanifkf12/hanif_skeleton/pkg/logger"
    "github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type productCreatedConsumer struct {
    productRepo repository.ProductRepository
}

func NewProductCreatedConsumer(repo repository.ProductRepository) contract.PubSubConsumer {
    return &productCreatedConsumer{productRepo: repo}
}

func (c *productCreatedConsumer) Consume(data appctx.PubSubData) appctx.PubSubResponse {
    ctx, span := telemetry.StartSpan(data.Ctx, "productCreatedConsumer.Consume")
    defer span.End()

    lf := logger.NewFields("ProductCreatedConsumer").WithTrace(ctx)
    lf.Append(logger.Any("message_id", data.Message.ID))

    logger.Info("Processing product created message", lf)

    // Parse message
    var product entity.Product
    if err := json.Unmarshal(data.Message.Data, &product); err != nil {
        telemetry.SpanError(ctx, err)
        lf.Append(logger.Any("error", err.Error()))
        logger.Error("Failed to parse product message", lf)
        return *appctx.NewPubSubResponse().WithError(err)
    }

    lf.Append(logger.Any("product_name", product.Name))

    // Save to database
    if err := c.productRepo.Create(ctx, &product); err != nil {
        telemetry.SpanError(ctx, err)
        lf.Append(logger.Any("error", err.Error()))
        logger.Error("Failed to create product", lf)
        return *appctx.NewPubSubResponse().WithError(err)
    }

    logger.Info("Product created successfully", lf)
    return *appctx.NewPubSubResponse().WithMessage("Product created")
}
```

#### Step 2: Register Consumer

File: `cmd/pubsub/pubsub.go`

```go
// Initialize repositories
productRepository := product.NewProductRepository(db)

// Register subscription
router.RegisterSubscription(pubsubRouter.SubscriptionConfig{
    SubscriptionID: "product-created-subscription",
    Consumer:       usecase.NewProductCreatedConsumer(productRepository),
    MaxConcurrent:  10,
})
```

---

## 📤 Guide: Menggunakan Publisher

### Contoh: Publish dari HTTP Handler

#### Step 1: Inject Publisher ke UseCase

File: `internal/usecase/create_product.go`

```go
package usecase

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/hanifkf12/hanif_skeleton/internal/appctx"
    "github.com/hanifkf12/hanif_skeleton/internal/entity"
    "github.com/hanifkf12/hanif_skeleton/internal/repository"
    "github.com/hanifkf12/hanif_skeleton/internal/usecase/contract"
    "github.com/hanifkf12/hanif_skeleton/pkg/logger"
    "github.com/hanifkf12/hanif_skeleton/pkg/pubsub"
    "github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
)

type createProduct struct {
    productRepo repository.ProductRepository
    publisher   pubsub.Publisher
}

func NewCreateProduct(repo repository.ProductRepository, pub pubsub.Publisher) contract.UseCase {
    return &createProduct{
        productRepo: repo,
        publisher:   pub,
    }
}

func (u *createProduct) Serve(data appctx.Data) appctx.Response {
    ctx := data.FiberCtx.UserContext()
    ctx, span := telemetry.StartSpan(ctx, "createProduct.Serve")
    defer span.End()

    lf := logger.NewFields("CreateProduct").WithTrace(ctx)

    // Parse request
    var req entity.Product
    if err := data.FiberCtx.BodyParser(&req); err != nil {
        logger.Error("Invalid request", lf)
        return *appctx.NewResponse().
            WithCode(fiber.StatusBadRequest).
            WithErrors(err.Error())
    }

    // Generate ID
    req.ID = uuid.New().String()

    // Save to database
    if err := u.productRepo.Create(ctx, &req); err != nil {
        telemetry.SpanError(ctx, err)
        lf.Append(logger.Any("error", err.Error()))
        logger.Error("Failed to create product", lf)
        return *appctx.NewResponse().
            WithCode(fiber.StatusInternalServerError).
            WithErrors(err.Error())
    }

    // Publish event (async)
    go func() {
        if _, err := u.publisher.Publish(ctx, "product-created", req); err != nil {
            logger.Error("Failed to publish product created event", 
                logger.NewFields("CreateProduct").
                    Append(logger.Any("error", err.Error())))
        }
    }()

    logger.Info("Product created successfully", lf)
    return *appctx.NewResponse().WithCode(fiber.StatusCreated).WithData(req)
}
```

#### Step 2: Setup Publisher di Router

File: `internal/router/router.go`

```go
func (rtr *router) Route() {
    db := bootstrap.RegistryDatabase(rtr.cfg, false)
    
    // Initialize Pub/Sub client
    ctx := context.Background()
    projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
    
    var publisher pubsub.Publisher
    if projectID != "" {
        client, err := pubsubClient.NewClient(ctx, projectID)
        if err != nil {
            logger.Error("Failed to create Pub/Sub client", 
                logger.NewFields("Router").Append(logger.Any("error", err.Error())))
        } else {
            publisher = pubsub.NewPublisher(client)
        }
    }

    // Initialize repositories
    productRepository := product.NewProductRepository(db)

    // Routes with publisher
    if publisher != nil {
        createProductUseCase := usecase.NewCreateProduct(productRepository, publisher)
        rtr.fiber.Post("/products", rtr.handle(
            handler.HttpRequest,
            createProductUseCase,
        ))
    }
}
```

---

## 🧪 Testing Guide

### Unit Test UseCase

File: `internal/usecase/get_product_test.go`

```go
package usecase

import (
    "context"
    "testing"
    "github.com/gofiber/fiber/v2"
    "github.com/hanifkf12/hanif_skeleton/internal/appctx"
    "github.com/hanifkf12/hanif_skeleton/internal/entity"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock Repository
type MockProductRepository struct {
    mock.Mock
}

func (m *MockProductRepository) GetByID(ctx context.Context, id string) (*entity.Product, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entity.Product), args.Error(1)
}

// Test
func TestGetProduct_Success(t *testing.T) {
    // Setup
    mockRepo := new(MockProductRepository)
    usecase := NewGetProduct(mockRepo)

    expectedProduct := &entity.Product{
        ID:   "123",
        Name: "Test Product",
    }

    mockRepo.On("GetByID", mock.Anything, "123").Return(expectedProduct, nil)

    // Create test context
    app := fiber.New()
    ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
    ctx.Params("id", "123")

    data := appctx.Data{
        FiberCtx: ctx,
    }

    // Execute
    resp := usecase.Serve(data)

    // Assert
    assert.Equal(t, fiber.StatusOK, resp.Code)
    mockRepo.AssertExpectations(t)
}
```

---

## 📊 Checklist: Menambahkan Feature Baru

### HTTP Endpoint
- [ ] Buat entity di `internal/entity/`
- [ ] Tambah repository interface di `internal/repository/contract.go`
- [ ] Implementasi repository di `internal/repository/<resource>/`
- [ ] Buat usecase di `internal/usecase/`
- [ ] Register route di `internal/router/router.go`
- [ ] Tambah migration jika perlu di `database/migration/`
- [ ] Buat unit test untuk usecase
- [ ] Test manual dengan curl/Postman
- [ ] Update dokumentasi API

### Pub/Sub Consumer
- [ ] Buat entity untuk message format
- [ ] Buat consumer usecase di `internal/usecase/`
- [ ] Register subscription di `cmd/pubsub/pubsub.go`
- [ ] Buat unit test untuk consumer
- [ ] Test dengan publish message ke topic
- [ ] Setup monitoring/alerting

---

## 🏗️ Best Practices

### 1. **Single Responsibility**
Setiap usecase hanya handle satu action:
- ✅ `GetProduct`, `CreateProduct`, `UpdateProduct`
- ❌ `ProductManager` yang handle semua

### 2. **Dependency Injection**
```go
// ✅ Good - Inject interface
func NewGetProduct(repo repository.ProductRepository) contract.UseCase

// ❌ Bad - Direct instantiation
func NewGetProduct() contract.UseCase {
    repo := product.NewProductRepository(db) // Tight coupling
}
```

### 3. **Error Handling**
```go
// ✅ Good - Specific errors
if err == sql.ErrNoRows {
    return NotFoundError
}

// ❌ Bad - Generic error
if err != nil {
    return err
}
```

### 4. **Logging**
```go
// ✅ Good - Structured with context
lf := logger.NewFields("Operation").WithTrace(ctx)
lf.Append(logger.Any("user_id", userID))
logger.Info("Success", lf)

// ❌ Bad - Unstructured
log.Println("Success for user", userID)
```

---

## 🚀 Commands Reference

```bash
# Run HTTP server
go run main.go http

# Run Pub/Sub worker
go run main.go pubsub

# Run migration
go run main.go db:migrate

# Build
go build -o app main.go

# Test
go test ./... -v

# Test with coverage
go test ./... -cover

# Lint
golangci-lint run
```

---

## 📖 Additional Resources

- [README-pubsub.md](README-pubsub.md) - Dokumentasi lengkap Pub/Sub
- [README-logging.md](README-logging.md) - Logging guide
- [README-signoz-logging.md](README-signoz-logging.md) - SignOz integration

---

**Kesimpulan**: Project ini sudah mengimplementasikan Clean Architecture dengan sangat baik. Layer terpisah jelas, dependency injection diterapkan, dan mudah untuk di-test dan di-maintain. 🎉

