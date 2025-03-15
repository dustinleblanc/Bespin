package websocket

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestWebSocketServer(t *testing.T) {
	// Create a new WebSocket server
	server := NewServer()
	go server.Start()
	defer server.Stop()

	// Create a test HTTP server
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", func(c *gin.Context) {
		server.HandleConnection(c.Writer, c.Request, "test-job-id")
	})

	// Create a test server
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	// Connect to WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer ws.Close()

	// Create a channel to receive WebSocket messages
	messages := make(chan JobStatus)
	go func() {
		for {
			var message JobStatus
			err := ws.ReadJSON(&message)
			if err != nil {
				close(messages)
				return
			}
			messages <- message
		}
	}()

	// Test job status updates
	statusUpdates := []struct {
		status string
		result interface{}
	}{
		{status: "pending", result: nil},
		{status: "running", result: nil},
		{status: "completed", result: "test result"},
		{status: "failed", result: "error message"},
	}

	for _, update := range statusUpdates {
		// Send status update
		server.NotifyJobStatus("test-job-id", update.status, update.result)

		// Wait for message
		select {
		case msg := <-messages:
			assert.Equal(t, "job_status", msg.Type)
			assert.Equal(t, "test-job-id", msg.JobID)
			assert.Equal(t, update.status, msg.Status)
			if update.result != nil {
				assert.Equal(t, update.result, msg.Result)
			}
		case <-time.After(time.Second):
			t.Fatalf("Timeout waiting for status update: %s", update.status)
		}
	}
}

func TestWebSocketServerMultipleJobs(t *testing.T) {
	// Create a new WebSocket server
	server := NewServer()
	go server.Start()
	defer server.Stop()

	// Create a test HTTP server
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", func(c *gin.Context) {
		jobID := c.Query("job_id")
		server.HandleConnection(c.Writer, c.Request, jobID)
	})

	// Create a test server
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	// Connect two clients for different jobs
	ws1, _, err := websocket.DefaultDialer.Dial(wsURL+"?job_id=job1", nil)
	if err != nil {
		t.Fatalf("Failed to connect first client: %v", err)
	}
	defer ws1.Close()

	ws2, _, err := websocket.DefaultDialer.Dial(wsURL+"?job_id=job2", nil)
	if err != nil {
		t.Fatalf("Failed to connect second client: %v", err)
	}
	defer ws2.Close()

	// Create channels to receive messages
	messages1 := make(chan JobStatus)
	messages2 := make(chan JobStatus)

	go func() {
		for {
			var message JobStatus
			err := ws1.ReadJSON(&message)
			if err != nil {
				close(messages1)
				return
			}
			messages1 <- message
		}
	}()

	go func() {
		for {
			var message JobStatus
			err := ws2.ReadJSON(&message)
			if err != nil {
				close(messages2)
				return
			}
			messages2 <- message
		}
	}()

	// Send status updates for both jobs
	server.NotifyJobStatus("job1", "running", nil)
	server.NotifyJobStatus("job2", "completed", "done")

	// Wait for messages
	select {
	case msg := <-messages1:
		assert.Equal(t, "job_status", msg.Type)
		assert.Equal(t, "job1", msg.JobID)
		assert.Equal(t, "running", msg.Status)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for message from job1")
	}

	select {
	case msg := <-messages2:
		assert.Equal(t, "job_status", msg.Type)
		assert.Equal(t, "job2", msg.JobID)
		assert.Equal(t, "completed", msg.Status)
		assert.Equal(t, "done", msg.Result)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for message from job2")
	}
}
