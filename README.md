# Expense Tracker API

A Personal Expense Tracker REST API built with Go and Beego v2.

## Overview

This project provides user registration, login, and expense management backed by CSV storage.
It is designed as an API-only backend with no frontend.

## Requirements

- Go 1.22 or higher
- Beego v2
- `bee` CLI for development (`go install github.com/beego/bee/v2@latest`)

## Setup

1. Clone the repository.
2. Install dependencies:

```bash
go mod download
```

3. Run this command in the terminal
```bash
cp conf/app.conf.sample conf/app.conf

4. Run the server:

```bash
bee run
```

The server starts on `http://localhost:8080` by default.

## Configuration

All configuration is stored in `conf/app.conf`.

- `appname = expense-tracker-api`
- `httpport = 8080`
- `runmode = dev`
- `copyrequestbody = true`
- `autorender = false`
- `users_csv_path = data/users.csv`
- `expenses_csv_path = data/expenses.csv`

The CSV files are created automatically when the API writes data.

## API Endpoints

### Health

`GET /api/v1/health`

### Authentication

`POST /api/v1/auth/register`

Request body:

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "secret123"
}
```

`POST /api/v1/auth/login`

Request body:

```json
{
  "email": "john@example.com",
  "password": "secret123"
}
```

### Expenses

All expense endpoints require the `X-User-ID` header.

`POST /api/v1/expenses`

`GET /api/v1/expenses`

Query parameters:
- `category`
- `date_from`
- `date_to`
- `sort_by` (`amount` or `expense_date`)
- `sort_order` (`asc` or `desc`)
- `limit`
- `page`

`GET /api/v1/expenses/:id`

`PUT /api/v1/expenses/:id`

`DELETE /api/v1/expenses/:id`

`GET /api/v1/expenses/summary?date_from=YYYY-MM-DD&date_to=YYYY-MM-DD`

## Example curl commands

Register:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"secret123"}'
```

Login:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"secret123"}'
```

Create expense:

```bash
curl -X POST http://localhost:8080/api/v1/expenses \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{"title":"Lunch","amount":350.50,"category":"Food","note":"Team lunch","expense_date":"2025-06-10"}'
```

Get expenses:

```bash
curl -X GET "http://localhost:8080/api/v1/expenses?limit=10&page=1" \
  -H "X-User-ID: 1"
```

Summary:

```bash
curl -X GET "http://localhost:8080/api/v1/expenses/summary?date_from=2025-06-01&date_to=2025-06-30" \
  -H "X-User-ID: 1"
```

## Testing

Run all tests:

```bash
go test ./...
```

Run static analysis:

```bash
go vet ./...
```

## Notes

- Data is stored in CSV format under `data/users.csv` and `data/expenses.csv`.
- Response format always includes `success`, `message`, and optional `data`.
