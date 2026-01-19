# Docker Setup Guide for VIGILUM

## Quick Start

### Prerequisites
- Docker Desktop installed (or Docker Engine + Docker Compose)
- ~4GB free disk space
- Git clone of VIGILUM

### Build and Run

**1. Build the Docker image:**
```bash
docker-compose build
```

**2. Start all services:**
```bash
docker-compose up -d
```

**3. Verify services are running:**
```bash
docker-compose ps
```

### Access Services

| Service | URL | Default Credentials |
|---------|-----|-------------------|
| Backend API | http://localhost:8080 | - |
| Temporal UI | http://localhost:8081 | - |
| Prometheus | http://localhost:9090 | - |
| Grafana | http://localhost:3000 | admin/admin |
| Jaeger Tracing | http://localhost:16686 | - |
| NATS | localhost:4222 | - |
| Redis | localhost:6379 | - |
| PostgreSQL | localhost:5432 | vigilum/vigilum |
| Qdrant | http://localhost:6333 | - |
| ClickHouse | http://localhost:8123 | - |

### Database

PostgreSQL is initialized with:
- **User**: vigilum
- **Password**: vigilum
- **Database**: vigilum

Connect via: `psql -h localhost -U vigilum -d vigilum`

### Stop Services

```bash
docker-compose down
```

To also remove data volumes:
```bash
docker-compose down -v
```

### View Logs

All services:
```bash
docker-compose logs -f
```

Specific service:
```bash
docker-compose logs -f backend
docker-compose logs -f postgres
docker-compose logs -f redis
```

### Share with Others

1. **Push to Docker Hub** (optional):
```bash
docker build -t your-username/vigilum:latest .
docker push your-username/vigilum:latest
```

2. **Export image**:
```bash
docker save vigilum:latest > vigilum.tar.gz
```
Others can load it with:
```bash
docker load < vigilum.tar.gz
```

3. **Or share the repository**:
Just clone and run `docker-compose up` - it will build automatically!

### Environment Variables

Edit `docker-compose.yml` to customize:
- Database credentials
- Service endpoints
- Log levels
- Port mappings

### Troubleshooting

**Port already in use:**
```bash
# Change ports in docker-compose.yml or use:
docker-compose down
```

**Out of disk space:**
```bash
docker system prune -a
```

**Container not starting:**
```bash
docker-compose logs backend
docker inspect vigilum-api
```

**Reset everything:**
```bash
docker-compose down -v
docker system prune -a
docker-compose build --no-cache
docker-compose up -d
```

### Performance Tips

1. **Allocate more resources** to Docker Desktop (Settings → Resources)
2. **Use `.dockerignore`** to exclude unnecessary files
3. **Enable BuildKit** for faster builds:
```bash
export DOCKER_BUILDKIT=1
```

### Health Checks

Services include health checks. Check status:
```bash
docker-compose ps
```

The backend service includes a `/health` endpoint:
```bash
curl http://localhost:8080/health
```

### For CI/CD

The Dockerfile supports multi-stage builds for small production images (~50MB).
All services auto-restart on failure (`restart: unless-stopped`).
