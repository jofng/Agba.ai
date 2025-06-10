#!/bin/bash

# Fix import paths in all Go services
cd /workspace/Agba.ai

echo "Fixing import paths in all services..."

# List of services to fix
services=(
    "api-gateway"
    "user-management-service"
    "notification-service"
    "config-service"
    "recording-service"
    "freeswitch-service"
    "monitoring-service"
)

for service in "${services[@]}"; do
    echo "Fixing $service..."
    
    # Change to service directory
    cd "services/$service"
    
    # Fix go.mod to use local module name
    if [ -f "go.mod" ]; then
        # Replace github.com/agba-ai with local module name
        sed -i "s|module github.com/agba-ai/$service|module $service|g" go.mod
    fi
    
    # Fix all Go files to use relative imports
    find . -name "*.go" -type f -exec sed -i "s|github.com/agba-ai/$service/|$service/|g" {} \;
    
    # Clean up go.mod and go.sum
    go mod tidy 2>/dev/null || true
    
    cd ../..
done

echo "Import path fixes completed!"