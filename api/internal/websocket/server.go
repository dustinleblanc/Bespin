package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Server represents a WebSocket server
type Server struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	upgrader   websocket.Upgrader
	logger     *log.Logger
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	// Track clients by job ID
	jobClients map[string][]*Client
}

// Client represents a WebSocket client
type Client struct {
	conn   *websocket.Conn
	server *Server
	send   chan []byte
	jobID  string
}

// JobStatus represents a job status update
type JobStatus struct {
	Type   string      `json:"type"`
	JobID  string      `json:"job_id"`
	Status string      `json:"status"`
	Result interface{} `json:"result,omitempty"`
}

// NewServer creates a new WebSocket server
func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins
			},
		},
		logger:     log.New(log.Writer(), "[WebSocket] ", log.LstdFlags),
		ctx:        ctx,
		cancel:     cancel,
		jobClients: make(map[string][]*Client),
	}
}

// Start starts the WebSocket server
func (s *Server) Start() {
	s.logger.Println("Starting WebSocket server")
	for {
		select {
		case <-s.ctx.Done():
			return
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			// Add client to job subscribers
			s.jobClients[client.jobID] = append(s.jobClients[client.jobID], client)
			s.mu.Unlock()
			s.logger.Printf("Client connected: %p, Job ID: %s", client, client.jobID)
		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
				// Remove client from job subscribers
				if clients, ok := s.jobClients[client.jobID]; ok {
					for i, c := range clients {
						if c == client {
							s.jobClients[client.jobID] = append(clients[:i], clients[i+1:]...)
							break
						}
					}
					if len(s.jobClients[client.jobID]) == 0 {
						delete(s.jobClients, client.jobID)
					}
				}
			}
			s.mu.Unlock()
			s.logger.Printf("Client disconnected: %p", client)
		case message := <-s.broadcast:
			s.mu.Lock()
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, client)
				}
			}
			s.mu.Unlock()
		}
	}
}

// Stop stops the WebSocket server
func (s *Server) Stop() {
	s.logger.Println("Stopping WebSocket server")
	s.cancel()
}

// HandleConnection handles a new WebSocket connection
func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request, jobID string) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		conn:   conn,
		server: s,
		send:   make(chan []byte, 256),
		jobID:  jobID,
	}

	s.register <- client

	go client.writePump()
	go client.readPump()
}

// NotifyJobStatus notifies clients about job status changes
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
	defer s.mu.Unlock()

	// Send to clients subscribed to this job
	if clients, ok := s.jobClients[jobID]; ok {
		for _, client := range clients {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(s.clients, client)
			}
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.server.logger.Printf("Error reading message: %v", err)
			}
			break
		}

		// Handle incoming messages if needed
		c.server.logger.Printf("Received message from client: %s", message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
