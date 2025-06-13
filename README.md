# POS System API

A complete Point of Sale (POS) system backend API built with Go and Echo framework.

## 🚀 Features

- **RESTful API**: Well-structured REST endpoints with proper HTTP status codes
- **API Documentation**: Auto-generated Swagger/OpenAPI documentation
- **Hot Reload**: Development server with automatic reload using Air
- **Docker Support**: Full containerization with Docker and Docker Compose
- **Structured Logging**: JSON-structured logging with slog
- **Input Validation**: Request validation with proper error handling
- **Environment Configuration**: Flexible configuration via `.env` files

## 🏗️ Architecture

This project follows **Clean Architecture** principles:

```
cmd/                    # Application entry points
└── main.go            # Main server application

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
```

## 🛠️ Tech Stack

- **Language**: Go 1.24
- **Web Framework**: Echo v4
- **Database**: MySQL 8.0
- **ORM**: GORM
- **Logging**: slog (structured logging)
- **Validation**: go-playground/validator
- **Documentation**: Swagger/OpenAPI 3.0
- **Hot Reload**: Air
- **Containerization**: Docker & Docker Compose

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

# Logger Configuration
LOG_LEVEL=info
```

### 4. Run the Application

#### Development Mode (with Hot Reload)

```bash
make dev
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

## 📚 API Documentation

### Interactive API Documentation

Once the server is running, visit:
- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **Health Check**: `http://localhost:8080/health`

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
  rh-pos
```

3. **Or use Docker Compose with production overrides**:
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## 🔐 Security Considerations

- **Database**: Use strong database passwords
- **HTTPS**: Always use HTTPS in production
- **Environment Variables**: Never commit sensitive data to version control
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

## 🐛 Troubleshooting

### Common Issues

1. **Database Connection Failed**:
   - Ensure MySQL is running
   - Check database credentials in `.env`
   - Verify database exists

2. **Port Already in Use**:
   - Change `SERVER_PORT` in `.env`
   - Kill process using port 8080: `lsof -ti:8080 | xargs kill`

### Getting Help

1. Check the logs: `docker-compose logs` or `air` output
2. Verify configuration: Ensure all environment variables are set
3. Database status: Check if MySQL is accessible 