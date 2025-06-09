# Messaging Infrastructure

This directory contains all messaging and caching infrastructure configurations for the Agba Voice AI Platform, including NATS cluster with JetStream, Redis cluster with Sentinel, and comprehensive monitoring.

## Overview

The messaging infrastructure provides high-performance, reliable communication and caching capabilities with:
- NATS cluster with JetStream for event streaming and service communication
- Redis cluster with Sentinel for high-availability caching and session management
- Cross-region replication for disaster recovery
- Comprehensive monitoring and alerting

## Directory Structure

```
messaging/
├── nats/              # NATS cluster with JetStream
├── redis/             # Redis cluster with Sentinel
├── monitoring/        # Monitoring and alerting
└── README.md         # This file
```

## Components

### 1. NATS Cluster with JetStream
- **Purpose**: High-performance message broker and event streaming
- **Configuration**: 3-node cluster with JetStream persistence
- **Location**: `nats/nats-cluster.yaml`
- **Features**:
  - Event streaming with persistence
  - Service-to-service communication
  - Multi-tenancy with accounts
  - TLS encryption and authentication
  - Cross-region replication support

### 2. Redis Cluster with Sentinel
- **Purpose**: High-availability caching and session management
- **Configuration**: 1 master + 2 replicas + 3 sentinels
- **Location**: `redis/redis-cluster.yaml`
- **Features**:
  - Automatic failover with Sentinel
  - Master-slave replication
  - Memory optimization and eviction policies
  - Connection pooling and monitoring
  - Security with authentication and encryption

### 3. Monitoring and Alerting
- **Purpose**: Performance monitoring and operational visibility
- **Location**: `monitoring/messaging-monitoring.yaml`
- **Features**:
  - Prometheus metrics collection
  - Grafana dashboards
  - Automated alerting for critical issues
  - Performance and capacity monitoring

## NATS Configuration

### Cluster Architecture
- **3-node cluster** for high availability
- **JetStream enabled** for persistent messaging
- **Multi-account setup** for security isolation
- **TLS encryption** for secure communication

### JetStream Streams
1. **CALLS**: Call events and lifecycle management
2. **ANALYTICS**: Analytics and metrics events
3. **WEBHOOKS**: Outbound webhook events
4. **BIOMETRICS**: Voice biometric events

### Account Structure
- **SYS**: System account for administration
- **AGBA**: Main application account
- **SERVICES**: Service-to-service communication

### Connection Strings
```bash
# Application connection
nats://agba_user:password@nats-client:4222

# Service connection
nats://service_user:password@nats-client:4222

# TLS connection
nats://agba_user:password@nats-client:4222?tls=true
```

## Redis Configuration

### Cluster Architecture
- **1 master** for write operations
- **2 replicas** for read scaling
- **3 sentinels** for automatic failover
- **Connection pooling** via application libraries

### Memory Management
- **2GB memory limit** per instance
- **LRU eviction policy** for cache optimization
- **AOF persistence** for data durability
- **Compression** for storage efficiency

### Connection Strings
```bash
# Master connection (writes)
redis://redis-master:6379

# Replica connection (reads)
redis://redis-replica:6379

# Sentinel connection (failover)
redis-sentinel://redis-sentinel:26379
```

### Cache Patterns
- **Session data**: User sessions and authentication
- **Configuration cache**: Agent and system configurations
- **Rate limiting**: API rate limiting counters
- **Real-time data**: Call state and temporary data

## Deployment

### Prerequisites
- Kubernetes cluster with persistent volume support
- Storage classes: `gp3` for persistent storage
- TLS certificates for NATS (optional but recommended)

### Installation Steps

1. **Deploy NATS Cluster**:
   ```bash
   kubectl apply -f nats/nats-cluster.yaml
   ```

2. **Deploy Redis Cluster**:
   ```bash
   kubectl apply -f redis/redis-cluster.yaml
   ```

3. **Enable Monitoring**:
   ```bash
   kubectl apply -f monitoring/messaging-monitoring.yaml
   ```

### Verification
```bash
# Check NATS cluster status
kubectl get pods -n agba-system -l app=nats
kubectl exec -n agba-system nats-0 -- nats server check

# Check Redis cluster status
kubectl get pods -n agba-system -l app=redis
kubectl exec -n agba-system redis-master-0 -- redis-cli -a password info replication

# Check Sentinel status
kubectl exec -n agba-system redis-sentinel-0 -- redis-cli -p 26379 -a password sentinel masters

# Test NATS connectivity
kubectl run nats-test --rm -i --tty --image=natsio/nats-box -- nats --server=nats-client:4222 --user=agba_user --password=password server check

# Test Redis connectivity
kubectl run redis-test --rm -i --tty --image=redis:7.2-alpine -- redis-cli -h redis-master -a password ping
```

## Performance Tuning

### NATS Optimization
- **Max connections**: 64K concurrent connections
- **Max payload**: 8MB message size
- **JetStream storage**: 10GB file store, 1GB memory store
- **Clustering**: Optimized for low-latency communication

