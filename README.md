# auditable-gorm

Audit queries using [gorm](https://gorm.io/).

# Install 

```sh
$ github.com/berlim/auditable-gorm
```

# Usage

```go
b, err := gorm.Open(sqlite.Open(DB_PATH), &gorm.Config{})

// just call register on your database
auditableGorm.Register(db)
```

## Context

To pass UUID and Address:

```go
b, err := gorm.Open(sqlite.Open(DB_PATH), &gorm.Config{})

// just call register on your database
auditableGorm.Register(db)

auditData := AuditData{UUID: "your-uuid", Address: "your-ip"}
ctx := context.WithValue(context.Background(), auditableGorm.AUDIT_DATA_CTX_KEY, auditData)
db = db.WithContext(ctx)
```
