# Multi-Tenant PDF Summary Ingestion Service

A simple microservice that accepts PDF files, summarizes them using AI, and stores results in tenant-specific databases.

## Features

- ✅ Upload PDF files via REST API
- ✅ Dynamic tenant-specific database creation (MongoDB)
- ✅ PDF text extraction
- ✅ AI-powered summarization (OpenAI)
- ✅ S3-compatible storage (MinIO)
- ✅ PostgreSQL master database for tenant metadata

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Kubernetes (Minikube) - Optional

### Local Development

1. **Clone and setup:**
```bash
cd /home/bacancy/Desktop/droadmap
cp .env.example .env
# Edit .env and add your OpenAI API key
```

2. **Start dependencies:**
```bash
docker-compose up -d
```

3. **Run the service:**
```bash
go mod download
go run cmd/api/main.go
```

4. **Test the API:**
```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -F "tenantName=acme_corp" \
  -F "pdf=@sample.pdf"
```

## API Endpoints

### Upload PDF
```
POST /api/v1/upload
Content-Type: multipart/form-data

Form fields:
- tenantName: string (required)
- pdf: file (required)

Response:
{
  "success": true,
  "data": {
    "document_id": "...",
    "tenant_name": "acme_corp",
    "file_name": "sample.pdf",
    "summary": "AI-generated summary...",
    "storage_url": "..."
  }
}
```

### Health Check
```
GET /health
```

## Architecture

```
Client -> Go API -> PostgreSQL (Master DB)
                 -> MongoDB (Tenant DBs)
                 -> MinIO (File Storage)
                 -> OpenAI (AI Summary)
```

## Project Structure

```
droadmap/
├── cmd/api/main.go           # Entry point
├── internal/
│   ├── handlers/             # HTTP handlers
│   ├── services/             # Business logic
│   ├── repository/           # Database operations
│   └── models/               # Data structures
├── docker-compose.yaml       # Local development
└── k8s/                      # Kubernetes manifests
```

## Database Schema

**PostgreSQL (Master):**
- `tenants` - Stores tenant metadata and DB connection info

**MongoDB (Per Tenant):**
- `documents` - Stores PDF data, extracted text, and summary

## Deployment

### Docker
```bash
docker build -t pdf-ingestion:latest .
docker run -p 8080:8080 pdf-ingestion:latest
```

### Kubernetes
```bash
kubectl apply -f k8s/
```

## Error Handling & Resilience

### OpenAI API Quota Management

The service handles OpenAI API quota limits gracefully:

**What happens when quota is exceeded (HTTP 429):**
1. ✅ Automatic retry with exponential backoff (1s, 2s, 4s)
2. ✅ Up to 3 retry attempts for transient errors
3. ✅ Automatic fallback to extractive summarization when AI service unavailable
4. ✅ Document is still successfully stored with fallback summary
5. ✅ No upload failure - service remains operational

**Retryable errors:**
- Status 429 (Rate limit/Quota exceeded)
- Status 500-504 (Server errors)
- Connection timeouts
- Network temporary failures

**Non-retryable errors:**
- Authentication failures
- Invalid API key
- Malformed requests

### Fallback Summary

When AI summarization fails, the system uses an intelligent fallback:
- Extracts first 500 characters from the document
- Intelligently ends at sentence boundary
- Clearly marked for manual review if needed

### Environment Variables

For proper error recovery, ensure these are set:

```bash
# OpenAI Configuration
OPENAI_API_KEY=sk-xxxx...      # Your OpenAI API key (optional if using fallback only)

# Service Configuration
PORT=8080                       # Server port
LOG_LEVEL=info                  # Logging level

# Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
MONGO_HOST=localhost
MONGO_PORT=27017
MINIO_ENDPOINT=localhost:9000
```

### Monitoring Quota Usage

To monitor OpenAI API quota usage:
1. Check OpenAI dashboard for current usage and quota limits
2. Look for "⚠ AI summarization failed" messages in logs
3. Implement quota alerts in your monitoring system
4. Consider upgrading plan or requesting higher quota if needed

**Temporary workaround if quota exceeded:**
- Delete the `OPENAI_API_KEY` environment variable
- Restart the service
- Fallback summaries will be used automatically
- No data loss - all documents are still ingested and stored

### Google Gemini API Free Tier Setup and Usage

The service can also utilize the Google Gemini API for summarization. To set up and use the free tier:

1. **Register for Google Cloud Account:**
   - Visit https://console.cloud.google.com/
   - Create a new project or select an existing one.
   - Enable the "Gemini API" in the "AI & Machine Learning" section.
   - Enable billing for the project.

2. **Generate API Key:**
   - Go to the "Credentials" section in the Google Cloud Console.
   - Click "Create Credentials" -> "API Key".
   - Copy the generated API key.

3. **Set Environment Variable:**
   - Add `GEMINI_API_KEY` to your `.env` file:
   ```bash
   GEMINI_API_KEY=your_gemini_api_key_here
   ```

4. **Configure Service to Use Gemini:**
   - The service will automatically use Gemini if `OPENAI_API_KEY` is not set.
   - If you want to force Gemini, set `OPENAI_API_KEY` to an empty string or remove it.

5. **Monitor Usage:**
   - Check Google Cloud Console for Gemini API usage.
   - Look for "⚠ AI summarization failed" messages in logs if Gemini quota is exceeded.


