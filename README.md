# LinkLens

A full-stack URL shortener with analytics built using Go, Gin, PostgreSQL, Redis, JWT authentication, and Docker.

## Features

* User registration and login
* JWT authentication with refresh tokens
* URL shortening with custom short codes
* Click tracking and analytics
* Redis caching
* Rate limiting
* PostgreSQL persistence
* Dockerized deployment

## Tech Stack

### Backend

* Go
* Gin

### Database

* PostgreSQL

### Cache

* Redis

### Authentication

* JWT

### Deployment

* Docker
* Docker Compose

## Project Structure

```text
internal/
├── config/
├── db/
├── handlers/
├── middleware/
├── models/
└── repository/

migrations/
web/static/
```

## Running Locally

```bash
docker compose up --build
```

Application:

```text
http://localhost:8080
```

## API Endpoints

### Authentication

* POST `/api/auth/register`
* POST `/api/auth/login`
* POST `/api/auth/refresh`

### URLs

* POST `/api/urls`
* GET `/api/urls`
* GET `/api/urls/:id`
* PUT `/api/urls/:id`
* DELETE `/api/urls/:id`

### Analytics

* GET `/api/analytics/:id`
* GET `/api/analytics/:id/timeline`

## Future Improvements

* Custom domains
* QR code generation
* Geo-location analytics
* Admin dashboard
* CI/CD pipeline
* HTTPS support

## Author

Tejdeep Kodati
