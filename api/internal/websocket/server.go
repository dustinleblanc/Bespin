// Package websocket provides a WebSocket server implementation for real-time job status updates.
// It supports:
// - Job-specific status notifications
// - Multiple clients per job
// - Status history for new connections
// - Future support for team and site-wide broadcasts
package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/olahol/melody"
)

// Server represents a WebSocket server that manages client connections and job status updates.
// It uses melody for WebSocket handling and maintains job-specific subscriptions and status history.
type Server struct {
	melody *melody.Melody
	logger *log.Logger
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	// Track latest status for each job
	jobStatuses map[string]JobStatus
}

// JobStatus represents a job status update message.
// It includes the job ID, current status, and optional result data.
type JobStatus struct {
	Type   string      `json:"type"`             // Message type, always "job_status"
	JobID  string      `json:"job_id"`           // ID of the job this status is for
	Status string      `json:"status"`           // Current status (pending, running, completed, failed)
	Result interface{} `json:"result,omitempty"` // Optional result data
}

// NewServer creates a new WebSocket server with default configuration.
// The server allows all origins and uses standard logging.
func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())

	// Create melody instance with default settings
	m := melody.New()

	// Configure melody
	m.Upgrader.CheckOrigin = func(r *http.Request) bool {
		return true // Allow all origins
	}

	// Create server instance
	s := &Server{
		melody:      m,
		logger:      log.New(log.Writer(), "[WebSocket] ", log.LstdFlags),
		ctx:         ctx,
		cancel:      cancel,
		jobStatuses: make(map[string]JobStatus),
	}

	// Set up melody handlers
	m.HandleConnect(s.handleConnect)
	m.HandleDisconnect(s.handleDisconnect)
	m.HandleMessage(s.handleMessage)

	return s
}

// Start starts the WebSocket server and begins processing client connections and messages.
// This method runs in a goroutine and continues until the server is stopped.
func (s *Server) Start() {
	s.logger.Println("Starting WebSocket server")
	// No need to run a separate goroutine as melody handles this internally
}

// Stop stops the WebSocket server and cancels all ongoing operations.
func (s *Server) Stop() {
	s.logger.Println("Stopping WebSocket server")
	s.cancel()
}

// HandleConnection handles a new WebSocket connection request.
// It upgrades the HTTP connection to a WebSocket connection and registers the client.
func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request, jobID string) {
	// Store job ID in the request context for melody to access
	ctx := context.WithValue(r.Context(), "job_id", jobID)
	r = r.WithContext(ctx)

	// Let melody handle the WebSocket upgrade
	s.melody.HandleRequest(w, r)
}

// NotifyJobStatus notifies all clients subscribed to a specific job about a status change.
// The status update is also stored for new clients that connect later.
func (s *Server) NotifyJobStatus(jobID string, status string, result interface{}) {
	s.logger.Printf("Notifying job status: %s, Status: %s", jobID, status)

	message := JobStatus{
		Type:   "job_status",
		JobID:  jobID,
		Status: status,
		Result: result,
	}

	data, err := json.Marshal(message)
	if err != nil {
		s.logger.Printf("Failed to marshal job status message: %v", err)
		return
	}

	s.mu.Lock()
	// Store the latest status
	s.jobStatuses[jobID] = message
	s.mu.Unlock()

	// Broadcast only to clients subscribed to this job
	s.melody.BroadcastFilter(data, func(session *melody.Session) bool {
		sessionJobID, ok := session.Request.Context().Value("job_id").(string)
		return ok && sessionJobID == jobID
	})
}

// handleConnect is called when a new WebSocket connection is established.
func (s *Server) handleConnect(session *melody.Session) {
	// Get job ID from context
	jobID, ok := session.Request.Context().Value("job_id").(string)
	if !ok {
		s.logger.Printf("No job ID found in session context")
		return
	}

	s.logger.Printf("Client connected: %p, Job ID: %s", session, jobID)

	// Send latest status if available
	s.mu.RLock()
	if status, ok := s.jobStatuses[jobID]; ok {
		data, err := json.Marshal(status)
		if err == nil {
			session.Write(data)
		}
	}
	s.mu.RUnlock()
}

// handleDisconnect is called when a WebSocket connection is closed.
func (s *Server) handleDisconnect(session *melody.Session) {
	jobID, ok := session.Request.Context().Value("job_id").(string)
	if ok {
		s.logger.Printf("Client disconnected: %p, Job ID: %s", session, jobID)
	} else {
		s.logger.Printf("Client disconnected: %p", session)
	}
}

// handleMessage is called when a message is received from a client.
// Currently, it only logs received messages. Future implementations may handle
// client-to-server messages for features like job cancellation or progress updates.
func (s *Server) handleMessage(session *melody.Session, msg []byte) {
	jobID, ok := session.Request.Context().Value("job_id").(string)
	if ok {
		s.logger.Printf("Received message from client %p (Job ID: %s): %s", session, jobID, msg)
	} else {
		s.logger.Printf("Received message from client: %s", msg)
	}
}
