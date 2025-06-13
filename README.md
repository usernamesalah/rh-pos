# POS System API

A complete Point of Sale (POS) system backend API built with Go, following Domain-Driven Design (DDD) and Clean Architecture principles.

## 🚀 Features

- **Authentication & Authorization**: JWT-based authentication with role-based access control
- **Product Management**: CRUD operations for products with stock management
- **Transaction Processing**: Create and manage sales transactions with **ACID-compliant database transactions**
- **Sales Reporting**: Generate comprehensive sales reports with date range filtering
- **RESTful API**: Well-structured REST endpoints with proper HTTP status codes
- **API Documentation**: Auto-generated Swagger/OpenAPI documentation
- **Database Migrations**: Versioned database schema management with Goose
- **Hot Reload**: Development server with automatic reload using Air
- **Docker Support**: Full containerization with Docker and Docker Compose
- **Structured Logging**: JSON-structured logging with slog
- **Input Validation**: Request validation with proper error handling
- **Environment Configuration**: Flexible configuration via `.env` files
- **CLI Interface**: Command-line interface for database management

## 🔒 Database Transaction Guarantees

This application ensures **ACID compliance** for all critical operations:

### Transaction Processing
- **Atomicity**: Stock updates and transaction creation happen in a single database transaction
- **Consistency**: Stock levels are validated before processing to prevent overselling
- **Isolation**: Concurrent transactions don't interfere with each other
- **Durability**: All changes are persisted before confirmation

### Example Transaction Flow
```go
// All operations within a single database transaction
db.Transaction(func(tx *gorm.DB) error {
    // 1. Validate product availability and stock
    // 2. Create transaction record
    // 3. Create transaction items
    // 4. Update product stock levels
    // If ANY step fails, ALL changes are rolled back
    return nil
})
```

This ensures that:
- Stock is never oversold
- Transactions are never partially created
- Data integrity is maintained under high concurrency

## 🏗️ Architecture

This project follows **Clean Architecture** and **Domain-Driven Design** principles:

```
cmd/                    # Application entry points
├── main.go            # Main server application
└── cli/               # CLI commands
    ├── root.go        # Root command
    ├── migrate.go     # Migration commands
    └── seed.go        # Seed command

internal/              # Private application code
├── domain/            # Domain layer (entities & interfaces)
│   ├── entities/      # Domain entities
│   └── interfaces/    # Repository & service interfaces
├── usecase/           # Use case layer (business logic)
├── repository/        # Repository layer (data access)
├── handler/           # Handler layer (HTTP controllers)
├── server/            # HTTP server setup
├── config/            # Configuration management
└── pkg/               # Shared packages
    └── database/      # Database utilities

migrations/            # Database migrations
api/                   # API documentation
docs/                  # Generated Swagger docs (auto-generated)
```

## 🛠️ Tech Stack

- **Language**: Go 1.24
- **Web Framework**: Echo v4
- **Database**: MySQL 8.0
- **ORM**: GORM
- **Authentication**: JWT with bcrypt
- **Logging**: slog (structured logging)
- **Validation**: go-playground/validator
- **Documentation**: Swagger/OpenAPI 3.0
- **Migrations**: Goose
- **Hot Reload**: Air
- **Containerization**: Docker & Docker Compose
- **CLI Framework**: Cobra

## 📋 Prerequisites

- Go 1.24 or higher
- MySQL 8.0 or higher
- Docker and Docker Compose (for containerized setup)
- Make (optional, for using Makefile commands)

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd rh-pos
```

### 2. Setup Development Environment

```bash
# Install dependencies and development tools
make setup
```

### 3. Environment Configuration

**Important**: Create a `.env` file from the provided template:

```bash
# Copy the example environment file
cp env.example .env
```

**Edit the `.env` file** with your specific configuration:

```bash
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_secure_password
DB_NAME=rh_pos

# JWT Configuration (MUST CHANGE IN PRODUCTION!)
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Logger Configuration
LOG_LEVEL=info
```

### ⚠️ Security Notes for Production:

1. **Change JWT_SECRET**: Use a cryptographically secure random string
2. **Strong Database Password**: Use a complex password for DB_PASSWORD
3. **Environment Variables**: Consider using environment variables instead of .env files in production
4. **File Permissions**: Ensure .env file has restricted permissions (600)

```bash
# Set secure permissions for .env file
chmod 600 .env
```

### 4. Database Setup

#### Option A: Using Docker (Recommended)

```bash
# Start MySQL container (reads from .env file automatically)
docker-compose up mysql -d

# Wait for MySQL to be ready (about 10-15 seconds)
# Check with: docker-compose logs mysql

# Run migrations
make migrate-up

# Seed initial data
make seed
```

#### Option B: Manual MySQL Setup

```bash
# Create database
mysql -u root -p -e "CREATE DATABASE rh_pos;"

# Run migrations
make migrate-up

# Seed initial data
make seed
```

### 5. Run the Application

#### Development Mode (with Hot Reload)

```bash
make run
# or
air
```

#### Production Mode

```bash
make build
./bin/rh-pos
```

#### Docker Mode

```bash
# Build and run everything with Docker (reads .env automatically)
make docker-up

# Or just the application (if you have MySQL running separately)
make docker-build
docker run -p 8080:8080 --env-file .env rh-pos
```

The API will be available at `http://localhost:8080`

## 📚 CLI Commands

The application provides a command-line interface for database management:

```bash
# Build the CLI
make build

# Run migrations
./bin/rh-pos migrate up    # Run all pending migrations
./bin/rh-pos migrate down  # Rollback the last migration
./bin/rh-pos migrate status # Show migration status

# Seed the database
./bin/rh-pos seed
```

## 📚 API Documentation

