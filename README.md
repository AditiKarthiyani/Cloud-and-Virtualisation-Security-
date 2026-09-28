# CVS App — Frontend and Backend

This folder contains a frontend (Apache) and backend (Go + PostgreSQL) stack. The configuration is intentionally insecure so you can practice identifying and exploiting container, webserver, and host misconfigurations.

Quick commands:

Build images:
```
make build
```

Run backend (Go API + PostgreSQL):
```
make run-backend
```

Run frontend (Apache/PHP):
```
make run-frontend
```

Run both:
```
make run
```

Stop and cleanup:
```
make stop
make clean
```

Notes:
- Frontend served on port `8080` by Apache.
- Backend and PostgreSQL run together in one container on ports `8081` and `5432`.
- PostgreSQL credentials are set via environment variables: `POSTGRES_USER=postgres`, `POSTGRES_PASSWORD=postgres`, `POSTGRES_DB=exercise`.
- Database initialization is handled by `backend/db-init.sql`.

Endpoints:
- `http://localhost:8080` => Apache/PHP frontend todo app.
- `http://localhost:8081/api/todos` => list and create todos.
- `http://localhost:8081/api/todos/<id>` => update or delete todo.
- `http://localhost:8081/api/dbstatus` => PostgreSQL health / todo count.