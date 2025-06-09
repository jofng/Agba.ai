package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ServiceProxy handles proxying requests to backend services
type ServiceProxy struct {
	logger   *zap.Logger
	services map[string]*ServiceConfig
}

// ServiceConfig represents configuration for a backend service
type ServiceConfig struct {
	Name     string
	BaseURL  string
	Timeout  time.Duration
	HealthPath string
	Proxy    *httputil.ReverseProxy
}

// NewServiceProxy creates a new service proxy
func NewServiceProxy(logger *zap.Logger) *ServiceProxy {
	return &ServiceProxy{
		logger:   logger,
		services: make(map[string]*ServiceConfig),
	}
}

// RegisterService registers a backend service
func (sp *ServiceProxy) RegisterService(name, baseURL string, timeout time.Duration) error {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid service URL %s: %w", baseURL, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	
	// Customize the proxy director
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Gateway", "agba-api-gateway")
	}

	// Add error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		sp.logger.Error("Proxy error", 
			zap.String("service", name),
			zap.String("url", r.URL.String()),
			zap.Error(err),
		)
		
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error": "Service temporarily unavailable"}`))
	}

	// Add response modifier
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Set("X-Gateway", "agba-api-gateway")
		resp.Header.Set("X-Service", name)
		return nil
	}

	config := &ServiceConfig{
		Name:       name,
		BaseURL:    baseURL,
		Timeout:    timeout,
		HealthPath: "/health",
		Proxy:      proxy,
	}

	sp.services[name] = config
	sp.logger.Info("Registered service", 
		zap.String("name", name),
		zap.String("baseURL", baseURL),
	)

	return nil
}

// ProxyRequest proxies a request to the specified service
func (sp *ServiceProxy) ProxyRequest(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		service, exists := sp.services[serviceName]
		if !exists {
			sp.logger.Error("Service not found", zap.String("service", serviceName))
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}

		// Log the request
		sp.logger.Info("Proxying request",
			zap.String("service", serviceName),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
		)

		// Set timeout
		c.Request = c.Request.WithContext(c.Request.Context())
		
		// Proxy the request
		service.Proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// HealthCheck checks the health of a service
func (sp *ServiceProxy) HealthCheck(serviceName string) (bool, error) {
	service, exists := sp.services[serviceName]
	if !exists {
		return false, fmt.Errorf("service %s not found", serviceName)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	healthURL := strings.TrimSuffix(service.BaseURL, "/") + service.HealthPath
	resp, err := client.Get(healthURL)
	if err != nil {
		sp.logger.Error("Health check failed",
			zap.String("service", serviceName),
			zap.String("url", healthURL),
			zap.Error(err),
		)
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// GetServiceStatus returns the status of all registered services
func (sp *ServiceProxy) GetServiceStatus() map[string]interface{} {
	status := make(map[string]interface{})
	
	for name, service := range sp.services {
		healthy, err := sp.HealthCheck(name)
		serviceStatus := map[string]interface{}{
			"name":    name,
			"baseURL": service.BaseURL,
			"healthy": healthy,
		}
		
		if err != nil {
			serviceStatus["error"] = err.Error()
		}
		
		status[name] = serviceStatus
	}
	
	return status
}

// LoadBalancer handles load balancing between multiple instances
type LoadBalancer struct {
	services []*ServiceConfig
	current  int
	logger   *zap.Logger
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(logger *zap.Logger) *LoadBalancer {
	return &LoadBalancer{
		services: make([]*ServiceConfig, 0),
		current:  0,
		logger:   logger,
	}
}

// AddService adds a service instance to the load balancer
func (lb *LoadBalancer) AddService(config *ServiceConfig) {
	lb.services = append(lb.services, config)
}

// GetNextService returns the next service using round-robin
func (lb *LoadBalancer) GetNextService() *ServiceConfig {
	if len(lb.services) == 0 {
		return nil
	}
	
	service := lb.services[lb.current]
	lb.current = (lb.current + 1) % len(lb.services)
	return service
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	maxFailures int
	failures    int
	lastFailure time.Time
	timeout     time.Duration
	state       string // "closed", "open", "half-open"
	logger      *zap.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration, logger *zap.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		failures:    0,
		timeout:     timeout,
		state:       "closed",
		logger:      logger,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	if cb.state == "open" {
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = "half-open"
			cb.logger.Info("Circuit breaker transitioning to half-open")
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	}

	err := fn()
	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		
		if cb.failures >= cb.maxFailures {
			cb.state = "open"
			cb.logger.Warn("Circuit breaker opened", zap.Int("failures", cb.failures))
		}
		return err
	}

	// Reset on success
	if cb.state == "half-open" {
		cb.state = "closed"
		cb.failures = 0
		cb.logger.Info("Circuit breaker closed")
	}

	return nil
}