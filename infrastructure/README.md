# Infrastructure Components

This directory contains all infrastructure-related configurations and deployment files for the Agba Voice AI Platform.

## Overview

The infrastructure is designed as a cloud-native, microservices-based architecture running on Kubernetes with high availability, scalability, and security in mind.

## Directory Structure

```
infrastructure/
├── kubernetes/          # Kubernetes configurations
│   ├── cluster/         # Cluster setup and configuration
│   ├── helm-charts/     # Helm charts for service deployment
│   ├── ingress/         # Ingress controller and routing
│   └── autoscaling/     # Autoscaling configurations
└── README.md           # This file
```

## Components

### 1. Kubernetes Cluster
- **Purpose**: Container orchestration platform
- **Configuration**: 3 master nodes, 6 worker nodes with autoscaling
- **Location**: `kubernetes/cluster/`
- **Features**:
  - Multi-zone deployment for high availability
  - Automatic node scaling based on demand
  - Network policies for security
  - Storage classes for persistent volumes

### 2. Helm Charts
- **Purpose**: Package management and deployment automation
- **Location**: `kubernetes/helm-charts/`
- **Features**:
  - Templated Kubernetes manifests
  - Environment-specific value overrides
  - Dependency management for external services
  - Rollback capabilities

### 3. Ingress Controller
- **Purpose**: External traffic routing and SSL termination
- **Technology**: NGINX Ingress Controller
- **Location**: `kubernetes/ingress/`
- **Features**:
  - SSL/TLS termination with automatic certificate management
  - Rate limiting and DDoS protection
  - WebSocket support for real-time communication
  - Security headers and CORS configuration

### 4. Autoscaling
- **Purpose**: Automatic resource scaling based on demand
- **Location**: `kubernetes/autoscaling/`
- **Features**:
  - Cluster Autoscaler for node scaling
  - Horizontal Pod Autoscaler (HPA) for pod scaling
  - Vertical Pod Autoscaler (VPA) for resource optimization
  - Custom metrics scaling for business logic

## Deployment Requirements

### Prerequisites
- Kubernetes cluster v1.28+
- Helm v3.8+
- kubectl configured with cluster access
- AWS CLI configured (for AWS deployments)

### Resource Requirements
- **Minimum**: 6 worker nodes (m5.2xlarge)
- **CPU**: 48 vCPUs total
- **Memory**: 192 GB total
- **Storage**: 1 TB persistent storage
- **Network**: 10 Gbps bandwidth

### Security Requirements
- TLS 1.2+ for all communications
- Network policies for service isolation
- RBAC for access control
- Pod Security Standards enforcement
- Regular security scanning and updates

## Monitoring and Observability

### Metrics Collection
- Prometheus for metrics scraping
- Grafana for visualization
- AlertManager for alerting

### Logging
- Fluentd for log collection
- Elasticsearch for log storage
- Kibana for log analysis

### Tracing
- Jaeger for distributed tracing
- OpenTelemetry for instrumentation

## High Availability

### Multi-Zone Deployment
- Services distributed across 3 availability zones
- Database replication across zones
- Load balancing with health checks

### Disaster Recovery
- Cross-region backup and replication
- Automated failover procedures
- Recovery time objective (RTO): 15 minutes
- Recovery point objective (RPO): 5 minutes

## Performance Targets

### Latency
- API response time: < 200ms (P95)
- Call setup time: < 3 seconds
- WebRTC connection: < 2 seconds

### Throughput
- 10,000+ concurrent calls
- 100+ calls per second setup rate
- 1M+ API requests per minute

### Availability
- 99.95% uptime SLA
- Planned maintenance windows: < 4 hours/month
- Unplanned downtime: < 2 hours/month

## Cost Optimization

### Resource Efficiency
- Autoscaling to match demand
- Spot instances for non-critical workloads
- Reserved instances for baseline capacity
- Resource quotas and limits

### Monitoring
- Cost tracking and alerting
- Resource utilization analysis
- Right-sizing recommendations
- Waste identification and elimination

## Getting Started

1. **Deploy Cluster**:
   ```bash
   kubectl apply -f kubernetes/cluster/cluster-config.yaml
   ```

2. **Install Helm Charts**:
   ```bash
   helm install agba-platform kubernetes/helm-charts/
   ```

3. **Configure Ingress**:
   ```bash
   kubectl apply -f kubernetes/ingress/nginx-ingress.yaml
   ```

4. **Enable Autoscaling**:
   ```bash
   kubectl apply -f kubernetes/autoscaling/cluster-autoscaler.yaml
   ```

## Troubleshooting

### Common Issues
- Pod scheduling failures: Check resource quotas and node capacity
- Network connectivity: Verify security groups and network policies
- Storage issues: Check persistent volume claims and storage classes
- Performance problems: Review resource limits and autoscaling settings

### Debugging Commands
```bash
# Check cluster status
kubectl get nodes
kubectl get pods --all-namespaces

# Check resource usage
kubectl top nodes
kubectl top pods --all-namespaces

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp

# Check logs
kubectl logs -f deployment/api-gateway -n agba-services
```

## Support

For infrastructure-related issues:
- Check the troubleshooting section above
- Review Kubernetes documentation
- Contact the DevOps team
- Create an issue in the project repository

## Contributing

When making infrastructure changes:
1. Test in development environment first
2. Update documentation
3. Follow security best practices
4. Get approval from DevOps team
5. Plan maintenance windows for production changes