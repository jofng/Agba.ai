# Database Infrastructure

This directory contains all database-related configurations for the Agba Voice AI Platform, including PostgreSQL cluster setup, connection pooling, backup systems, and monitoring.

## Overview

The database infrastructure is designed for high availability, scalability, and data protection with:
- Master-slave PostgreSQL replication
- Connection pooling with PgBouncer
- Automated backup and recovery
- Comprehensive monitoring and alerting

## Directory Structure

```
database/
├── postgresql/          # PostgreSQL cluster configuration
├── pgbouncer/          # Connection pooling setup
├── backup/             # Backup and recovery system
├── monitoring/         # Database monitoring and alerts
└── README.md          # This file
```

## Components

### 1. PostgreSQL Cluster
- **Purpose**: High-availability database cluster
- **Configuration**: 1 master + 2 read replicas
- **Location**: `postgresql/postgresql-cluster.yaml`
- **Features**:
  - Streaming replication for high availability
  - WAL archiving for point-in-time recovery
  - Optimized configuration for performance
  - Persistent storage with automatic scaling

### 2. PgBouncer Connection Pooling
- **Purpose**: Database connection management and pooling
- **Location**: `pgbouncer/pgbouncer-config.yaml`
- **Features**:
  - Transaction-level pooling for optimal performance
  - Support for read/write splitting
  - Connection limits and timeout management
  - Automatic scaling based on load

### 3. Backup and Recovery System
- **Purpose**: Data protection and disaster recovery
- **Location**: `backup/backup-system.yaml`
- **Features**:
  - Daily and weekly automated backups
  - Point-in-time recovery capabilities
  - S3 integration for off-site storage
  - Automated retention management

### 4. Database Monitoring
- **Purpose**: Performance monitoring and alerting
- **Location**: `monitoring/pg-monitoring.yaml`
- **Features**:
  - Prometheus metrics collection
  - Custom query performance monitoring
  - Replication lag monitoring
  - Automated alerting for critical issues

## Database Schema

### Core Databases
1. **agba_core**: Main application data
   - Organizations and users
   - Agent configurations
   - Call metadata
   - System configurations

2. **agba_analytics**: Analytics and reporting data
   - Real-time event data
   - Aggregated metrics
   - Performance statistics
   - Business intelligence data

3. **agba_recordings**: Call recordings and transcriptions
   - Audio file metadata
   - Transcription data
   - Recording policies
   - Compliance data

4. **agba_biometrics**: Voice biometric data
   - Voice prints and templates
   - Authentication logs
   - Security policies
   - Identity verification data

## Deployment

### Prerequisites
- Kubernetes cluster with persistent volume support
- Storage classes: `gp3` for databases, `efs` for shared storage
- AWS credentials for S3 backup (optional)

### Installation Steps

1. **Create Namespace**:
   ```bash
   kubectl create namespace agba-data
   ```

2. **Deploy PostgreSQL Cluster**:
   ```bash
   kubectl apply -f postgresql/postgresql-cluster.yaml
   ```

3. **Deploy PgBouncer**:
   ```bash
   kubectl apply -f pgbouncer/pgbouncer-config.yaml
   ```

4. **Setup Backup System**:
   ```bash
   kubectl apply -f backup/backup-system.yaml
   ```

5. **Enable Monitoring**:
   ```bash
   kubectl apply -f monitoring/pg-monitoring.yaml
   ```

### Verification
```bash
# Check PostgreSQL pods
kubectl get pods -n agba-data -l app=postgresql

# Check PgBouncer status
kubectl get pods -n agba-data -l app=pgbouncer

# Verify replication
kubectl exec -n agba-data postgresql-master-0 -- psql -U agba_user -d agba_core -c "SELECT * FROM pg_stat_replication;"

# Test connection through PgBouncer
kubectl exec -n agba-data -it deployment/pgbouncer -- psql -h localhost -U agba_user -d agba_core -c "SELECT version();"
```

## Configuration

### Connection Strings

**Write Operations (Master)**:
```
postgresql://agba_user:password@pgbouncer:5432/agba_core
```

**Read Operations (Replicas)**:
```
postgresql://agba_user:password@pgbouncer-ro:5432/agba_core_ro
```

### Environment Variables
```bash
# Primary database connection
POSTGRES_URL=postgresql://agba_user:password@pgbouncer:5432/agba_core

# Read-only connection for analytics
POSTGRES_RO_URL=postgresql://agba_user:password@pgbouncer-ro:5432/agba_core_ro

# Individual database connections
ANALYTICS_DB_URL=postgresql://agba_user:password@pgbouncer:5432/agba_analytics
RECORDINGS_DB_URL=postgresql://agba_user:password@pgbouncer:5432/agba_recordings
BIOMETRICS_DB_URL=postgresql://agba_user:password@pgbouncer:5432/agba_biometrics
```

