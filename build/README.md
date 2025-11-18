# Docker Setup

This directory contains Docker configuration files for the booking service.

## Files

- `Dockerfile` - Multi-stage Docker build for the Go application
- `docker-compose.yml` - Docker Compose configuration with PostgreSQL, Redis, and the service
- `config.yaml` - Example configuration file for Docker environment
- `.dockerignore` - Files to exclude from Docker build context

## Quick Start

1. **Build and start all services:**
   ```bash
   cd build
   docker-compose up -d
   ```

2. **Run database migrations:**
   ```bash
   docker-compose exec booking-svc ./booking-svc service migrate up --config /app/config/config.yaml
   ```

3. **View logs:**
   ```bash
   docker-compose logs -f booking-svc
   ```

4. **Stop all services:**
   ```bash
   docker-compose down
   ```

5. **Stop and remove volumes (clean slate):**
   ```bash
   docker-compose down -v
   ```

## Configuration

The service uses the `config.yaml` file in the build directory. You can modify it or mount your own configuration file:

```yaml
volumes:
  - /path/to/your/config.yaml:/app/config/config.yaml:ro
```

## Environment Variables

You can override configuration using environment variables or by mounting a custom config file.

## Services

- **booking-svc** - Main application service (port 8080)
- **postgres** - PostgreSQL database (port 5432)
- **redis** - Redis cache (port 6379)

## Development

To rebuild the service after code changes:

```bash
docker-compose build booking-svc
docker-compose up -d booking-svc
```

## Production Considerations

1. Change the JWT secret key in `config.yaml`
2. Use strong database passwords
3. Configure proper SSL/TLS for database connections
4. Set up proper logging and monitoring
5. Use secrets management for sensitive configuration
6. Configure resource limits in docker-compose.yml