### Interactive API Documentation

Once the server is running, visit:
- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **Health Check**: `http://localhost:8080/health`

### Authentication

All endpoints except `/auth/login` and `/health` require authentication.

**Login**:
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "username": "admin",
  "role": "admin"
}
```

**Using the Token**:
```bash
curl -H "Authorization: Bearer <your-jwt-token>" \
  http://localhost:8080/profile
```

### API Endpoints

#### Authentication
- `POST /auth/login` - User login
- `GET /profile` - Get current user profile

#### Products
- `GET /products` - List products (with pagination)
- `GET /products/{id}` - Get product by ID
- `PUT /products/{id}` - Update product
- `PUT /products/{id}/stock` - Update product stock

#### Transactions
- `GET /transactions` - List transactions (with pagination)
- `POST /transactions` - Create new transaction
- `GET /transactions/{id}` - Get transaction by ID

#### Reports
- `GET /reports` - Get sales report (with date range)

### Example Requests

**Create Transaction**:
```bash
curl -X POST http://localhost:8080/transactions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"product_id": 1, "quantity": 2},
      {"product_id": 2, "quantity": 1}
    ],
    "user": "admin",
    "payment_method": "cash",
    "discount": 0,
    "total_price": 44000
  }'
```

**Get Sales Report**:
```bash
curl "http://localhost:8080/reports?start_date=2024-01-01&end_date=2024-12-31" \
  -H "Authorization: Bearer <token>"
```

## 🗄️ Database Schema

### Users
- `id` (Primary Key)
- `username` (Unique)
- `password` (Hashed with bcrypt)
- `role` (admin/user)
- `created_at`, `updated_at`

### Products
- `id` (Primary Key)
- `image` (URL/path)
- `name`
- `sku` (Unique)
- `harga_modal` (Cost price)
- `harga_jual` (Selling price)
- `stock`
- `created_at`, `updated_at`

### Transactions
- `id` (Primary Key)
- `user`
- `payment_method`
- `discount`
- `total_price`
- `created_at`, `updated_at`

### Transaction Items
- `id` (Primary Key)
- `transaction_id` (Foreign Key)
- `product_id` (Foreign Key)
- `quantity`
- `price`
- `created_at`, `updated_at`

## 🔧 Development

### Available Make Commands

```bash
make help           # Show all available commands
make run            # Run with hot reload
make build          # Build binary
make test           # Run tests
make test-coverage  # Run tests with coverage
make deps           # Download dependencies
make fmt            # Format code
make lint           # Run linter
make clean          # Clean build artifacts

# Database
make migrate-up     # Run migrations
make migrate-down   # Rollback migrations
make migrate-status # Show migration status
make seed           # Seed initial data

# Docker
make docker-build   # Build Docker image
make docker-up      # Start all services
make docker-down    # Stop all services

# Development
make dev-up         # Start MySQL + App with hot reload
make dev-down       # Stop development environment

# Documentation
make swagger-gen    # Generate Swagger docs
```

### Hot Reload Development

The project uses [Air](https://github.com/cosmtrek/air) for hot reload during development:

```bash
# Install Air if not already installed
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
# or
make run
```

Air is configured via `air.toml` to watch for changes in `.go` files and automatically rebuild and restart the server.

### Database Migrations

Create a new migration:
```bash
make migrate-create NAME=add_new_table
```

Run migrations:
```bash
# Set your database DSN
export DB_DSN="root:password@tcp(localhost:3306)/rh_pos?charset=utf8mb4&parseTime=True&loc=Local"

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# View coverage report
open coverage.html
```

## 🐳 Docker Deployment

### Development with Docker Compose

```bash
# Start all services (MySQL + App)
make docker-up

# View logs
make docker-logs

# Stop all services
make docker-down
```

### Production Deployment

1. **Build the image**:
```bash
docker build -t rh-pos .
```

2. **Run with environment variables**:
```bash
docker run -d \
  --name rh-pos \
  -p 8080:8080 \
  -e DB_HOST=your-mysql-host \
  -e DB_USER=your-db-user \
  -e DB_PASSWORD=your-db-password \
  -e DB_NAME=rh_pos \
  -e JWT_SECRET=your-production-secret \
  rh-pos
```

3. **Or use Docker Compose with production overrides**:
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## 🔐 Security Considerations

- **JWT Secrets**: Change the default JWT secret in production
- **Database**: Use strong database passwords
- **HTTPS**: Always use HTTPS in production
- **Environment Variables**: Never commit sensitive data to version control
- **User Passwords**: Passwords are hashed using bcrypt with cost 10
- **Input Validation**: All inputs are validated before processing
- **SQL Injection**: Protected by GORM's prepared statements

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## ✨ Default Credentials

After running `make seed`, you can login with:
- **Username**: `admin`
- **Password**: `admin123`

## 🐛 Troubleshooting

### Common Issues

1. **Database Connection Failed**:
   - Ensure MySQL is running
   - Check database credentials in `.env`
   - Verify database exists

2. **Port Already in Use**:
   - Change `SERVER_PORT` in `.env`
   - Kill process using port 8080: `lsof -ti:8080 | xargs kill`

3. **Migration Errors**:
   - Ensure database exists
   - Check database permissions
   - Verify migration files syntax

4. **JWT Token Invalid**:
   - Check JWT secret configuration
   - Ensure token format: `Bearer <token>`
   - Verify token hasn't expired (24 hours)

### Getting Help

1. Check the logs: `docker-compose logs` or `air` output
2. Verify configuration: Ensure all environment variables are set
3. Database status: Check if MySQL is accessible
4. API documentation: Visit `/swagger/` for endpoint details

---

Built with ❤️ using Go and Clean Architecture principles. 