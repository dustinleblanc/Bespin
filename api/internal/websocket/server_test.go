package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestWebSocketServer(t *testing.T) {
	// Create and start server
	server := NewServer()
	go server.Start()
	defer server.Stop()

	// Create test HTTP server
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleConnection(w, r, "test-job-id")
	}))
	defer httpServer.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")

	// Connect WebSocket client
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer ws.Close()

	// Create a channel to receive messages
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

	// Test different job statuses
	testCases := []struct {
		status string
		result interface{}
	}{
		{status: "pending", result: nil},
		{status: "running", result: map[string]interface{}{"progress": float64(50)}},
		{status: "completed", result: "test result"},
		{status: "failed", result: "error message"},
	}

	for _, tc := range testCases {
		t.Run(tc.status, func(t *testing.T) {
			// Send status update
			server.NotifyJobStatus("test-job-id", tc.status, tc.result)

			// Wait for message
			select {
			case msg := <-messages:
				assert.Equal(t, "job_status", msg.Type)
				assert.Equal(t, "test-job-id", msg.JobID)
				assert.Equal(t, tc.status, msg.Status)
				if tc.result != nil {
					assert.Equal(t, tc.result, msg.Result)
				}
			case <-time.After(time.Second):
				t.Fatalf("Timeout waiting for status update: %s", tc.status)
			}
		})
	}

	// Test multiple clients for the same job
	ws2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect second client: %v", err)
	}
	defer ws2.Close()

	messages2 := make(chan JobStatus)
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

	// Send a status update that both clients should receive
	server.NotifyJobStatus("test-job-id", "final", "done")

	// Wait for messages from both clients
	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-messages:
			assert.Equal(t, "job_status", msg1.Type)
			assert.Equal(t, "test-job-id", msg1.JobID)
			assert.Equal(t, "final", msg1.Status)
			assert.Equal(t, "done", msg1.Result)
		case msg2 := <-messages2:
			assert.Equal(t, "job_status", msg2.Type)
			assert.Equal(t, "test-job-id", msg2.JobID)
			assert.Equal(t, "final", msg2.Status)
			assert.Equal(t, "done", msg2.Result)
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for messages from both clients")
		}
	}
}

func TestWebSocketServerMultipleJobs(t *testing.T) {
	// Create and start server
	server := NewServer()
	go server.Start()
	defer server.Stop()

	// Create test HTTP server
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobID := r.URL.Query().Get("job_id")
		server.HandleConnection(w, r, jobID)
	}))
	defer httpServer.Close()

	// Helper function to create a client for a job
	createClient := func(jobID string) (*websocket.Conn, chan JobStatus) {
		wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "?job_id=" + jobID
		ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect client for job %s: %v", jobID, err)
		}

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

		return ws, messages
	}

	// Create clients for different jobs
	ws1, messages1 := createClient("job1")
	defer ws1.Close()
	ws2, messages2 := createClient("job2")
	defer ws2.Close()

	// Send status updates to different jobs
	server.NotifyJobStatus("job1", "running", nil)
	server.NotifyJobStatus("job2", "completed", "success")

	// Check that each client receives only its job's updates
	select {
	case msg := <-messages1:
		assert.Equal(t, "job_status", msg.Type)
		assert.Equal(t, "job1", msg.JobID)
		assert.Equal(t, "running", msg.Status)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for job1 message")
	}

	select {
	case msg := <-messages2:
		assert.Equal(t, "job_status", msg.Type)
		assert.Equal(t, "job2", msg.JobID)
		assert.Equal(t, "completed", msg.Status)
		assert.Equal(t, "success", msg.Result)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for job2 message")
	}
}
