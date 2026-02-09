# Chirpy API 🐦

Chirpy is a simple REST API for a mini social network where users can
create, read, delete and manage **chirps** (posts), using **JWT +
Refresh Token** authentication, with support for premium subscriptions
and external webhooks.

------------------------------------------------------------------------

## Table of Contents

-   Features
-   Tech Stack
-   Installation
-   Configuration
-   Running the Server
-   Authentication
-   API Endpoints
    -   Healthcheck
    -   Users
    -   Login & Tokens
    -   Chirps
    -   Admin
    -   Polka Webhooks
-   Error Codes
-   Security

------------------------------------------------------------------------

## Features

-   ✅ User registration
-   ✅ Login with JWT
-   ✅ Token refresh & revoke
-   ✅ Full CRUD for chirps
-   ✅ Chirp filtering & sorting
-   ✅ Chirp deletion by owner only
-   ✅ User upgrade via external webhook (Polka)
-   ✅ Admin metrics & reset
-   ✅ File server visit counter middleware

------------------------------------------------------------------------

## Tech Stack

-   Go (net/http)
-   PostgreSQL
-   SQLC
-   JWT Authentication
-   Atomic counters
-   Webhooks
-   dotenv

------------------------------------------------------------------------

## Installation

``` bash
git clone https://github.com/andrei-himself/chirpy.git
cd chirpy
go mod download
```

------------------------------------------------------------------------

## Configuration

Create a `.env` file in the project root:

``` env
DB_URL=postgres://user:password@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
SECRET=super-secret-key
POLKA_KEY=polka-secret-key
```

------------------------------------------------------------------------

## Running the Server

``` bash
go run main.go
```

The server will start on:

    http://localhost:8080

------------------------------------------------------------------------

## Authentication

All protected routes require a Bearer token:

    Authorization: Bearer <TOKEN>

Chirpy uses: - **Access Token (JWT)** -- short-lived authentication -
**Refresh Token** -- used to obtain a new access token

------------------------------------------------------------------------

# API Endpoints

## Healthcheck

### `GET /api/healthz`

Returns server status.

**Response:**

    200 OK
    OK

------------------------------------------------------------------------

## Users

### Create User

### `POST /api/users`

``` json
{
  "email": "test@test.com",
  "password": "password123"
}
```

**Response: 201 Created**

``` json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "email": "test@test.com",
  "is_chirpy_red": false
}
```

------------------------------------------------------------------------

### Update User (email & password)

### `PUT /api/users`

**Authorization required**

``` json
{
  "email": "new@email.com",
  "password": "newPassword"
}
```

------------------------------------------------------------------------

## Login & Tokens

### Login

### `POST /api/login`

``` json
{
  "email": "test@test.com",
  "password": "password123"
}
```

**Response:**

``` json
{
  "id": "uuid",
  "email": "test@test.com",
  "token": "JWT_TOKEN",
  "refresh_token": "REFRESH_TOKEN",
  "is_chirpy_red": false
}
```

------------------------------------------------------------------------

### Refresh Token

### `POST /api/refresh`

**Header:**

    Authorization: Bearer <REFRESH_TOKEN>

**Response:**

``` json
{
  "token": "NEW_JWT_TOKEN"
}
```

------------------------------------------------------------------------

### Revoke Refresh Token

### `POST /api/revoke`

**Header:**

    Authorization: Bearer <REFRESH_TOKEN>

**Response:**

    204 No Content

------------------------------------------------------------------------

## Chirps

### Create Chirp

### `POST /api/chirps`

**Authorization required**

``` json
{
  "body": "Hello, this is my first chirp!"
}
```

Rules: - Max 140 characters - Censored words: `kerfuffle`, `sharbert`,
`fornax`

------------------------------------------------------------------------

### Get All Chirps

### `GET /api/chirps`

Optional query params: - `author_id=<uuid>` -- filter by author -
`sort=desc` -- sort by newest first

------------------------------------------------------------------------

### Get Single Chirp

### `GET /api/chirps/{chirpID}`

------------------------------------------------------------------------

### Delete Chirp

### `DELETE /api/chirps/{chirpID}`

Only the chirp author can delete it.

------------------------------------------------------------------------

## Admin

### Metrics

### `GET /admin/metrics`

Returns how many times the file server was accessed.

------------------------------------------------------------------------

### Reset Database

### `POST /admin/reset`

⚠️ Available only when:

    PLATFORM=dev

Deletes all users and chirps from the database.

------------------------------------------------------------------------

## Polka Webhooks

### `POST /api/polka/webhooks`

**Header:**

    Authorization: ApiKey <POLKA_KEY>

``` json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "UUID"
  }
}
```

Upgrades the user to **Chirpy Red**.

------------------------------------------------------------------------

## Error Codes

  Code   Meaning
  ------ -----------------------
  200    OK
  201    Created
  204    No Content
  400    Bad Request
  401    Unauthorized
  403    Forbidden
  404    Not Found
  500    Internal Server Error

------------------------------------------------------------------------

## Security

-   Passwords are securely **hashed**
-   JWT tokens are signed using `SECRET`
-   Refresh tokens are stored in the database
-   Tokens can be revoked at any time
-   Admin reset is protected by `PLATFORM` environment variable

------------------------------------------------------------------------

## Author

Created by **Andrei Himself**\
Educational REST API project built with Go.
