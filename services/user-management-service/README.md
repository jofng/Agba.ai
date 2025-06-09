# User Management Service

The User Management Service provides comprehensive user authentication, authorization, and organization management for the Agba Voice AI Platform. It handles user registration, login, role-based access control, organization management, and user profiles with advanced security features.

## Overview

The User Management Service offers centralized identity and access management with:
- **User Authentication & Authorization** with JWT tokens and OAuth2 integration
- **Organization Management** with multi-tenant support and role-based access
- **Role-Based Access Control (RBAC)** with granular permissions
- **User Profile Management** with preferences and activity tracking
- **Session Management** with device tracking and security monitoring
- **Email Verification & Password Reset** with secure token-based workflows
- **Audit Logging** for compliance and security monitoring
- **Two-Factor Authentication** for enhanced security

## Architecture

### Technology Stack
- **Language**: Go (Golang) 1.21
- **Framework**: Gin HTTP framework for REST API
- **Database**: PostgreSQL for persistent storage of users, organizations, and audit logs
- **Cache**: Redis for session storage and rate limiting
- **Messaging**: NATS for real-time notifications and event streaming
- **Authentication**: JWT tokens with refresh token rotation
- **Email**: SMTP integration for notifications and verification
- **Security**: bcrypt password hashing, rate limiting, and account lockout

### Key Components
1. **Authentication Manager**: Handles login, logout, and token management
2. **User Service**: Manages user accounts, profiles, and preferences
3. **Organization Service**: Handles organization management and multi-tenancy
4. **Role Service**: Manages roles and permissions with RBAC
5. **Session Service**: Tracks user sessions and device information
6. **Audit Service**: Logs all user actions for compliance and security
7. **Email Service**: Sends verification emails and notifications

## Features

### User Authentication & Authorization
- **JWT-based authentication** with access and refresh tokens
- **OAuth2 integration** for third-party authentication providers
- **Two-factor authentication** with TOTP and backup codes
- **Account lockout protection** against brute force attacks
- **Password policies** with complexity requirements
- **Session management** with device tracking and revocation
- **Rate limiting** for login attempts and registration

### Organization Management
- **Multi-tenant architecture** with organization isolation
- **Organization settings** and customization options
- **User invitations** with role-based access assignment
- **Subscription management** with usage limits and quotas
- **Organization-specific roles** and permissions
- **Billing and contact information** management

### Role-Based Access Control
- **Hierarchical role system** with system and organization roles
- **Granular permissions** for fine-grained access control
- **Role inheritance** and permission aggregation
- **Dynamic role assignment** with expiration support
- **Permission scoping** (system, organization, self)
- **Default roles** for common use cases

### User Profile Management
- **Comprehensive user profiles** with personal information
- **Avatar upload** and profile picture management
- **User preferences** and settings customization
- **Activity tracking** and audit trails
- **Language and timezone** preferences
- **Contact information** management

### Security Features
- **Password hashing** with bcrypt and salt
- **Account verification** via email tokens
- **Password reset** with secure token-based workflow
- **Session security** with IP and device tracking
- **Audit logging** for all user actions
- **Rate limiting** and DDoS protection
- **Input validation** and sanitization

## API Endpoints

### Authentication
```
POST   /api/v1/auth/login                       # User login
POST   /api/v1/auth/logout                      # User logout
POST   /api/v1/auth/refresh                     # Refresh access token
POST   /api/v1/auth/forgot-password             # Request password reset
POST   /api/v1/auth/reset-password              # Reset password with token
POST   /api/v1/auth/verify-email                # Verify email address
POST   /api/v1/auth/resend-verification         # Resend verification email
```

### User Management
```
POST   /api/v1/users/register                   # User registration
GET    /api/v1/users                            # List users (admin)
GET    /api/v1/users/me                         # Get current user profile
PUT    /api/v1/users/me                         # Update current user profile
PUT    /api/v1/users/me/password                # Change password
GET    /api/v1/users/{id}                       # Get user by ID
PUT    /api/v1/users/{id}                       # Update user (admin)
DELETE /api/v1/users/{id}                       # Delete user (admin)
POST   /api/v1/users/{id}/activate              # Activate user account
POST   /api/v1/users/{id}/deactivate            # Deactivate user account
```

