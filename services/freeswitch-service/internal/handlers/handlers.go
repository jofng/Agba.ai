package handlers

import (
	"fmt"
	"net/http"

	"github.com/agba-ai/freeswitch-service/internal/models"
	"github.com/agba-ai/freeswitch-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// FreeSWITCHHandler handles HTTP requests for FreeSWITCH operations
type FreeSWITCHHandler struct {
	service *service.FreeSWITCHService
	logger  *zap.Logger
}

// NewFreeSWITCHHandler creates a new FreeSWITCH handler
func NewFreeSWITCHHandler(service *service.FreeSWITCHService, logger *zap.Logger) *FreeSWITCHHandler {
	return &FreeSWITCHHandler{
		service: service,
		logger:  logger,
	}
}

// OriginateCall handles call origination requests
// @Summary Originate a new call
// @Description Initiates an outbound call through FreeSWITCH
// @Tags calls
// @Accept json
// @Produce json
// @Param request body models.CallRequest true "Call request"
// @Success 200 {object} models.CallResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /calls/originate [post]
func (h *FreeSWITCHHandler) OriginateCall(c *gin.Context) {
	var request models.CallRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	response, err := h.service.OriginateCall(c.Request.Context(), &request)
	if err != nil {
		h.logger.Error("Failed to originate call", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HangupCall handles call hangup requests
// @Summary Hangup a call
// @Description Hangs up an active call
// @Tags calls
// @Param call_id path string true "Call ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /calls/{call_id}/hangup [post]
func (h *FreeSWITCHHandler) HangupCall(c *gin.Context) {
	callIDStr := c.Param("call_id")
	callID, err := uuid.Parse(callIDStr)
	if err != nil {
		h.logger.Error("Invalid call ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid call ID"})
		return
	}

	if err := h.service.HangupCall(c.Request.Context(), callID); err != nil {
		h.logger.Error("Failed to hangup call", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Call hung up successfully"})
}

// TransferCall handles call transfer requests
// @Summary Transfer a call
// @Description Transfers a call to a new destination
// @Tags calls
// @Accept json
// @Produce json
// @Param request body models.CallTransferRequest true "Transfer request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /calls/transfer [post]
func (h *FreeSWITCHHandler) TransferCall(c *gin.Context) {
	var request models.CallTransferRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.service.TransferCall(c.Request.Context(), &request); err != nil {
		h.logger.Error("Failed to transfer call", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Call transferred successfully"})
}

// HoldCall handles call hold/unhold requests
// @Summary Hold or unhold a call
// @Description Puts a call on hold or removes from hold
// @Tags calls
// @Accept json
// @Produce json
// @Param request body models.CallHoldRequest true "Hold request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /calls/hold [post]
func (h *FreeSWITCHHandler) HoldCall(c *gin.Context) {
	var request models.CallHoldRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.service.HoldCall(c.Request.Context(), &request); err != nil {
		h.logger.Error("Failed to hold/unhold call", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	action := "held"
	if !request.Hold {
		action = "unheld"
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Call %s successfully", action)})
}

// ControlRecording handles call recording control requests
// @Summary Control call recording
// @Description Start, stop, pause, or resume call recording
// @Tags calls
// @Accept json
// @Produce json
// @Param request body models.CallRecordingRequest true "Recording control request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /calls/recording [post]
func (h *FreeSWITCHHandler) ControlRecording(c *gin.Context) {
	var request models.CallRecordingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.service.ControlRecording(c.Request.Context(), &request); err != nil {
		h.logger.Error("Failed to control recording", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Recording %s successfully", request.Action)})
}

// GetCall retrieves call information
// @Summary Get call information
// @Description Retrieves detailed information about a specific call
// @Tags calls
// @Param call_id path string true "Call ID"
// @Success 200 {object} models.Call
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /calls/{call_id} [get]
func (h *FreeSWITCHHandler) GetCall(c *gin.Context) {
	callIDStr := c.Param("call_id")
	callID, err := uuid.Parse(callIDStr)
	if err != nil {
		h.logger.Error("Invalid call ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid call ID"})
		return
	}

	call, err := h.service.GetCall(c.Request.Context(), callID)
	if err != nil {
		h.logger.Error("Failed to get call", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Call not found"})
		return
	}

	c.JSON(http.StatusOK, call)
}

// GetActiveCalls retrieves all active calls
// @Summary Get active calls
// @Description Retrieves a list of all currently active calls
// @Tags calls
// @Success 200 {array} models.Call
// @Failure 500 {object} map[string]string
// @Router /calls/active [get]
func (h *FreeSWITCHHandler) GetActiveCalls(c *gin.Context) {
	calls, err := h.service.GetActiveCalls(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get active calls", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"calls": calls,
		"count": len(calls),
	})
}

// GetCallStats retrieves call statistics
// @Summary Get call statistics
// @Description Retrieves comprehensive call statistics and metrics
// @Tags calls
// @Success 200 {object} models.CallStats
// @Failure 500 {object} map[string]string
// @Router /calls/stats [get]
func (h *FreeSWITCHHandler) GetCallStats(c *gin.Context) {
	stats, err := h.service.GetCallStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get call stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// HealthCheck performs health checks
// @Summary Health check
// @Description Performs comprehensive health checks for the FreeSWITCH service
// @Tags health
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /health [get]
func (h *FreeSWITCHHandler) HealthCheck(c *gin.Context) {
	healthy, checks := h.service.HealthCheck(c.Request.Context())

	status := http.StatusOK
	if !healthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"healthy": healthy,
		"checks":  checks,
		"service": "freeswitch-service",
		"version": "1.0.0",
	})
}

// GetMetrics retrieves service metrics
// @Summary Get service metrics
// @Description Retrieves various metrics about the FreeSWITCH service
// @Tags metrics
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /metrics [get]
func (h *FreeSWITCHHandler) GetMetrics(c *gin.Context) {
	stats, err := h.service.GetCallStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics := gin.H{
		"calls": gin.H{
			"total":        stats.TotalCalls,
			"active":       stats.ActiveCalls,
			"successful":   stats.SuccessfulCalls,
			"failed":       stats.FailedCalls,
			"success_rate": stats.CallSuccessRate,
		},
		"performance": gin.H{
			"average_call_time": stats.AverageCallTime,
			"peak_concurrency":  stats.PeakConcurrency,
		},
		"timestamp": stats.LastUpdated,
	}

	c.JSON(http.StatusOK, metrics)
}

// RegisterRoutes registers all HTTP routes for the FreeSWITCH handler
func (h *FreeSWITCHHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		// Call management routes
		calls := v1.Group("/calls")
		{
			calls.POST("/originate", h.OriginateCall)
			calls.POST("/:call_id/hangup", h.HangupCall)
			calls.POST("/transfer", h.TransferCall)
			calls.POST("/hold", h.HoldCall)
			calls.POST("/recording", h.ControlRecording)
			calls.GET("/:call_id", h.GetCall)
			calls.GET("/active", h.GetActiveCalls)
			calls.GET("/stats", h.GetCallStats)
		}

		// Health and metrics routes
		v1.GET("/health", h.HealthCheck)
		v1.GET("/metrics", h.GetMetrics)
	}

	// Root health check
	router.GET("/health", h.HealthCheck)
}