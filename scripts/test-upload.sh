#!/bin/bash

# Simple script to test the upload API

API_URL="${API_URL:-http://localhost:8080/api/v1/upload}"
TENANT_NAME="${TENANT_NAME:-test_company}"
PDF_FILE="${PDF_FILE:-sample.pdf}"

echo "🧪 Testing PDF Upload API"
echo "=========================="
echo "API URL: $API_URL"
echo "Tenant: $TENANT_NAME"
echo "PDF File: $PDF_FILE"
echo ""

# Check if PDF file exists
if [ ! -f "$PDF_FILE" ]; then
    echo "❌ Error: PDF file '$PDF_FILE' not found!"
    echo "Please provide a PDF file or set PDF_FILE environment variable."
    exit 1
fi

echo "📤 Uploading PDF..."
response=$(curl -s -X POST "$API_URL" \
    -F "tenantName=$TENANT_NAME" \
    -F "pdf=@$PDF_FILE" \
    -w "\n%{http_code}")

# Extract HTTP status code (last line)
http_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | head -n -1)

echo ""
echo "HTTP Status: $http_code"
echo "Response:"
echo "$body" | jq '.' 2>/dev/null || echo "$body"

if [ "$http_code" = "200" ]; then
    echo ""
    echo "✅ Upload successful!"
else
    echo ""
    echo "❌ Upload failed!"
    exit 1
fi