### Role Management
```
GET    /api/v1/users/{id}/roles                 # Get user roles
POST   /api/v1/users/{id}/roles                 # Assign role to user
DELETE /api/v1/users/{id}/roles/{roleId}        # Remove role from user
GET    /api/v1/users/{id}/permissions           # Get user permissions
GET    /api/v1/roles                            # List roles
GET    /api/v1/roles/{id}                       # Get role details
POST   /api/v1/roles                            # Create role
PUT    /api/v1/roles/{id}                       # Update role
DELETE /api/v1/roles/{id}                       # Delete role
```

### Organization Management
```
POST   /api/v1/organizations/register           # Register organization
GET    /api/v1/organizations                    # List organizations
GET    /api/v1/organizations/{id}               # Get organization details
PUT    /api/v1/organizations/{id}               # Update organization
DELETE /api/v1/organizations/{id}               # Delete organization
GET    /api/v1/organizations/{id}/users         # Get organization users
POST   /api/v1/organizations/{id}/users         # Add user to organization
DELETE /api/v1/organizations/{id}/users/{userId} # Remove user from organization
POST   /api/v1/organizations/{id}/invite        # Invite user to organization
```

### Session Management
```
GET    /api/v1/sessions                         # Get user sessions
DELETE /api/v1/sessions/{id}                    # Revoke specific session
DELETE /api/v1/sessions                         # Revoke all sessions
GET    /api/v1/users/{id}/sessions              # Get user sessions (admin)
DELETE /api/v1/users/{id}/sessions/{sessionId}  # Revoke user session (admin)
```

### Profile Management
```
GET    /api/v1/profile                          # Get user profile
PUT    /api/v1/profile                          # Update user profile
POST   /api/v1/profile/avatar                   # Upload profile avatar
DELETE /api/v1/profile/avatar                   # Delete profile avatar
GET    /api/v1/profile/preferences              # Get user preferences
PUT    /api/v1/profile/preferences              # Update user preferences
GET    /api/v1/profile/activity                 # Get user activity log
```

### Audit & Administration
```
GET    /api/v1/audit                            # Get audit logs
GET    /api/v1/audit/{id}                       # Get specific audit log
GET    /api/v1/audit/users/{userId}             # Get user audit logs
GET    /api/v1/audit/organizations/{orgId}      # Get organization audit logs
GET    /api/v1/admin/stats                      # Get system statistics
GET    /api/v1/admin/health                     # Get system health
```

## Configuration

### Environment Variables

#### Server Configuration
```bash
AGBA_USER_ENVIRONMENT=production                # Environment (development/production)
AGBA_USER_VERSION=1.0.0                        # Service version
AGBA_USER_SERVER_PORT=8080                     # HTTP server port
AGBA_USER_SERVER_READ_TIMEOUT=30s              # Request read timeout
AGBA_USER_SERVER_WRITE_TIMEOUT=30s             # Response write timeout
AGBA_USER_SERVER_IDLE_TIMEOUT=120s             # Connection idle timeout
```

#### Database Configuration
```bash
AGBA_USER_DATABASE_DRIVER=postgres             # Database driver
AGBA_USER_DATABASE_DSN=postgresql://user:pass@host:5432/db # Database connection string
AGBA_USER_DATABASE_MAX_OPEN_CONNS=25           # Maximum open connections
AGBA_USER_DATABASE_MAX_IDLE_CONNS=5            # Maximum idle connections
AGBA_USER_DATABASE_CONN_MAX_LIFETIME=300s      # Connection maximum lifetime
```

#### Redis Configuration
```bash
AGBA_USER_REDIS_ADDRESS=redis-master:6379      # Redis server address
AGBA_USER_REDIS_PASSWORD=your_password         # Redis password
AGBA_USER_REDIS_DB=3                           # Redis database number
AGBA_USER_REDIS_POOL_SIZE=10                   # Connection pool size
```

#### NATS Configuration
```bash
AGBA_USER_NATS_URL=nats://nats-client:4222     # NATS server URL
AGBA_USER_NATS_USERNAME=agba_user              # NATS username
AGBA_USER_NATS_PASSWORD=your_password          # NATS password
```

