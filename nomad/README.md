# Nomad Jobs for Multi-Tenant PDF Ingestion Service

Complete Nomad job specifications for deploying your PDF summarization service to a Nomad cluster.

---

## 📁 What's Included

### Job Specifications (Ready to Deploy)

| File | Purpose | Replicas | Status |
|------|---------|----------|--------|
| `api.nomad.hcl` | Main API service | 2 | ✅ Ready |
| `mongodb.nomad.hcl` | Tenant databases | 1 | ✅ Ready |
| `postgres.nomad.hcl` | Master database | 1 | ✅ Ready |
| `minio.nomad.hcl` | Object storage | 1 | ✅ Ready |

### Documentation

| File | Purpose |
|------|---------|
| `DEPLOYMENT_GUIDE.md` | Step-by-step deployment instructions |
| `DOCKER_COMPOSE_TO_NOMAD_MAPPING.md` | Detailed comparison with docker-compose.yaml |
| `README.md` | This file |

### Setup Scripts

| File | Purpose |
|------|---------|
| `volumes-setup.sh` | Create required host volumes |

---

## 🚀 Quick Start (5 Minutes)

### Prerequisites

```bash
# Check installations
nomad --version
consul --version
docker --version

# Start Nomad (Terminal 1)
nomad agent -dev

# Start Consul (Terminal 2)
consul agent -dev

# Build Docker image (Terminal 3)
cd /home/bacancy/Desktop/droadmap
docker build -t droadmap-api:latest .
```

### Deploy (Terminal 3)

```bash
# Create volumes
cd nomad
chmod +x volumes-setup.sh
./volumes-setup.sh

# Deploy services
nomad job run postgres.nomad.hcl
nomad job run mongodb.nomad.hcl
nomad job run minio.nomad.hcl
nomad job run api.nomad.hcl

# Check status
nomad job status api
```

### Test

```bash
# API should be running
curl http://localhost:8080/health

# Upload a test PDF
curl -X POST http://localhost:8080/api/v1/upload \
  -F "tenantName=test" \
  -F "pdf=@../test-document-1.pdf"
```

---

## 📊 Architecture

```
┌─────────────────────────────────────────────┐
│          Nomad Cluster (dc1)                │
├─────────────────────────────────────────────┤
│                                              │
│  API Service (2 replicas)                   │
│  ├─ Instance 1: api.service.consul:8080    │
│  └─ Instance 2: api.service.consul:8080    │
│                                              │
│  Supporting Services (1 each)                │
│  ├─ PostgreSQL: postgres.service.consul     │
│  ├─ MongoDB: mongodb.service.consul         │
│  └─ MinIO: minio.service.consul             │
│                                              │
└─────────────────────────────────────────────┘
        ↓
┌─────────────────────────────────────────────┐
│    Consul (Service Discovery & DNS)         │
├─────────────────────────────────────────────┤
│  - Service Registry                         │
│  - Health Checks                            │
│  - KV Store (Tenant DB connections)         │
│  - DNS (*.service.consul)                   │
└─────────────────────────────────────────────┘
```

---

## 🔑 Key Features

### Service Discovery
- Automatic service registration with Consul
- DNS-based service discovery: `service-name.service.consul`
- No hardcoded IP addresses
- Services can move between nodes

### High Availability
- API replicas: Run 2-3 instances
- Automatic load balancing
- Health checks on all services
- Auto-restart on failure

### Resource Management
- Explicit CPU/memory allocation
- Nomad enforces resource limits
- Prevents node overload
- Easy scaling

### Data Persistence
- Named volumes for databases
- Data survives container restarts
- Location-independent storage

---

## 📖 Mapping from Docker Compose

All services are based on your existing `docker-compose.yaml`:

| Aspect | Docker Compose | Nomad |
|--------|---|---|
| **Configuration** | YAML | HCL |
| **Scaling** | Manual | Dynamic (count = N) |
| **Service Discovery** | Hardcoded DNS | Consul DNS |
| **Multi-node** | Not supported | Built-in |
| **Health Checks** | Built-in | Enhanced |

See `DOCKER_COMPOSE_TO_NOMAD_MAPPING.md` for detailed comparison.

---

## 🛠️ File Descriptions

### api.nomad.hcl

**Main API service with:**
- 2 replicas for load balancing (change `count = 2` to scale)
- Port 8080 mapped
- HTTP health check (/health)
- Graceful shutdown (30s kill timeout)
- Service registration with Consul

**Key Differences from Docker Compose:**
- Uses Consul DNS for database connections
- `postgres` → `postgres.service.consul`
- `mongodb` → `mongodb.service.consul`
- `minio:9000` → `minio.service.consul:9000`

### mongodb.nomad.hcl

**MongoDB service with:**
- Single replica (persistent data)
- Port 27017 mapped
- Two volumes: `/data/db` and `/data/configdb`
- Script-based health check (mongosh ping)
- Initialization script support

### postgres.nomad.hcl

**PostgreSQL service with:**
- Single replica (persistent data)
- Port 5432 mapped
- Volume: `/var/lib/postgresql/data`
- Script-based health check (pg_isready)
- Initialization SQL script support

### minio.nomad.hcl

**MinIO object storage with:**
- Single replica
- API port 9000 + Console port 9001
- Volume: `/data`
- HTTP health check
- Service mesh ready

---

## 📋 Common Tasks

### Scale API to 4 Replicas

```bash
# Edit api.nomad.hcl
# Change: count = 2
# To:     count = 4

nomad job run api.nomad.hcl
nomad job status api  # Shows 4/4 running
```

### View Service Logs

