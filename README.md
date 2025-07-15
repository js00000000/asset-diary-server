# Asset Diary Go API

This project is a Go-based RESTful API for managing assets, users, and related operations. It is designed to be modular, secure, and easy to deploy.

## Features
- User authentication (JWT-based)
- Profile management (view, update, change password)
- Account management (CRUD)
- Trade management (CRUD)
- Database migrations
- Dockerized development environment
- Configurable via environment variables

## Project Structure
```
server/
├── db/                # Database connection and logic
│   └── db.go
├── handlers/          # HTTP handlers (controllers)
│   ├── account.go
│   ├── auth.go
│   ├── profile.go
│   ├── refresh_logout.go
│   └── trade.go
├── middleware/        # JWT and other middleware
│   └── jwt.go
├── models/            # Data models
│   ├── account.go
│   ├── change_password.go
│   ├── investment_profile.go
│   ├── profile_response.go
│   ├── trade.go
│   ├── user.go
│   └── user_update.go
├── migrations/        # SQL migration files
├── routes/            # (reserved for route grouping)
├── main.go            # Entry point
├── docker-compose.yml # Docker configuration
├── .env               # Environment variables
├── go.mod, go.sum     # Go dependencies
├── openapi.json       # OpenAPI spec
```

## Getting Started

### Prerequisites
- Go 1.23+
- Docker & Docker Compose (optional, for containerized setup)
- PostgreSQL (or your chosen DB)

### Setup
1. Clone the repository:
   ```bash
   git clone <repo-url>
   cd asset-diary/server
   ```
2. Copy and configure environment variables:
   ```bash
   cp .env.example .env
   # Edit .env as needed
   ```
3. Run database migrations (using your preferred migration tool):
   ```bash
   # Example: migrate -path ./migrations -database $DATABASE_URL up
   ```
4. Start the API:
   ```bash
   go run main.go
   # or with Docker
   docker-compose up --build
   ```

### Usage
- The API will be available at `http://localhost:3000` by default.
- Use tools like Postman or curl to interact with the endpoints.
- See `openapi.json` for the full API specification.


## API Endpoints

### Auth
- `POST /api/auth/sign-in` — User login
- `POST /api/auth/sign-up` — User registration
- `POST /api/auth/refresh` — Refresh JWT
- `POST /api/auth/logout` — Logout (JWT required)

### Profile
- `GET /api/profile` — Get user profile (JWT required)
- `PUT /api/profile` — Update user profile (JWT required)
- `POST /api/profile/change-password` — Change password (JWT required)

### Environment Variables

- `DATABASE_URL` - PostgreSQL connection string (e.g., `postgres://user:pass@localhost:5432/dbname`)
- `JWT_SECRET` - Secret key for JWT signing
- `JWT_EXPIRATION` - JWT expiration time (e.g., `24h`)
- `PORT` - Port to run the server on (default: `8080`)
- `CRON_API_KEY` - API key required for accessing the cron endpoints

### Accounts
- `GET /api/accounts` — List accounts (JWT required)
- `POST /api/accounts` — Create account (JWT required)
- `PUT /api/accounts/:id` — Update account (JWT required)
- `DELETE /api/accounts/:id` — Delete account (JWT required)

### Trades
- `GET /api/trades` — List trades (JWT required)
- `POST /api/trades` — Create trade (JWT required)
- `PUT /api/trades/:id` — Update trade (JWT required)
- `DELETE /api/trades/:id` — Delete trade (JWT required)

### Holdings
- `GET /api/holdings` — List holdings (JWT required)

### Cron Endpoints
These endpoints are protected by API key authentication (X-API-Key header).
- `POST /api/cron/update-exchange-rates` — Updates all exchange rates from the external API
- `POST /api/cron/record-daily-assets-value` — Records the current total asset values for all users

## Development
- Code is organized by feature (handlers, models, db)
- Use Go modules for dependency management (`go.mod`, `go.sum`)
- Lint and test before pushing changes
- Environment variables are managed with `.env` (see `.env.example`)
- API documentation is available in `openapi.json` (Swagger UI integration possible)


## Dependencies
- [gin-gonic/gin](https://github.com/gin-gonic/gin) — HTTP web framework
- [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) — JWT authentication
- [joho/godotenv](https://github.com/joho/godotenv) — Environment variable loader
- [lib/pq](https://github.com/lib/pq) — PostgreSQL driver
- [google/uuid](https://github.com/google/uuid) — UUID support

## License
MIT

## Author
- [Your Name Here]

---
Feel free to customize this README to better fit your exact project details and team!