#### Authentication Configuration
```bash
AGBA_USER_AUTH_JWT_SECRET=your_jwt_secret      # JWT signing secret
AGBA_USER_AUTH_JWT_ISSUER=agba.ai              # JWT issuer
AGBA_USER_AUTH_JWT_AUDIENCE=agba.ai            # JWT audience
AGBA_USER_AUTH_ACCESS_TOKEN_TTL=15m            # Access token lifetime
AGBA_USER_AUTH_REFRESH_TOKEN_TTL=168h          # Refresh token lifetime (7 days)
AGBA_USER_AUTH_SESSION_TTL=24h                 # Session lifetime
AGBA_USER_AUTH_MAX_SESSIONS=5                  # Maximum concurrent sessions
```

#### Password Policy Configuration
```bash
AGBA_USER_AUTH_PASSWORD_MIN_LENGTH=8           # Minimum password length
AGBA_USER_AUTH_PASSWORD_MAX_LENGTH=128         # Maximum password length
AGBA_USER_AUTH_PASSWORD_REQUIRE_UPPER=true     # Require uppercase letters
AGBA_USER_AUTH_PASSWORD_REQUIRE_LOWER=true     # Require lowercase letters
AGBA_USER_AUTH_PASSWORD_REQUIRE_DIGIT=true     # Require digits
AGBA_USER_AUTH_PASSWORD_REQUIRE_SPECIAL=true   # Require special characters
```

#### Account Security Configuration
```bash
AGBA_USER_AUTH_MAX_LOGIN_ATTEMPTS=5            # Maximum failed login attempts
AGBA_USER_AUTH_LOCKOUT_DURATION=30m            # Account lockout duration
AGBA_USER_AUTH_TWO_FACTOR_ENABLED=false        # Enable two-factor authentication
AGBA_USER_AUTH_EMAIL_VERIFICATION_REQUIRED=true # Require email verification
AGBA_USER_AUTH_EMAIL_VERIFICATION_TTL=24h      # Email verification token lifetime
AGBA_USER_AUTH_PASSWORD_RESET_TTL=1h           # Password reset token lifetime
```

#### Rate Limiting Configuration
```bash
AGBA_USER_AUTH_RATE_LIMIT_ENABLED=true         # Enable rate limiting
AGBA_USER_AUTH_LOGIN_RATE_LIMIT=10             # Login attempts per minute
AGBA_USER_AUTH_REGISTRATION_RATE_LIMIT=5       # Registrations per hour
```

#### User Management Configuration
```bash
AGBA_USER_USER_MANAGEMENT_DEFAULT_ROLE=user    # Default role for new users
AGBA_USER_USER_MANAGEMENT_ALLOW_SELF_REGISTRATION=true # Allow self-registration
AGBA_USER_USER_MANAGEMENT_REQUIRE_EMAIL_VERIFICATION=true # Require email verification
AGBA_USER_USER_MANAGEMENT_ALLOW_USERNAME_LOGIN=true # Allow username login
AGBA_USER_USER_MANAGEMENT_ALLOW_EMAIL_LOGIN=true # Allow email login
AGBA_USER_USER_MANAGEMENT_USERNAME_MIN_LENGTH=3 # Minimum username length
AGBA_USER_USER_MANAGEMENT_USERNAME_MAX_LENGTH=50 # Maximum username length
```

#### Email Configuration
```bash
AGBA_USER_EMAIL_ENABLED=true                   # Enable email notifications
AGBA_USER_EMAIL_PROVIDER=smtp                  # Email provider (smtp, sendgrid, ses)
AGBA_USER_EMAIL_SMTP_HOST=smtp.gmail.com       # SMTP server host
AGBA_USER_EMAIL_SMTP_PORT=587                  # SMTP server port
AGBA_USER_EMAIL_SMTP_USER=noreply@agba.ai      # SMTP username
AGBA_USER_EMAIL_SMTP_PASS=your_password        # SMTP password
AGBA_USER_EMAIL_FROM_EMAIL=noreply@agba.ai     # From email address
AGBA_USER_EMAIL_FROM_NAME=Agba.ai              # From name
```