```bash
# Get allocation ID
nomad job status api

# View logs
nomad alloc logs <alloc_id>

# Follow logs
nomad alloc logs -follow <alloc_id>
```

### Stop All Services

```bash
nomad job stop api
nomad job stop mongodb
nomad job stop postgres
nomad job stop minio
```

### Update Database Password

1. Edit the `.hcl` file (e.g., `mongodb.nomad.hcl`)
2. Change the password in the `env` block
3. Run: `nomad job run mongodb.nomad.hcl`
4. Nomad will perform a rolling update

---

## 🔍 Monitoring

### Nomad Dashboard
- **URL:** http://localhost:4646
- **Shows:** Jobs, allocations, resource usage, logs

### Consul Dashboard
- **URL:** http://localhost:8500
- **Shows:** Services, health, KV store, DNS queries

### API Endpoints
```bash
# Health check
curl http://localhost:8080/health

# API status
nomad job status api

# Service discovery
curl http://localhost:8500/v1/catalog/service/api
```

---

## 🚨 Troubleshooting

### Services won't start

```bash
# Check Nomad logs
nomad server logs

# Check specific job logs
nomad alloc logs <alloc_id>

# Common issues:
# 1. Docker image not found: docker build -t droadmap-api:latest .
# 2. Volumes don't exist: ./volumes-setup.sh
# 3. Ports in use: lsof -i :8080
```

### Services not discoverable

```bash
# Check Consul
consul members

# Check service registration
consul catalog service api

# Test DNS
dig api.service.consul @127.0.0.1 -p 8600

# Check Consul DNS port
netstat -tln | grep 8600
```

### API can't connect to database

```bash
# Verify database is running
nomad job status mongodb

# Check API logs
nomad alloc logs <api_alloc_id> | grep -i "connection\|error"

# Test database connection from API container
nomad alloc exec <api_alloc_id> \
  curl mongodb.service.consul:27017
```

---

## 📝 Configuration Management

### Environment Variables

Edit the `env` block in any `.nomad.hcl` file:

```hcl
env {
  POSTGRES_DB = "master_db"
  GEMINI_API_KEY = "${GEMINI_API_KEY}"  # From shell environment
}
```

### Secrets in Consul KV

```bash
# Store API key in Consul
consul kv put config/gemini_api_key "your-actual-key"

# Reference in Nomad (for advanced setups)
# env {
#   GEMINI_API_KEY = data.consul.config.gemini_api_key
# }
```

### Tenant Database Connections

Your application stores connections in Consul KV:

```bash
consul kv put tenant/acme_corp/mongo_uri "mongodb://admin:pass@mongodb.service.consul:27017/acme_corp"
```

API code queries these dynamically.

---

## 🔄 Workflow

### Development Workflow

```
1. Make code changes
2. docker build -t droadmap-api:latest .
3. nomad job run api.nomad.hcl
4. Test: curl http://localhost:8080/health
5. View logs: nomad alloc logs -follow <alloc_id>
```

### Production Workflow

```
1. Build image: docker build -t droadmap-api:latest .
2. Push to registry: docker push my-registry/droadmap-api:latest
3. Update job spec: image = "my-registry/droadmap-api:latest"
4. Deploy: nomad job run api.nomad.hcl
5. Nomad handles rolling updates automatically
```

---

## 📚 Next Steps

### Now ✅
- Deploy locally
- Test all endpoints
- Verify data persistence

### Coming Soon ⬜
- Terraform for infrastructure provisioning
- Multi-node cluster setup
- CI/CD pipeline integration
- Monitoring (Prometheus + Grafana)
- Backup & recovery procedures

---

## 📖 Documentation

| Document | Content |
|----------|---------|
| `DEPLOYMENT_GUIDE.md` | Complete deployment walkthrough |
| `DOCKER_COMPOSE_TO_NOMAD_MAPPING.md` | Line-by-line comparison with docker-compose |
| `../NOMAD_REQUIREMENTS.md` | Architecture & design decisions |
| `../README.md` | Main project README |

---

## 🤔 FAQ

**Q: Can I run this without Consul?**
A: Nomad requires service discovery. Consul is recommended. Nomad also supports other service meshes.

**Q: How do I update API replicas from 2 to 3?**
A: Edit `api.nomad.hcl`, change `count = 3`, then run `nomad job run api.nomad.hcl`.

**Q: Will my data persist across restarts?**
A: Yes! Databases use named volumes that survive container restarts.

**Q: Can I use managed databases instead (AWS RDS)?**
A: Yes! Update connection strings in job specs. MongoDB and PostgreSQL don't need to run in Nomad.

**Q: How do I scale to multiple nodes?**
A: Deploy Nomad clients on additional machines, same job specs work across all nodes.

---

## 🆘 Getting Help

1. Check `DEPLOYMENT_GUIDE.md` for step-by-step instructions
2. Review `DOCKER_COMPOSE_TO_NOMAD_MAPPING.md` for specific service details
3. Check Nomad logs: `nomad server logs`
4. Check job logs: `nomad alloc logs <alloc_id>`
5. Refer to official docs:
   - Nomad: https://www.nomadproject.io/docs
   - Consul: https://www.consul.io/docs

---

## ✨ Summary

You now have production-ready Nomad job specifications that:

✅ Match your Docker Compose configuration exactly
✅ Support multi-node deployments
✅ Include automatic service discovery
✅ Provide health checks and auto-restart
✅ Enable easy scaling
✅ Persist data reliably
✅ Support your multi-tenant requirements

**Ready to deploy? Start with:** `./volumes-setup.sh` then follow `DEPLOYMENT_GUIDE.md`


