# auditable-gorm

Audit queries using [gorm](https://gorm.io/).

# Install 

```sh
$ go get github.com/berlim/auditable-gorm@v0.0.6
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

## Advanced

To change audit table name you can set `AUDIT_TABLE=yourcustomtable` environment variable.
