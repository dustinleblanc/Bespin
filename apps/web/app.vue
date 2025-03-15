<!--
  Main Application Component

  This is the root component of the Vue.js application.
  It provides a simple UI for generating random text using the API.

  The component:
  1. Connects to the API server using Socket.IO
  2. Provides a button to create random text generation jobs
  3. Displays the generated text when the job is completed
  4. Shows error messages if something goes wrong
-->
<template>
  <div class="container mx-auto px-4 py-8">
    <h1 class="text-4xl font-bold text-center mb-8">Welcome to Bespin</h1>
    <p class="text-center text-gray-600 mb-8">Your cloud development platform</p>

    <!-- Job Creation Button -->
    <div class="max-w-md mx-auto">
      <button
        @click="createJob"
        :disabled="loading"
        class="w-full bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded disabled:opacity-50"
      >
        {{ loading ? 'Generating...' : 'Generate Random Text' }}
      </button>

      <!-- Results Display -->
      <div v-if="result" class="mt-8 p-4 bg-gray-100 rounded">
        <h2 class="font-bold mb-2">Generated Text:</h2>
        <p class="text-gray-700">{{ result }}</p>
      </div>

      <!-- Error Display -->
      <div v-if="error" class="mt-4 p-4 bg-red-100 text-red-700 rounded">
        {{ error }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { io, Socket } from 'socket.io-client';

// Reactive state variables
const loading = ref(false);  // Tracks if a job is in progress
const result = ref('');      // Stores the generated text result
const error = ref('');       // Stores any error messages
let socket: Socket | null = null;  // Socket.IO connection

/**
 * Initialize the application when the component is mounted
 *
 * This function:
 * 1. Tests the API connection
 * 2. Sets up the Socket.IO connection
 * 3. Configures event listeners for Socket.IO events
 */
onMounted(async () => {
  const config = useRuntimeConfig();
  const apiUrl = config.public.apiUrl;
  console.log('Connecting to API at:', apiUrl);

  // First, test if the API is accessible
  try {
    // Try the root endpoint first
    console.log('Fetching root endpoint:', apiUrl);
    const rootResponse = await fetch(apiUrl);
    const rootText = await rootResponse.text();
    console.log('Root API response:', rootText);

    // Then try the test endpoint
    const testUrl = `${apiUrl}/api/test`;
    console.log('Fetching test endpoint:', testUrl);
    const testResponse = await fetch(testUrl);
    console.log('Test response status:', testResponse.status);

    if (testResponse.ok) {
      const data = await testResponse.json();
      console.log('API test successful:', data);
    } else {
      console.error('API test failed:', testResponse.status);
      // Try to get the error message
      try {
        const errorText = await testResponse.text();
        console.error('Error response:', errorText);
        error.value = `API test failed: ${testResponse.status} - ${errorText}`;
      } catch (textError) {
        error.value = `API test failed: ${testResponse.status}`;
      }
      return;
    }
  } catch (err) {
    console.error('API test error:', err);
    error.value = `API test error: ${err instanceof Error ? err.message : 'Unknown error'}`;
    return;
  }

  // Create socket connection with simplified configuration
  socket = io(apiUrl, {
    transports: ['polling'], // Start with polling only
    reconnectionAttempts: 10,
    reconnectionDelay: 1000,
    timeout: 30000,
    forceNew: true,
    autoConnect: true
  });

  console.log('Socket.IO client initialized with options:', {
    url: apiUrl,
    transports: socket.io.opts.transports
  });

  // Socket connection event handlers
  socket.on('connect', () => {
    console.log('Connected to server with ID:', socket?.id);
    console.log('Transport used:', socket?.io?.engine?.transport?.name);
    error.value = '';
  });

  socket.on('connect_error', (err) => {
    console.error('Connection error:', err);
    error.value = 'Failed to connect to server: ' + err.message;
  });

  socket.on('error', (err) => {
    console.error('Socket error:', err);
    error.value = 'Socket error: ' + (err.message || 'Unknown error');
  });

  // Listen for reconnection attempts
  socket.io.on('reconnect_attempt', (attempt) => {
    console.log(`Reconnection attempt ${attempt}...`);
  });

  socket.io.on('reconnect', () => {
    console.log('Reconnected to server');
    error.value = '';
  });

  socket.io.on('reconnect_failed', () => {
    console.error('Failed to reconnect');
    error.value = 'Failed to reconnect to server after multiple attempts';
  });
});

/**
 * Clean up resources when the component is unmounted
 *
 * This function disconnects the Socket.IO connection to prevent memory leaks.
 */
onUnmounted(() => {
  if (socket) {
    socket.disconnect();
    socket = null;
  }
});

/**
 * Create a new random text generation job
 *
 * This function:
 * 1. Sends a request to the API to create a new job
 * 2. Sets up a listener for the job completion event
 * 3. Updates the UI based on the job result or error
 */
async function createJob() {
  if (!socket) {
    error.value = 'Not connected to server';
    return;
  }

  try {
    // Update UI state
    loading.value = true;
    error.value = '';
    result.value = '';

    // Send job request to API
    const response = await fetch(`${useRuntimeConfig().public.apiUrl}/api/jobs/random-text`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ length: 100 })
    });

    if (!response.ok) {
      throw new Error('Failed to create job');
    }

    // Get the job ID from the response
    const { jobId } = await response.json();
    console.log('Job created with ID:', jobId);

    // Listen for job completion
    console.log(`Setting up listener for job-completed:${jobId}`);
    socket.once(`job-completed:${jobId}`, (jobResult: string) => {
      console.log('Job completed event received:', jobResult);
      result.value = jobResult;
      loading.value = false;
    });
  } catch (err) {
    // Handle errors
    error.value = err instanceof Error ? err.message : 'An error occurred';
    loading.value = false;
  }
}
</script>
