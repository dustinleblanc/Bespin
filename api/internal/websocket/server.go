package websocket

import (
	"context"
	"log"
	"net/http"

	"github.com/dustinleblanc/go-bespin/internal/queue"
	"github.com/gorilla/websocket"
)

// Server handles WebSocket connections
type Server struct {
	jobQueue   queue.JobQueueInterface
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	logger     *log.Logger
	upgrader   websocket.Upgrader
}

// Client represents a WebSocket client
type Client struct {
	server *Server
	conn   *websocket.Conn
	send   chan []byte
	id     string
}

// NewServer creates a new WebSocket server
func NewServer(jobQueue queue.JobQueueInterface) *Server {
	return &Server{
		jobQueue:   jobQueue,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		logger:     log.New(log.Writer(), "[WebSocket] ", log.LstdFlags),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
	}
}

// Start starts the WebSocket server
func (s *Server) Start(ctx context.Context) {
	s.logger.Println("Starting WebSocket server")

	// Start the client manager
	go s.run()

	// Start listening for Redis messages
	go s.listenForJobCompletions(ctx)
}

// run runs the client manager
func (s *Server) run() {
	for {
		select {
		case client := <-s.register:
			s.clients[client] = true
			s.logger.Printf("Client connected: %s", client.id)
			s.logger.Printf("Total connected clients: %d", len(s.clients))
		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
				s.logger.Printf("Client disconnected: %s", client.id)
				s.logger.Printf("Remaining connected clients: %d", len(s.clients))
			}
		case message := <-s.broadcast:
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, client)
				}
			}
		}
	}
}

// listenForJobCompletions listens for job completion events from Redis
func (s *Server) listenForJobCompletions(ctx context.Context) {
	pubsub := s.jobQueue.GetRedisClient().PSubscribe(ctx, "job-completed:*")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		s.logger.Printf("Received message: %s", msg.Channel)

		// Extract job ID from channel name (not used directly but logged for debugging)
		// jobID := msg.Channel[len("job-completed:"):]

		// Broadcast to all clients
		for client := range s.clients {
			select {
			case client.send <- []byte(msg.Payload):
				s.logger.Printf("Sent job completion event to client: %s", client.id)
			default:
				close(client.send)
				delete(s.clients, client)
			}
		}
	}
}

// ServeWs handles WebSocket requests from clients
func (s *Server) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Printf("Error upgrading connection: %v", err)
		return
	}

	client := &Client{
		server: s,
		conn:   conn,
		send:   make(chan []byte, 256),
		id:     r.RemoteAddr,
	}

	s.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
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

			// Add queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
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
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.server.logger.Printf("Error reading message: %v", err)
			}
			break
		}
	}
}
