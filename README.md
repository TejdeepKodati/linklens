# LinkLens

A production-ready URL shortener and analytics platform built with Go, Gin, PostgreSQL, Redis, JWT authentication, Docker, and AWS EC2.

LinkLens allows users to create custom short URLs, track clicks, monitor analytics, and manage links through a modern dashboard.

---

## Features

### Authentication

* User registration and login
* JWT access token authentication
* Refresh token support
* Protected API routes

### URL Management

* Create shortened URLs
* Custom short codes
* URL expiration support
* Activate / deactivate links
* Delete links

### Analytics

* Total clicks tracking
* Unique visitor tracking
* Browser analytics
* Device analytics
* Referrer analytics
* Timeline-based click statistics

### Performance

* Redis caching for fast redirects
* Rate limiting middleware
* Optimized PostgreSQL queries

### Deployment

* Dockerized application
* Multi-container architecture
* AWS EC2 deployment
* Docker Compose orchestration

---

## Tech Stack

### Backend

* Go
* Gin

### Database

* PostgreSQL

### Cache

* Redis

### Authentication

* JWT (JSON Web Tokens)

### Infrastructure & Deployment

* Docker
* Docker Compose
* AWS EC2

### Frontend

* HTML
* CSS
* JavaScript

---

## Architecture

```text
┌─────────────────────┐
│      Frontend       │
│    HTML / CSS / JS  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│      Gin API        │
│   Go Backend App    │
└──────────┬──────────┘
           │
     ┌─────┴─────┐
     ▼           ▼
┌─────────┐ ┌─────────┐
│  Redis  │ │Postgres │
│  Cache  │ │Database │
└─────────┘ └─────────┘
```

---

## Project Structure

```text
linklens/
│
├── internal/
│   ├── config/
│   ├── db/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   └── repository/
│
├── migrations/
│
├── web/
│   └── static/
│
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── main.go
```

---

## API Endpoints

### Authentication

| Method | Endpoint           |
| ------ | ------------------ |
| POST   | /api/auth/register |
| POST   | /api/auth/login    |
| POST   | /api/auth/refresh  |

### URL Management

| Method | Endpoint      |
| ------ | ------------- |
| POST   | /api/urls     |
| GET    | /api/urls     |
| GET    | /api/urls/:id |
| PUT    | /api/urls/:id |
| DELETE | /api/urls/:id |

### Analytics

| Method | Endpoint                    |
| ------ | --------------------------- |
| GET    | /api/analytics/:id          |
| GET    | /api/analytics/:id/timeline |

### Redirect

| Method | Endpoint    |
| ------ | ----------- |
| GET    | /:shortCode |

---

## Running Locally

### Prerequisites

* Docker
* Docker Compose

### Clone Repository

```bash
git clone https://github.com/TejdeepKodati/linklens.git
cd linklens
```

### Start Application

```bash
docker compose up --build
```

Application will be available at:

```text
http://localhost:8080
```

---

## Environment Variables

```env
ENV=development
PORT=8080

DATABASE_URL=postgres://postgres:postgres@localhost:5432/linklens?sslmode=disable

REDIS_URL=redis://localhost:6379

JWT_SECRET=your-secret-key

BASE_URL=http://localhost:8080
```

---

## Deployment

LinkLens is deployed using Docker containers on AWS EC2.

### Services

* Go/Gin Application
* PostgreSQL Database
* Redis Cache

All services run in isolated Docker containers and communicate through a private Docker network managed by Docker Compose.

Deployment workflow:

```text
GitHub
   ↓
AWS EC2
   ↓
Docker Compose
   ↓
Go App + PostgreSQL + Redis
```

---

## Screenshots

### Login

![Registration](screenshots/registration_page.png)

### Interface

![Interface](screenshots/interface.png)

### URL Creation

![URL Creation](screenshots/shortened_url_copied.png)

### Analytics

![Analytics](screenshots/analytics.png)

---

## Highlights

* Built REST APIs using Go and Gin
* Implemented JWT authentication and refresh tokens
* Designed PostgreSQL data models and repositories
* Integrated Redis caching for high-performance redirects
* Implemented click analytics and tracking
* Added rate limiting middleware
* Containerized the application using Docker
* Deployed and hosted on AWS EC2
* Built a complete full-stack application from frontend to cloud deployment

---

## Future Improvements

* Custom domains
* HTTPS with Let's Encrypt
* QR code generation
* Geo-location analytics
* Admin dashboard
* GitHub Actions CI/CD pipeline
* Email verification
* Password reset support
* Link sharing statistics

---

## Author

### Tejdeep Kodati

* GitHub: https://github.com/TejdeepKodati

---

## License

This project is intended for educational and portfolio purposes.