#### Storage Configuration
```bash
AGBA_USER_STORAGE_PROVIDER=local               # Storage provider (local, s3, gcs)
AGBA_USER_STORAGE_LOCAL_PATH=./uploads         # Local storage path
AGBA_USER_STORAGE_MAX_FILE_SIZE=10485760       # Maximum file size (10MB)
AGBA_USER_STORAGE_ALLOWED_TYPES=image/jpeg,image/png,image/gif # Allowed file types
```

## Development

### Local Development Setup
```bash
# Clone repository
git clone https://github.com/ibilola104/Agba.ai.git
cd Agba.ai/services/user-management-service

# Install dependencies
go mod download

# Set required environment variables
export AGBA_USER_DATABASE_DSN=postgresql://localhost:5432/agba_core
export AGBA_USER_REDIS_ADDRESS=localhost:6379
export AGBA_USER_REDIS_PASSWORD=your_redis_password
export AGBA_USER_NATS_URL=nats://localhost:4222
export AGBA_USER_NATS_PASSWORD=your_nats_password
export AGBA_USER_AUTH_JWT_SECRET=your_jwt_secret

# Start the service
go run cmd/main.go
```

### Testing
```bash
# Run unit tests
go test ./...

# Run integration tests
go test -tags=integration ./...

# Run tests with coverage
go test -cover ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Test user registration
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d @examples/register-user.json

# Test user login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d @examples/login-user.json

# Test protected endpoint
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <access_token>"
```

### Building
```bash
# Build binary
go build -o user-management-service cmd/main.go

# Build Docker image
docker build -t agba/user-management-service:latest .

# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -o user-management-service-linux cmd/main.go
```

## Deployment

### Kubernetes Deployment
```bash
# Deploy to Kubernetes
kubectl apply -f deployments/deployment.yaml

# Check deployment status
kubectl get pods -l app=user-management-service -n agba-system
kubectl logs -l app=user-management-service -n agba-system

# Check service endpoints
kubectl get endpoints user-management-service -n agba-system

# Scale deployment
kubectl scale deployment user-management-service --replicas=5 -n agba-system
```

### Database Setup
```bash
# Run database schema setup
kubectl apply -f deployments/deployment.yaml

# Verify database tables
kubectl exec -it postgresql-0 -n agba-data -- \
  psql -U agba_user -d agba_core -c "\dt user*"

# Check database health
kubectl exec -it deployment/user-management-service -n agba-system -- \
  curl http://localhost:8080/ready
```

### Initial Setup
```bash
# Create first admin user (run once)
kubectl exec -it deployment/user-management-service -n agba-system -- \
  curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@agba.ai",
    "password": "SecurePassword123!",
    "confirm_password": "SecurePassword123!",
    "first_name": "System",
    "last_name": "Administrator",
    "accept_terms": true
  }'

# Assign admin role (replace user_id with actual ID)
kubectl exec -it postgresql-0 -n agba-data -- \
  psql -U agba_user -d agba_core -c "
    INSERT INTO user_roles (user_id, role_id) 
    VALUES ('<user_id>', '00000000-0000-0000-0000-000000000002');
  "
```

## Security

### Authentication & Authorization
- **JWT token-based authentication** with secure signing and validation
- **Role-based access control** with granular permissions
- **Session management** with device tracking and security monitoring
- **Two-factor authentication** for enhanced account security
- **OAuth2 integration** for third-party authentication providers
- **API key authentication** for service-to-service communication

### Data Protection & Privacy
- **Password hashing** with bcrypt and secure salt generation
- **Secure token generation** for email verification and password reset
- **Data encryption** in transit with TLS and at rest for sensitive data
- **Input validation** and sanitization to prevent injection attacks
- **Rate limiting** to prevent brute force and DDoS attacks
- **Account lockout** protection against credential stuffing

### Compliance & Governance
- **Comprehensive audit logging** for all user actions and changes
- **Data retention policies** with configurable retention periods
- **Privacy controls** for user data and profile information
- **GDPR compliance** with data export and deletion capabilities
- **Session security** with IP tracking and device fingerprinting

## Monitoring & Observability

### Health Checks
- **`/health`**: Basic service health check
- **`/ready`**: Readiness check including database and Redis connectivity
- **`/metrics`**: Prometheus metrics endpoint for monitoring

