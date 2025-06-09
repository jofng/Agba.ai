package routes

import (
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.uber.org/zap"
)

// Router handles routing configuration
type Router struct {
    engine *gin.Engine
    logger *zap.Logger
}

// NewRouter creates a new router instance
func NewRouter(logger *zap.Logger) *Router {
    engine := gin.New()
    return &Router{
        engine: engine,
        logger: logger,
    }
}

// SetupRoutes configures all routes
func (r *Router) SetupRoutes() *gin.Engine {
    // Add global middleware
    r.engine.Use(gin.Logger())
    r.engine.Use(gin.Recovery())
    
    // Health endpoints
    r.engine.GET("/health", r.healthCheck)
    r.engine.GET("/ready", r.readinessCheck)
    r.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    // API routes
    api := r.engine.Group("/api/v1")
    {
        api.GET("/status", r.getStatus)
    }
    
    return r.engine
}

func (r *Router) healthCheck(c *gin.Context) {
    c.JSON(200, gin.H{"status": "healthy"})
}

func (r *Router) readinessCheck(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ready"})
}

func (r *Router) getStatus(c *gin.Context) {
    c.JSON(200, gin.H{"status": "running"})
}
