# Security Infrastructure

This directory contains all security-related configurations for the Agba Voice AI Platform, including TLS certificate management, OAuth2 authentication, API key management, and network security policies.

## Overview

The security infrastructure provides comprehensive protection with:
- Automated TLS certificate management with Let's Encrypt
- OAuth2 authentication server for centralized identity management
- API key management system for service authentication
- Network security policies for micro-segmentation
- Security monitoring and compliance controls

## Directory Structure

```
security/
├── tls/                    # TLS certificate management
├── oauth2/                 # OAuth2 authentication server
├── api-keys/              # API key management system
├── network-policies/      # Network security policies
└── README.md             # This file
```

## Components

### 1. TLS Certificate Management
- **Purpose**: Automated certificate provisioning and renewal
- **Location**: `tls/cert-manager.yaml`
- **Features**:
  - Cert-Manager for automatic certificate lifecycle
  - Let's Encrypt integration for public certificates
  - Internal CA for service-to-service communication
  - Automatic renewal and monitoring
  - Multi-domain and wildcard certificate support

### 2. OAuth2 Authentication Server
- **Purpose**: Centralized authentication and authorization
- **Location**: `oauth2/oauth2-server.yaml`
- **Features**:
  - OAuth2 and OpenID Connect support
  - Multi-client configuration (dashboard, API, mobile, WebRTC)
  - JWT token management with configurable TTL
  - Session management with Redis
  - Rate limiting and security controls

### 3. API Key Management System
- **Purpose**: API key generation, validation, and lifecycle management
- **Location**: `api-keys/api-key-manager.yaml`
- **Features**:
  - Centralized API key management
  - Scope-based permissions and rate limiting
  - High-performance validation service
  - Usage tracking and analytics
  - Automatic key rotation and expiration

### 4. Network Security Policies
- **Purpose**: Network micro-segmentation and traffic control
- **Location**: `network-policies/network-security.yaml`
- **Features**:
  - Kubernetes Network Policies for all namespaces
  - Default deny-all with explicit allow rules
  - Service-to-service communication controls
  - Pod Security Policies and Security Context Constraints
  - Network security monitoring with Falco

## Security Architecture

### Authentication Flow
1. **User Authentication**: OAuth2 with PKCE for web/mobile clients
2. **Service Authentication**: API keys for service-to-service communication
3. **Internal Communication**: mTLS with internal CA certificates
4. **Token Validation**: Centralized JWT validation with Redis caching

### Authorization Model
- **Role-Based Access Control (RBAC)**: User roles and permissions
- **Scope-Based API Access**: Fine-grained API permissions
- **Resource-Level Security**: Per-resource access controls
- **Rate Limiting**: Per-user and per-API-key rate limits

### Network Security
- **Zero Trust Architecture**: Default deny with explicit allow
- **Namespace Isolation**: Separate namespaces for different components
- **Service Mesh Ready**: Prepared for Istio/Linkerd integration
- **Traffic Encryption**: TLS for all external and internal communication

## Deployment

### Prerequisites
- Kubernetes cluster with Network Policy support
- DNS management for certificate validation
- SMTP server for email notifications (optional)

### Installation Steps

1. **Deploy Certificate Management**:
   ```bash
   kubectl apply -f tls/cert-manager.yaml
   ```

2. **Deploy OAuth2 Server**:
   ```bash
   kubectl apply -f oauth2/oauth2-server.yaml
   ```

3. **Deploy API Key Management**:
   ```bash
   kubectl apply -f api-keys/api-key-manager.yaml
   ```

4. **Apply Network Policies**:
   ```bash
   kubectl apply -f network-policies/network-security.yaml
   ```

### Verification
```bash
# Check certificate status
kubectl get certificates -A
kubectl describe certificate api-gateway-tls -n agba-system

# Check OAuth2 server
kubectl get pods -n agba-system -l app=oauth2-server
kubectl logs -n agba-system deployment/oauth2-server

# Check API key services
kubectl get pods -n agba-system -l app=api-key-manager
kubectl get pods -n agba-system -l app=api-key-validator

# Verify network policies
kubectl get networkpolicies -A
kubectl describe networkpolicy api-gateway-policy -n agba-system
```

## Configuration

### TLS Certificates

