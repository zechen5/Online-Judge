# Login Flow Subset

This subset packages the login-related frontend and backend into a smaller,
focused demo inside the main repository.

## Included

- Static frontend served by the subset backend
- `/api/v1/auth/register`
- `/api/v1/auth/login`
- `/api/v1/auth/me`
- JWT-based auth middleware
- SQLite-backed user storage
- Default admin bootstrap

## Run

```bash
go run ./login-flow-subset/cmd/server
```

Open `http://127.0.0.1:8090`.

## Defaults

- `OJ_LOGIN_SUBSET_PORT=8090`
- `OJ_LOGIN_SUBSET_DSN=login-flow-subset.db`
- `OJ_LOGIN_SUBSET_JWT_SECRET=login-subset-secret`
- `OJ_LOGIN_SUBSET_ADMIN_USERNAME=admin`
- `OJ_LOGIN_SUBSET_ADMIN_PASSWORD=admin123456`

## Notes

- The auth flow is real.
- The problem list on the page is mock data so the frontend still feels like
  the judge project without dragging in the full backend.