## Performance Tuning

### PostgreSQL Optimization
- **Shared Buffers**: 256MB (25% of available memory)
- **Effective Cache Size**: 1GB (75% of available memory)
- **Work Memory**: 4MB per connection
- **Maintenance Work Memory**: 64MB for maintenance operations

### PgBouncer Configuration
- **Pool Mode**: Transaction (optimal for microservices)
- **Default Pool Size**: 25 connections per database
- **Max Client Connections**: 1000 concurrent clients
- **Connection Timeouts**: 600 seconds idle timeout

### Connection Pooling Strategy
- Use transaction pooling for application connections
- Separate read and write connection pools
- Monitor connection usage and adjust pool sizes
- Implement connection retry logic in applications

## Backup and Recovery

### Backup Schedule
- **Daily Backups**: 2:00 AM UTC (full backup + WAL archiving)
- **Weekly Backups**: Sunday 1:00 AM UTC (comprehensive backup)
- **Retention**: 30 days local, 90 days in S3

### Recovery Procedures

**Point-in-Time Recovery**:
```bash
# List available backups
kubectl exec -n agba-data backup-pod -- ls /backups/

# Restore from specific backup
kubectl create job --from=cronjob/postgresql-backup-daily restore-job
kubectl set env job/restore-job BACKUP_NAME=agba_backup_20250609_020000
```

**Database Restore**:
```bash
# Restore individual database
kubectl exec -n agba-data postgresql-master-0 -- pg_restore -d agba_core /backups/agba_backup_latest/agba_core.sql
```

## Monitoring and Alerting

### Key Metrics
- **Connection Count**: Active connections vs. max connections
- **Replication Lag**: Delay between master and replicas
- **Query Performance**: Slow queries and execution times
- **Database Size**: Storage usage and growth trends
- **Backup Status**: Backup success/failure notifications

### Critical Alerts
- PostgreSQL instance down
- Replication lag > 30 seconds
- Connection count > 80% of maximum
- Backup failure
- Disk space < 20% free

### Monitoring Dashboards
Access Grafana dashboards for:
- Database performance overview
- Replication status and lag
- Query performance analysis
- Connection pool utilization
- Backup and recovery status

## Security

### Access Control
- Role-based database access
- Encrypted connections (TLS)
- Network policies for pod-to-pod communication
- Secrets management for credentials

### Data Protection
- Encryption at rest for persistent volumes
- Encrypted backups in S3
- Regular security updates
- Audit logging for database access

### Compliance
- GDPR compliance for data retention
- HIPAA compliance for healthcare data
- SOC 2 controls for security
- Regular security assessments

## Troubleshooting

### Common Issues

1. **Connection Refused**:
   ```bash
   # Check PostgreSQL status
   kubectl get pods -n agba-data -l app=postgresql
   kubectl logs -n agba-data postgresql-master-0
   ```

2. **Replication Lag**:
   ```bash
   # Check replication status
   kubectl exec -n agba-data postgresql-master-0 -- psql -U agba_user -c "SELECT * FROM pg_stat_replication;"
   ```

3. **PgBouncer Issues**:
   ```bash
   # Check PgBouncer logs
   kubectl logs -n agba-data deployment/pgbouncer
   
   # Test connection
   kubectl exec -n agba-data -it deployment/pgbouncer -- psql -h localhost -p 5432 -U agba_user -d agba_core
   ```

4. **Backup Failures**:
   ```bash
   # Check backup job logs
   kubectl logs -n agba-data job/postgresql-backup-daily
   
   # Verify S3 connectivity
   kubectl exec -n agba-data backup-pod -- aws s3 ls s3://agba-database-backups/
   ```

### Performance Issues
- Monitor slow queries with pg_stat_statements
- Check connection pool utilization
- Analyze query execution plans
- Review database configuration parameters

### Recovery Procedures
- Follow documented backup and recovery procedures
- Test recovery processes regularly
- Maintain up-to-date documentation
- Have emergency contact procedures

## Maintenance

### Regular Tasks
- Monitor database performance metrics
- Review and optimize slow queries
- Update PostgreSQL and extensions
- Test backup and recovery procedures
- Review and adjust configuration parameters

### Scaling Procedures
- Add read replicas for increased read capacity
- Scale PgBouncer instances for more connections
- Increase storage capacity as needed
- Optimize queries and indexes for performance

## Support

For database-related issues:
- Check monitoring dashboards first
- Review logs for error messages
- Verify connectivity and configuration
- Contact database administration team
- Escalate to vendor support if needed

## Contributing

When making database changes:
1. Test in development environment first
2. Follow database migration best practices
3. Update documentation and monitoring
4. Plan maintenance windows for production
5. Have rollback procedures ready