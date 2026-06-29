# Card Authorization Service

A backend service for managing prepaid cards and simulating card authorization flows. The service supports card creation, balance top-up, freezing/unfreezing cards, purchase authorization, transaction reversal, idempotent authorization requests, and transaction history.

This project was built as a focused card authorization API using Go, Gin, PostgreSQL, GORM, and database migrations.

## Highlights

- Create prepaid cards with generated card numbers
- Retrieve card details and current balance
- Freeze and unfreeze cards
- Top up card balance
- Authorize transactions against card balance
- Decline transactions for frozen cards, missing cards, currency mismatch, and insufficient funds
- Reverse approved authorizations and refund the card balance
- Use row-level locking during balance deduction and reversal
- Support idempotency keys for authorization requests
- Store schema changes in SQL migrations
- Cover service-layer behavior with unit tests and mocks
- Provide Docker and Docker Compose setup for local development

## Tech Stack

- Go
- Gin HTTP framework
- PostgreSQL
- GORM
- golang-migrate
- Zap logger
- Viper configuration
- Testify for unit tests
- Docker and Docker Compose

## Project Structure

```text
cmd/server/              Application entry point and route wiring
config/                  Environment configuration loader
internal/dto/            Request and response payloads
internal/errors/         Application error codes and mapping helpers
internal/handler/        HTTP handlers
internal/middleware/     Request logging and idempotency middleware
internal/model/          Database models
internal/repository/     Database access layer
internal/service/        Business logic
migrations/              SQL migration files
pkg/cardnumber/          Card number generator
pkg/database/            PostgreSQL connection and transaction manager
pkg/logger/              Zap logger setup
```

## Core Flow

Authorization follows this flow:

1. Receive transaction request with card number, merchant, amount, and currency.
2. Check idempotency key when provided.
3. Find the card by card number.
4. Decline if the card does not exist, is frozen, has a currency mismatch, or has insufficient balance.
5. Lock the card row with `SELECT FOR UPDATE`.
6. Deduct the amount and create an approved authorization record inside one database transaction.
7. Return the authorization result and remaining balance.

Reversal locks the card row again, restores the approved amount, and marks the authorization as `REVERSED`.

## API Endpoints

### Health Check

```http
GET /health
```

### Cards

```http
POST /cards
GET /cards/:id
POST /cards/:id/freeze
POST /cards/:id/unfreeze
POST /cards/:id/topup
GET /cards/:id/transactions
```

### Transactions

```http
POST /transactions/authorize
POST /transactions/:authorizationId/reverse
```

`POST /transactions/authorize` supports an optional `Idempotency-Key` header.

## Example Requests

### Create a Card

```bash
curl -X POST http://localhost:8080/cards \
  -H "Content-Type: application/json" \
  -d '{
    "cardholderName": "Dewa Ditya Sanjaya",
    "currency": "IDR",
    "initialBalance": 500000
  }'
```

### Top Up a Card

```bash
curl -X POST http://localhost:8080/cards/{cardId}/topup \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100000
  }'
```

### Authorize a Transaction

```bash
curl -X POST http://localhost:8080/transactions/authorize \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: request-001" \
  -d '{
    "cardNumber": "1234567890123456",
    "merchantId": "MRC-001",
    "merchantName": "Coffee Shop",
    "currency": "IDR",
    "amount": 75000
  }'
```

### Reverse an Authorization

```bash
curl -X POST http://localhost:8080/transactions/{authorizationId}/reverse
```

## Environment Variables

Create a `.env` file from `.env.example`.

```env
APP_PORT=8080
APP_ENV=development

DB_HOST=localhost
DB_USER=cardauth
DB_PASSWORD=cardauth_secret
DB_NAME=card_auth_db
DB_PORT=5432
DB_SSLMODE=disable
```

When running the app service inside Docker Compose, use `DB_HOST=postgres` because the application connects through the Compose network.

## Running Locally

### 1. Start PostgreSQL

```bash
docker compose up -d postgres
```

### 2. Run Database Migrations

```bash
make migrate-up
```

The Makefile uses this database URL:

```text
postgres://cardauth:cardauth_secret@localhost:5432/card_auth_db?sslmode=disable
```

Update `DB_URL` in the Makefile if your local database credentials are different.

### 3. Start the API

```bash
make run
```

The API will listen on:

```text
http://localhost:8080
```

## Running With Docker Compose

```bash
docker compose up --build
```

Note: migrations are not automatically executed by the application container. Run migrations before starting the full stack, or add a migration step to the Compose workflow.

## Testing

Run all tests:

```bash
make test
```

Run service tests only:

```bash
make test-service
```

## Database Schema

The service uses three main tables:

- `cards`: stores card number, cardholder name, status, currency, and balance
- `authorizations`: stores approved/reversed authorization records
- `idempotency_keys`: stores processed idempotency keys for authorization requests

Migrations are stored in the `migrations/` directory and include matching `up` and `down` files.

## Comments

This project has a clean separation between handlers, services, repositories, models, and DTOs. The service layer is where most business rules live, which makes the code easier to test. Using a transaction manager abstraction is also a good decision because it keeps transaction behavior testable with mocks.

The strongest part of the implementation is the authorization flow: it checks card state, validates currency, handles insufficient funds, locks the card row before balance mutation, and records the authorization inside a database transaction. That is the right direction for a card/payment-style service where concurrency matters.

## Future Improvements

- Use an integer minor-unit money representation, such as cents, instead of `float64` for balances and amounts.
- Store declined authorization attempts too, so transaction history can include both approved and declined decisions.
- Improve idempotency response storage so repeated requests can return the original response exactly, including remaining balance and decline reason.
- Add idempotency support for reversal requests.
- Add request/response examples for error cases in API documentation.
- Add integration tests with PostgreSQL, especially for concurrent authorization and reversal scenarios.
- Add OpenAPI/Swagger documentation.
- Add authentication and authorization if the API is exposed outside a trusted environment.
- Add pagination for transaction history.
- Add structured validation for supported currencies.
- Add a migration runner or startup command in Docker Compose.
- Avoid copying `.env` into the Docker image; pass configuration only through environment variables or secrets.
- Add CI checks for tests, formatting, linting, and migrations.

## License

This project is licensed under the terms in the `LICENSE` file.
