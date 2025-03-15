# Database System

The database system provides a connection to PostgreSQL using GORM for object-relational mapping.

## Architecture

The database system is structured using a clean architecture approach:

- `gorm.go` - GORM database connection and utilities

## Components

### GORM Database

The `GormDB` struct wraps a GORM database connection:

```go
type GormDB struct {
    DB *gorm.DB
}
```

It provides methods for connecting to the database, closing the connection, and running auto migrations.

## Configuration

Database connection parameters are configured through environment variables:

- `DB_HOST` - PostgreSQL host (default: "localhost")
- `DB_PORT` - PostgreSQL port (default: "5432")
- `DB_USER` - PostgreSQL user (default: "postgres")
- `DB_PASSWORD` - PostgreSQL password (default: "postgres")
- `DB_NAME` - PostgreSQL database name (default: "bespin")

## Usage

### Creating a GORM Database Connection

```go
// Create a GORM database connection
db, err := database.NewGormDB()
if err != nil {
    log.Fatalf("Failed to connect to database: %v", err)
}
```

### Running Auto Migrations

```go
// Auto migrate models
if err := db.AutoMigrate(&models.WebhookReceipt{}); err != nil {
    log.Fatalf("Failed to run auto migrations: %v", err)
}
```

### Closing the Database Connection

```go
// Close the database connection
if err := db.Close(); err != nil {
    log.Printf("Error closing database connection: %v", err)
}
```

## Models

The database system works with models defined in the `pkg/models` package. The main models include:

- `WebhookReceipt` - Stores received webhooks

### WebhookReceipt Model

```go
type WebhookReceipt struct {
    ID        string         `json:"id" gorm:"primaryKey;type:uuid"`
    Source    string         `json:"source" gorm:"index;type:varchar(255)"`
    Event     string         `json:"event" gorm:"index;type:varchar(255)"`
    Payload   datatypes.JSON `json:"payload" gorm:"type:jsonb"`
    Headers   datatypes.JSON `json:"headers" gorm:"type:jsonb"`
    Signature string         `json:"signature" gorm:"type:text"`
    Verified  bool           `json:"verified" gorm:"index"`
    CreatedAt time.Time      `json:"created_at" gorm:"index;autoCreateTime"`
}
```

## GORM Features

The database system leverages GORM's features:

### Type-Safe Database Operations

GORM provides type-safe database operations, reducing the risk of runtime errors.

### Automatic Migrations

GORM can automatically create and update database tables based on struct definitions.

### Relationship Management

GORM supports various relationship types (one-to-one, one-to-many, many-to-many).

### Query Building

GORM provides a fluent API for building complex queries.

### Transaction Support

GORM supports database transactions for atomic operations.

## Examples

### Basic CRUD Operations

```go
// Create
db.Create(&models.WebhookReceipt{
    ID:        uuid.New().String(),
    Source:    "github",
    Event:     "push",
    Payload:   payload,
    Headers:   headers,
    Signature: signature,
    Verified:  true,
    CreatedAt: time.Now(),
})

// Read
var receipt models.WebhookReceipt
db.First(&receipt, "id = ?", id)

// Update
db.Model(&receipt).Update("verified", true)

// Delete
db.Delete(&receipt)
```

### Querying

```go
// Find all webhook receipts from a specific source
var receipts []models.WebhookReceipt
db.Where("source = ?", "github").Find(&receipts)

// Find all verified webhook receipts
db.Where("verified = ?", true).Find(&receipts)

// Find webhook receipts with pagination
db.Limit(10).Offset(0).Order("created_at DESC").Find(&receipts)

// Count webhook receipts
var count int64
db.Model(&models.WebhookReceipt{}).Count(&count)
```

### Transactions

```go
// Start a transaction
tx := db.Begin()

// Perform operations
if err := tx.Create(&receipt).Error; err != nil {
    tx.Rollback()
    return err
}

// Commit the transaction
tx.Commit()
```
