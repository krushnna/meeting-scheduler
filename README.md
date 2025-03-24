
# Meeting Scheduler API

An API for scheduling meetings across time zones. The application helps an event organizer find optimal meeting slots by comparing the provided available times of all participants. The system recommends time slots that work for everyone or those that suit the most participants.

## Features

- **Event Management:**  
  Create, update, retrieve, and delete events.

- **Time Slot Management:**  
  Create, update, retrieve, and delete proposed time slots for events.

- **User & Availability:**  
  Register users and record their availability for each event.

- **Time Slot Recommendations:**  
  Automatically recommend meeting slots based on participant availability.

- **RESTful API:**  
  Built using Go, adhering to REST conventions.

- **Swagger Documentation:**  
  Interactive API docs available via Swagger UI.

- **Docker & Docker Compose:**  
  Containerized application for simplified deployment.

## Project Structure

```
.
├── controllers        # API endpoint handlers 
├── docs               # Swagger/OpenAPI documentation (manually maintained openapi.yaml)
├── initializers       # Database and other application initialization code
├── models             # Database models and input/output data structures
├── repository         # Data access layer for CRUD operations
├── routers            # Gin router setup and route definitions
├── services           # Business logic for events, timeslots, users, and recommendations
├── .env               # Environment variables
├── Dockerfile         # Docker build instructions for production image
├── docker-compose.yml # Docker Compose configuration for app + PostgreSQL
├── go.mod             # Go modules file
└── main.go            # Application entrypoint: loads environment, initializes DB and router, then starts the server.
```

## Getting Started

### Prerequisites

- **Go:** Version 1.24.1 or later.
- **Docker & Docker Compose:** For containerized deployment.
- **PostgreSQL:** The app uses PostgreSQL as its database (Docker Compose runs a container for you).

### Environment Variables

Create a `.env` file at the root of your project with the following variables (adjust as needed):

```dotenv
# Environment
APP_ENV=development  # production/staging
GIN_MODE=debug
LOG_LEVEL=info
# Server
API_PORT=8080

# Database
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=meetingscheduler
DB_SSLMODE=disable

# Timezone (for database connections)
TZ=UTC

```

### Running Locally (Without Docker)

1. **Install Dependencies:**

   ```bash
   go mod download
   ```

2. **Set Up PostgreSQL:**  
   Make sure PostgreSQL is running and accessible based on the connection details in your `.env` file.

3. **Run the Application:**

   ```bash
   go run main.go
   ```

4. **Access the API:**
   - Health check: [http://localhost:8080/health](http://localhost:8080/health)
   - Swagger UI: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)  
     The Swagger UI automatically loads the API spec from `/docs/openapi.yaml`.

### Running with Docker

#### Dockerfile

The provided `Dockerfile` builds your Go application in a multi-stage process and copies your manually maintained API docs into the final image.

```dockerfile
# Use Golang Alpine as builder
FROM golang:1.24.1-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to leverage caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# (Since you are manually maintaining openapi.yaml, we don't run swag init here)
# RUN go install github.com/swaggo/swag/cmd/swag@latest
# RUN swag init -g main.go -o docs

# Build the application binary
RUN CGO_ENABLED=0 GOOS=linux go build -o meeting-scheduler .

# Use a minimal alpine image for production
FROM alpine:3.18

# Install necessary dependencies
RUN apk --no-cache add ca-certificates tzdata bash

WORKDIR /root/

# Copy the compiled binary, .env file, and docs folder from builder stage
COPY --from=builder /app/meeting-scheduler .
COPY --from=builder /app/.env .
COPY --from=builder /app/docs ./docs

# Expose the application port
EXPOSE 8080

# Set the default command
CMD ["./meeting-scheduler"]
```

#### Docker Compose

A sample `docker-compose.yml` to run both the API and PostgreSQL:

```yaml
version: '3.8'
services:
  db:
    image: postgres:14-alpine
    environment:
      POSTGRES_DB: meetingscheduler
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  app:
    build: .
    depends_on:
      - db
    environment:
      APP_ENV: development
      GIN_MODE: debug
      API_PORT: "8080"
      DB_HOST: db
      DB_PORT: "5432"
      DB_USER: postgres
      DB_PASSWORD: postgres
      DB_NAME: meetingscheduler
      DB_SSLMODE: disable
      TZ: UTC
    ports:
      - "8080:8080"

volumes:
  postgres_data:
```

#### Build & Run

From the project root, run:

```bash
docker-compose up --build
```

- **API is available at:** [http://localhost:8080](http://localhost:8080)
- **Swagger UI:** [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

### Accessing the OpenAPI Specification

Your manually maintained API specification is stored at `docs/openapi.yaml`. Swagger UI is configured to load it automatically by setting the URL option. If needed, you can also view it directly at:

```
http://localhost:8080/docs/openapi.yaml
```

### Inspecting the Database

To verify data in PostgreSQL:

1. **List running containers:**

   ```bash
   docker ps
   ```

2. **Connect to the PostgreSQL container:**

   ```bash
   docker exec -it <db_container_name> psql -U postgres -d meetingscheduler
   ```

3. **List tables and query data:**

   ```sql
   \dt
   SELECT * FROM events;
   SELECT * FROM users;
   SELECT * FROM time_slots;
   SELECT * FROM user_availabilities;
   ```

4. **Exit psql:**

   ```sql
   \q
   ```

## API Endpoints Overview

The API follows a RESTful pattern. Here are some key endpoints:

### Events

- `POST /api/v1/events` – Create a new event.
- `GET /api/v1/events` – Retrieve all events.
- `GET /api/v1/events/{id}` – Retrieve a specific event.
- `PUT /api/v1/events/{id}` – Update an event.
- `DELETE /api/v1/events/{id}` – Delete an event.
- `GET /api/v1/events/{id}/recommendations` – Get recommended time slots for an event.

### Time Slots

- `POST /api/v1/events/{id}/timeslots` – Create a time slot for an event.
- `GET /api/v1/events/{id}/timeslots` – Retrieve all time slots for an event.
- `PUT /api/v1/events/{id}/timeslots/{slotId}` – Update a time slot.
- `DELETE /api/v1/events/{id}/timeslots/{slotId}` – Delete a time slot.

### Users

- `POST /api/v1/users` – Create a new user.
- `GET /api/v1/users` – Retrieve all users.
- `GET /api/v1/users/{id}` – Retrieve a specific user.
- `PUT /api/v1/users/{id}` – Update a user.
- `DELETE /api/v1/users/{id}` – Delete a user.
- `POST /api/v1/users/{id}/events/{eventId}/availability` – Record a user's availability for an event.
- `GET /api/v1/users/{id}/events/{eventId}/availability` – Retrieve a user's availability for an event.

For more details on each endpoint, visit the Swagger UI.