### Redis Optimization
- **Memory policy**: allkeys-lru for optimal cache performance
- **Persistence**: AOF with everysec fsync for durability
- **Replication**: Async replication for performance
- **Connection limits**: Optimized for microservices workload

### Monitoring Metrics
- **NATS**: Connection count, message rate, JetStream storage
- **Redis**: Memory usage, connection count, command rate, replication lag
- **System**: CPU, memory, disk I/O, network throughput

## Event Streaming Patterns

### NATS JetStream Usage
```go
// Publish event
js.Publish("calls.started", callEvent)

// Subscribe to events
sub, _ := js.Subscribe("calls.*", func(msg *nats.Msg) {
    // Process call event
})

// Durable consumer
js.Subscribe("analytics.>", handler, nats.Durable("analytics-processor"))
```

### Redis Caching Patterns
```go
// Set with expiration
rdb.Set(ctx, "session:"+userID, sessionData, 24*time.Hour)

// Get with fallback
val, err := rdb.Get(ctx, "config:"+agentID).Result()
if err == redis.Nil {
    // Load from database
}

// Rate limiting
pipe := rdb.Pipeline()
pipe.Incr(ctx, "rate:"+userID)
pipe.Expire(ctx, "rate:"+userID, time.Minute)
```

## Security

### NATS Security
- **Account-based isolation** for multi-tenancy
- **TLS encryption** for data in transit
- **Authentication** with username/password
- **Authorization** with subject-based permissions

### Redis Security
- **Password authentication** for all connections
- **Command renaming** to disable dangerous commands
- **Network isolation** with Kubernetes network policies
- **Encryption at rest** for persistent volumes

### Access Control
- **Role-based access** to different accounts/databases
- **Network policies** for pod-to-pod communication
- **Secrets management** for credentials
- **Audit logging** for security events

## Monitoring and Alerting

### Key Metrics
- **NATS**: Server status, connection count, message throughput, JetStream storage
- **Redis**: Instance status, memory usage, connection count, replication lag
- **Performance**: Latency, throughput, error rates

### Critical Alerts
- NATS/Redis instance down
- High memory usage (>80%)
- Connection count approaching limits
- Replication lag issues
- JetStream storage full

### Dashboards
Access Grafana dashboards for:
- NATS cluster overview and JetStream metrics
- Redis cluster status and performance
- Message throughput and latency
- Cache hit rates and performance

## Disaster Recovery

### Cross-Region Setup
- **NATS**: JetStream replication across regions
- **Redis**: Cross-region replica setup
- **Backup**: Regular snapshots and WAL archiving
- **Failover**: Automated failover procedures

### Recovery Procedures
```bash
# NATS stream backup
nats stream backup CALLS /backup/calls-stream

# Redis backup
redis-cli -h redis-master -a password --rdb /backup/redis-dump.rdb

# Restore procedures
nats stream restore CALLS /backup/calls-stream
redis-cli -h redis-master -a password --pipe < /backup/redis-dump.rdb
```

## Troubleshooting

### Common Issues

1. **NATS Connection Issues**:
   ```bash
   # Check NATS server logs
   kubectl logs -n agba-system nats-0
   
   # Test connectivity
   kubectl exec -n agba-system nats-0 -- nats server check
   ```

2. **Redis Connection Issues**:
   ```bash
   # Check Redis logs
   kubectl logs -n agba-system redis-master-0
   
   # Test connectivity
   kubectl exec -n agba-system redis-master-0 -- redis-cli -a password ping
   ```

3. **JetStream Issues**:
   ```bash
   # Check JetStream status
   kubectl exec -n agba-system nats-0 -- nats stream ls
   
   # Check consumer status
   kubectl exec -n agba-system nats-0 -- nats consumer ls CALLS
   ```

4. **Sentinel Issues**:
   ```bash
   # Check Sentinel status
   kubectl exec -n agba-system redis-sentinel-0 -- redis-cli -p 26379 -a password sentinel masters
   
   # Check failover logs
   kubectl logs -n agba-system redis-sentinel-0
   ```

### Performance Issues
- Monitor message throughput and latency
- Check memory usage and eviction rates
- Analyze slow queries and operations
- Review connection pool utilization

### Capacity Planning
- Monitor storage usage trends
- Plan for message retention policies
- Scale replicas based on read load
- Optimize memory allocation

## Maintenance

### Regular Tasks
- Monitor cluster health and performance
- Review and optimize message retention
- Update NATS and Redis versions
- Test failover procedures
- Review security configurations

### Scaling Procedures
- Add NATS nodes for increased capacity
- Scale Redis replicas for read performance
- Adjust JetStream storage limits
- Optimize connection pooling

## Support

For messaging infrastructure issues:
- Check monitoring dashboards first
- Review logs for error messages
- Test connectivity and authentication
- Contact infrastructure team for escalation
- Follow documented recovery procedures

## Contributing

When making messaging changes:
1. Test in development environment first
2. Follow message schema standards
3. Update monitoring and alerting
4. Document configuration changes
5. Plan maintenance windows for production