# Online Judge Backend

Backend scaffold for the AI-assisted online judge proposal.

## Run

```bash
go run ./cmd/server
```

Defaults:

- `OJ_DB_DRIVER=sqlite`
- `OJ_DB_DSN=app.db`
- `OJ_PORT=8080`

Switch to MySQL by setting `OJ_DB_DRIVER=mysql` and `OJ_DB_DSN`.
