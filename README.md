# Personal Expense Tracker API

A RESTful API built with **Go** and **Beego v2** for tracking personal expenses.
All data is stored in CSV files using Go's standard `encoding/csv` library.

---

## Tech Stack

| Tool    | Version  |
|---------|----------|
| Go      | 1.22+    |
| Beego   | v2.3.4   |
| Bee CLI | v2       |
| Storage | CSV only |

---

## Project Structure
```
expense-tracker-api/
├── conf/
│   ├── app.conf.sample     # Copy this to app.conf and fill values
│
├── controllers/
│   ├── base.go             # Shared response helpers
│   ├── health.go           # Health check endpoint (Swagger documented)
│   ├── auth.go             # Register + Login APIs (Swagger enabled)
│   ├── middleware.go       # X-User-ID auth validation middleware
│   └── expense.go          # Expense CRUD + filtering + summary APIs
│
├── models/
│   ├── user.go             # User struct + CSV persistence logic
│   ├── expense.go          # Expense struct + CSV operations
│   └── filter.go           # Filtering, sorting, summary business logic
│
├── routers/
│   └── router.go           # All API routes + Swagger namespace registration
│
├── docs/
│   └── swagger.go          # Swagger entry point (generated + metadata)
│
├── swagger/                # Swagger UI static files (auto-generated / copied)
│   ├── index.html          # Swagger UI entry page
│   ├── swagger.json        # OpenAPI spec file
│   ├── swagger-ui-bundle.js
│   ├── swagger-ui.css
│   └── ...                 # Other Swagger assets
│
├── data/                   # Auto-created CSV storage (gitignored in production)
│
├── main.go                 # Application entry point
├── go.mod                  # Go module dependencies
└── README.md               # Project documentation
```
---

## Setup

### 1. Clone the repository

```bash
git clone https://github.com/sourav-islam/personal-expense-tracker-api.git
cd expense-tracker-api
```

### 2. Install Go dependencies

```bash
go mod tidy
```

### 3. Install Bee CLI

```bash
go install github.com/beego/bee/v2@latest
```

### 4. Configure the app

```bash
cp conf/app.conf.sample conf/app.conf
```

`conf/app.conf` is gitignored. The default values work as-is for local development.

### 5. Run the server

```bash
bee run
```

Or without bee:

```bash
go run main.go
```

Server starts at: `http://localhost:8080`

### 6. Generate & view Swagger docs

```bash
bee generate docs
# Then open: http://localhost:8080/swagger/
```

---

## Running Tests

```bash
# Run all tests
go test ./...

# With coverage report
go test ./... -cover

# Detailed coverage per function
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

---

## API Reference

### Health

| Method | Endpoint         | Auth | Description    |
|--------|------------------|------|----------------|
| GET    | /api/v1/health   | No   | Server status  |

### Auth

| Method | Endpoint                | Auth | Description      |
|--------|-------------------------|------|------------------|
| POST   | /api/v1/auth/register   | No   | Register account |
| POST   | /api/v1/auth/login      | No   | Login            |

### Expenses

| Method | Endpoint                      | Auth | Description              |
|--------|-------------------------------|------|--------------------------|
| POST   | /api/v1/expenses              | Yes  | Create expense           |
| GET    | /api/v1/expenses              | Yes  | List (filterable/sorted) |
| GET    | /api/v1/expenses/:id          | Yes  | Get one expense          |
| PUT    | /api/v1/expenses/:id          | Yes  | Update expense           |
| DELETE | /api/v1/expenses/:id          | Yes  | Delete expense           |
| GET    | /api/v1/expenses/summary      | Yes  | Spending summary         |

**Auth header:** `X-User-ID: <user_id>` (obtained from login response)

### Query Parameters for GET /api/v1/expenses

| Parameter  | Type   | Required | Description                          |
|------------|--------|----------|--------------------------------------|
| category   | string | No       | One of the allowed categories        |
| date_from  | string | No       | Start date YYYY-MM-DD                |
| date_to    | string | No       | End date YYYY-MM-DD                  |
| sort_by    | string | No       | `amount` or `expense_date`           |
| sort_order | string | No       | `asc` or `desc` (default: `desc`)    |
| limit      | int    | No       | Max results to return                |

### Allowed Categories

`Food` `Transport` `Housing` `Entertainment` `Shopping`
`Healthcare` `Education` `Utilities` `Other`

---

## Sample curl Commands

### Health Check

```bash
curl http://localhost:8080/api/v1/health
```

### Register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"secret123"}'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"secret123"}'
```

### Create Expense

```bash
curl -X POST http://localhost:8080/api/v1/expenses \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{"title":"Lunch","amount":350.50,"category":"Food","note":"Team lunch","expense_date":"2025-06-01"}'
```

### List All Expenses

```bash
curl http://localhost:8080/api/v1/expenses \
  -H "X-User-ID: 1"
```

### List with Filters

```bash
# By category
curl "http://localhost:8080/api/v1/expenses?category=Food" -H "X-User-ID: 1"

# By date range
curl "http://localhost:8080/api/v1/expenses?date_from=2025-06-01&date_to=2025-06-30" \
  -H "X-User-ID: 1"

# Sorted by amount descending
curl "http://localhost:8080/api/v1/expenses?sort_by=amount&sort_order=desc" \
  -H "X-User-ID: 1"

# Combined: Food in June, cheapest first, max 5
curl "http://localhost:8080/api/v1/expenses?category=Food&date_from=2025-06-01&date_to=2025-06-30&sort_by=amount&sort_order=asc&limit=5" \
  -H "X-User-ID: 1"
```

### Get One Expense

```bash
curl http://localhost:8080/api/v1/expenses/1 -H "X-User-ID: 1"
```

### Update Expense

```bash
curl -X PUT http://localhost:8080/api/v1/expenses/1 \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{"title":"Dinner","amount":500.00,"category":"Food","note":"Updated","expense_date":"2025-06-01"}'
```

### Delete Expense

```bash
curl -X DELETE http://localhost:8080/api/v1/expenses/1 -H "X-User-ID: 1"
```

### Summary

```bash
curl "http://localhost:8080/api/v1/expenses/summary?date_from=2025-06-01&date_to=2025-06-30" \
  -H "X-User-ID: 1"
```

---