### Key Metrics
- **Authentication metrics**: Login success/failure rates, token generation
- **User activity metrics**: Registration rates, active users, session duration
- **Security metrics**: Failed login attempts, account lockouts, suspicious activity
- **Performance metrics**: Response times, database query performance
- **System metrics**: Memory usage, CPU utilization, connection pool status

### Alerting Rules
- **Authentication failures**: High rate of failed login attempts
- **Account security**: Suspicious login patterns or account lockouts
- **Service availability**: Service downtime or high error rates
- **Performance degradation**: High response times or database issues
- **Security incidents**: Potential brute force attacks or data breaches

## Troubleshooting

### Common Issues

1. **Authentication Failures**
   ```bash
   # Check JWT secret configuration
   kubectl get secret user-management-secrets -o yaml
   
   # Check user account status
   curl http://localhost:8080/api/v1/users/me \
     -H "Authorization: Bearer <token>"
   
   # Check authentication logs
   kubectl logs -l app=user-management-service -n agba-system | grep auth
   ```

2. **Database Connection Issues**
   ```bash
   # Check database connectivity
   kubectl exec -it deployment/user-management-service -n agba-system -- \
     pg_isready -h pgbouncer -p 5432 -U agba_user
   
   # Check connection pool status
   curl http://localhost:8080/ready
   
   # Monitor database metrics
   curl http://localhost:8080/metrics | grep database
   ```

3. **Email Delivery Problems**
   ```bash
   # Check email configuration
   kubectl get configmap user-management-service-config -o yaml
   
   # Check SMTP credentials
   kubectl get secret user-management-secrets -o yaml
   
   # Test email connectivity
   kubectl logs -l app=user-management-service -n agba-system | grep email
   ```

4. **Session Management Issues**
   ```bash
   # Check Redis connectivity
   kubectl exec -it deployment/user-management-service -n agba-system -- \
     redis-cli -h redis-master -a <password> ping
   
   # Check active sessions
   curl http://localhost:8080/api/v1/sessions \
     -H "Authorization: Bearer <token>"
   
   # Monitor session metrics
   curl http://localhost:8080/metrics | grep session
   ```

### Debug Commands
```bash
# Service health check
curl http://localhost:8080/health

# Detailed readiness check
curl http://localhost:8080/ready

# Get service metrics
curl http://localhost:8080/metrics

# Test user registration
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d @test-user.json

# Test user login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"password"}'

# Get current user info
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <token>"

# List user roles
curl http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer <token>"

# Get system statistics
curl http://localhost:8080/api/v1/admin/stats \
  -H "Authorization: Bearer <admin_token>"
```

### Performance Debugging
```bash
# Check authentication performance
kubectl logs -l app=user-management-service -n agba-system | grep "auth"

# Monitor resource usage
kubectl top pods -l app=user-management-service -n agba-system

# Check HPA status
kubectl get hpa user-management-service-hpa -n agba-system

# Database performance
kubectl exec -it postgresql-0 -n agba-data -- \
  psql -U agba_user -d agba_core -c "SELECT * FROM pg_stat_activity;"

# Redis performance
kubectl exec -it deployment/user-management-service -n agba-system -- \
  redis-cli -h redis-master -a <password> info stats
```

## Contributing

### Development Guidelines
1. **Follow Go coding standards**: Use gofmt, golint, and go vet
2. **Add comprehensive unit tests**: Aim for >80% code coverage
3. **Update API documentation**: Document all endpoint changes
4. **Ensure proper error handling**: Use structured error responses
5. **Test authentication flows**: Thoroughly test login, registration, and token management
6. **Security testing**: Include security tests for authentication and authorization

### Code Review Process
1. **Create feature branch**: Branch from main for new features
2. **Implement with tests**: Add comprehensive tests for new functionality
3. **Update documentation**: Update README and API documentation
4. **Submit pull request**: Include detailed description and test results
5. **Address feedback**: Respond to code review comments
6. **Merge after approval**: Merge only after CI/CD validation

### Release Process
1. **Version tagging**: Follow semantic versioning (v1.0.0)
2. **Automated testing**: Full test suite execution including integration tests
3. **Security scanning**: Vulnerability scanning and dependency checks
4. **Staged deployment**: Deploy to dev → staging → production
5. **Authentication validation**: Verify authentication functionality post-deployment