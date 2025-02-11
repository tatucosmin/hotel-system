# ticketr

ticketr is a ticket management system that allows users to create, update, and manage tickets. It includes user authentication, role-based permissions, and the ability to store closed tickets to S3.

## Features

- User authentication with JWT
- Role-based permissions (Admin, Staff, Customer)
- Create, update, and manage tickets
- Store closed tickets to S3
- Middleware for logging and authentication
- RESTful API endpoints

## Getting Started

### Prerequisites

- Go 1.23.1 or later
- Docker
- Docker Compose
- PostgreSQL

### Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/tatucosmin/ticketr.git
    cd ticketr
    ```

`NOTE: Some parts of the config are highly dependent on the ENV variable inside .envrc`

2. Copy the example environment file and update the variables inside of it:
    ```sh
    cp .envrc.example .envrc
    ```

3. Start the services using Docker Compose:
    ```sh
    docker-compose up -d
    ```

4. Run the database migrations:
    ```sh
    make db_migrate_up
    ```

5. Run terraform and 

5. Start the server:
    ```sh
    make run
    ```

### Usage

The server will be running on the host and port specified in the [.envrc](http://_vscodecontentref_/26) file. You can interact with the API using tools like `curl`, Postman, or HTTPie.

### API Endpoints

Public routes:
- `POST /api/auth/signup` - Sign up a new user
- `POST /api/auth/signin` - Sign in an existing user
- `POST /api/auth/refresh` - Refresh access token

Auth routes:
- `GET /ping` - Health check endpoint
- `GET /api/ticket` - Get a specific ticket
- `POST /api/ticket` - Create a new ticket
- `PUT /api/ticket` - Update an existing ticket

Admin only:
- `GET /api/tickets` - Get all tickets (Admin only)