**Public Certificates** (Let's Encrypt):
- `api.agba.ai` - API Gateway
- `dashboard.agba.ai` - Dashboard
- `auth.agba.ai` - OAuth2 Server
- `webrtc.agba.ai` - WebRTC Gateway

**Internal Certificates** (Internal CA):
- `*.agba-system.svc.cluster.local` - System services
- `*.agba-data.svc.cluster.local` - Data services
- `*.agba-telephony.svc.cluster.local` - Telephony services

### OAuth2 Clients

**Dashboard Client**:
```yaml
client_id: agba-dashboard
grant_types: [authorization_code, refresh_token]
scopes: [openid, profile, email, dashboard:read, dashboard:write]
```

**API Client**:
```yaml
client_id: agba-api
grant_types: [client_credentials, authorization_code]
scopes: [api:read, api:write, calls:manage, agents:manage]
```

**Mobile Client**:
```yaml
client_id: agba-mobile
grant_types: [authorization_code, refresh_token]
scopes: [openid, profile, mobile:access]
public: true
```

### API Key Scopes

**Call Management**:
- `calls:read` - Read call information
- `calls:write` - Create and manage calls
- `calls:delete` - Delete calls

**Agent Management**:
- `agents:read` - Read agent configurations
- `agents:write` - Create and modify agents
- `agents:delete` - Delete agents

**Analytics Access**:
- `analytics:read` - Access analytics data
- `recordings:read` - Access call recordings
- `recordings:delete` - Delete call recordings

**Administrative**:
- `admin:read` - Administrative read access
- `admin:write` - Administrative write access
- `webhooks:manage` - Manage webhook configurations

## Security Best Practices

### Certificate Management
- Use Let's Encrypt for public-facing services
- Implement certificate rotation monitoring
- Use internal CA for service-to-service communication
- Monitor certificate expiration with alerts

### Authentication & Authorization
- Implement OAuth2 with PKCE for public clients
- Use short-lived access tokens (1 hour)
- Implement refresh token rotation
- Monitor authentication failures and rate limit

### API Security
- Validate all API keys on every request
- Implement rate limiting per API key
- Log all API access for audit trails
- Use HTTPS for all API communication

### Network Security
- Implement default deny network policies
- Use namespace isolation for different components
- Monitor network traffic for anomalies
- Implement egress filtering for external access

## Monitoring and Alerting

### Security Metrics
- **Certificate Expiration**: Monitor certificate validity
- **Authentication Failures**: Track failed login attempts
- **API Key Usage**: Monitor API key usage patterns
- **Network Policy Violations**: Detect policy violations

### Critical Alerts
- Certificate expiring within 7 days
- High authentication failure rate
- API key rate limit exceeded
- Network policy violations
- Suspicious network activity

### Security Dashboards
Access Grafana dashboards for:
- Certificate management overview
- Authentication and authorization metrics
- API key usage and rate limiting
- Network security monitoring

## Compliance

### Standards Compliance
- **SOC 2 Type 2**: Security controls and monitoring
- **GDPR**: Data protection and privacy controls
- **HIPAA**: Healthcare data protection (if applicable)
- **PCI DSS**: Payment card data security (if applicable)

### Audit Requirements
- **Access Logging**: All authentication and API access
- **Change Tracking**: Configuration and permission changes
- **Security Events**: Failed authentications and policy violations
- **Data Access**: Tracking of sensitive data access

### Data Protection
- **Encryption at Rest**: All persistent data encrypted
- **Encryption in Transit**: TLS for all communication
- **Key Management**: Secure key storage and rotation
- **Data Retention**: Configurable retention policies

## Troubleshooting

### Common Issues

1. **Certificate Issues**:
   ```bash
   # Check certificate status
   kubectl describe certificate api-gateway-tls -n agba-system
   
   # Check cert-manager logs
   kubectl logs -n cert-manager deployment/cert-manager
   ```

2. **OAuth2 Authentication Issues**:
   ```bash
   # Check OAuth2 server logs
   kubectl logs -n agba-system deployment/oauth2-server
   
   # Test OAuth2 endpoint
   curl -k https://auth.agba.ai/health
   ```

3. **API Key Validation Issues**:
   ```bash
   # Check API key validator logs
   kubectl logs -n agba-system deployment/api-key-validator
   
   # Test API key validation
   curl -H "Authorization: Bearer agba_test_key" https://api.agba.ai/health
   ```

4. **Network Policy Issues**:
   ```bash
   # Check network policy status
   kubectl get networkpolicies -A
   
   # Test connectivity
   kubectl exec -n agba-system deployment/api-gateway -- curl redis-master:6379
   ```

### Security Incident Response
1. **Immediate Response**: Isolate affected components
2. **Investigation**: Analyze logs and security events
3. **Containment**: Apply additional security controls
4. **Recovery**: Restore services with enhanced security
5. **Post-Incident**: Update security policies and procedures

## Maintenance

### Regular Tasks
- Monitor certificate expiration and renewal
- Review and rotate API keys and secrets
- Update security policies and network rules
- Conduct security assessments and penetration testing
- Review access logs and audit trails

### Security Updates
- Keep cert-manager and security components updated
- Apply security patches promptly
- Update TLS configurations and cipher suites
- Review and update network policies
- Conduct regular security training

## Support

For security-related issues:
- Check security monitoring dashboards first
- Review security logs and audit trails
- Follow incident response procedures
- Contact security team for escalation
- Document security events and resolutions

## Contributing

When making security changes:
1. Follow security best practices and standards
2. Test in development environment first
3. Update security documentation and procedures
4. Conduct security review and approval
5. Plan maintenance windows for production changes