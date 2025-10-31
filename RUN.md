# 🚀 Cara Menjalankan Project

## Prerequisites
- Go 1.23+
- **SQLite** (built-in, no setup needed!)

## Quick Start - SATU COMMAND!

```bash
go run cmd/api/main.go
```

Server akan running di: **http://localhost:8080**

---

## 📝 Test API dengan cURL

### Health Check
```bash
curl http://localhost:8080/health
```

### Create User
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","age":25}'
```

### Get All Users
```bash
curl http://localhost:8080/api/users
```

### Get User by ID
```bash
curl http://localhost:8080/api/users/1
```

### Update User
```bash
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Doe","age":26}'
```

### Delete User
```bash
curl -X DELETE http://localhost:8080/api/users/1
```

### 🔥 Batch Create (Goroutine Example)
```bash
curl -X POST http://localhost:8080/api/users/batch \
  -H "Content-Type: application/json" \
  -d '[{"name":"Alice","email":"alice@example.com","age":25},{"name":"Bob","email":"bob@example.com","age":30},{"name":"Charlie","email":"charlie@example.com","age":35}]'
```

### �� Get Stats (Goroutine Example)
```bash
curl http://localhost:8080/api/users/stats
```

---

## 🎯 Features Implemented

✅ **GORM** - SQLite ORM (No setup needed!)  
✅ **Goroutines** - Concurrent operations (batch create, stats)  
✅ **Context** - Request timeout & cancellation  
✅ **Auto Migration** - Database schema otomatis  
✅ **Batch Operations** - Create multiple users concurrently  
✅ **Statistics** - Calculate stats with goroutines  
✅ **Soft Delete** - GORM soft delete feature  
✅ **Thread-Safe** - Concurrent access dengan context  

---

## 📁 Database File

Data disimpan di file: `goproject.db` (otomatis dibuat)
