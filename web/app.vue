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
    <h1 class="text-4xl font-bold text-center mb-8">Bespin</h1>
    <p class="text-center text-gray-600 mb-8">Cloud Job Processing Platform</p>

    <div class="max-w-4xl mx-auto">
      <!-- API Test Section -->
      <div class="bg-white shadow-md rounded-lg p-6 mb-6">
        <h2 class="text-2xl font-bold mb-4">API Connection</h2>
        <div class="flex items-center mb-4">
          <button
            class="bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded"
            @click="testApi"
          >
            Test API Connection
          </button>
          <span v-if="testResponse" class="ml-4 text-green-500">Connected!</span>
          <span v-if="testError" class="ml-4 text-red-500">{{ testError }}</span>
        </div>
        <div v-if="testResponse" class="bg-gray-100 p-4 rounded">
          <pre>{{ JSON.stringify(testResponse, null, 2) }}</pre>
        </div>
      </div>

      <!-- Job Creation Section -->
      <div class="bg-white shadow-md rounded-lg p-6 mb-6">
        <h2 class="text-2xl font-bold mb-4">Create Random Text Job</h2>
        <div class="mb-4">
          <label class="block text-gray-700 mb-2">Text Length:</label>
          <input
            v-model="textLength"
            type="number"
            min="1"
            max="1000"
            class="border rounded py-2 px-3 w-full"
          />
        </div>
        <button
          class="bg-green-500 hover:bg-green-600 text-white font-bold py-2 px-4 rounded"
          @click="createRandomTextJob"
        >
          Create Job
        </button>
        <div v-if="jobError" class="mt-4 text-red-500">{{ jobError }}</div>
        <div v-if="jobResponse" class="mt-4 bg-gray-100 p-4 rounded">
          <pre>{{ JSON.stringify(jobResponse, null, 2) }}</pre>
        </div>
      </div>

      <!-- WebSocket Section -->
      <div class="bg-white shadow-md rounded-lg p-6">
        <h2 class="text-2xl font-bold mb-4">Job Updates</h2>
        <div class="flex items-center mb-4">
          <button
            v-if="!wsConnected"
            class="bg-purple-500 hover:bg-purple-600 text-white font-bold py-2 px-4 rounded"
            @click="connectWebSocket"
          >
            Connect to WebSocket
          </button>
          <button
            v-else
            class="bg-red-500 hover:bg-red-600 text-white font-bold py-2 px-4 rounded"
            @click="disconnectWebSocket"
          >
            Disconnect
          </button>
          <span v-if="wsConnected" class="ml-4 text-green-500">Connected to WebSocket</span>
        </div>
        <div v-if="jobUpdates.length === 0" class="text-gray-500">
          No job updates received yet. Connect to WebSocket and create a job to see updates.
        </div>
        <div v-else class="space-y-4">
          <div
            v-for="(update, index) in jobUpdates"
            :key="index"
            class="border-l-4 p-4 rounded"
            :class="{
              'border-yellow-500 bg-yellow-50': update.status === 'queued',
              'border-blue-500 bg-blue-50': update.status === 'processing',
              'border-green-500 bg-green-50': update.status === 'completed',
              'border-red-500 bg-red-50': update.status === 'failed',
            }"
          >
            <div class="flex justify-between">
              <h3 class="font-bold">Job {{ update.id }}</h3>
              <span class="text-sm">{{ new Date(update.updated_at).toLocaleString() }}</span>
            </div>
            <div class="mt-2">
              <span
                class="inline-block px-2 py-1 text-xs rounded-full"
                :class="{
                  'bg-yellow-200 text-yellow-800': update.status === 'queued',
                  'bg-blue-200 text-blue-800': update.status === 'processing',
                  'bg-green-200 text-green-800': update.status === 'completed',
                  'bg-red-200 text-red-800': update.status === 'failed',
                }"
              >
                {{ update.status }}
              </span>
            </div>
            <div v-if="update.result" class="mt-2">
              <div class="font-semibold">Result:</div>
              <div class="bg-white p-2 rounded mt-1 text-sm">
                {{
                  update.result.length > 100
                    ? update.result.substring(0, 100) + '...'
                    : update.result
                }}
              </div>
            </div>
            <div v-if="update.error" class="mt-2 text-red-600">
              <div class="font-semibold">Error:</div>
              <div>{{ update.error }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

// Define types for API responses
interface TestResponse {
  message: string
  timestamp?: string
}

interface Job {
  id: string
  type: string
  status: 'queued' | 'processing' | 'completed' | 'failed'
  data: {
    length?: number
  }
  result?: string
  error?: string
  created_at: string
  updated_at: string
}

interface JobUpdate {
  id: string
  status: 'queued' | 'processing' | 'completed' | 'failed'
  result?: string
  error?: string
  updated_at: string
}

const apiUrl = useRuntimeConfig().public.apiUrl
const testResponse = ref<TestResponse | null>(null)
const testError = ref<string | null>(null)
const jobResponse = ref<Job | null>(null)
const jobError = ref<string | null>(null)
const jobUpdates = ref<JobUpdate[]>([])
const textLength = ref(100)
const wsConnected = ref(false)
let ws: WebSocket | null = null

// Test API connection
async function testApi() {
  try {
    console.info('Testing API connection to:', `${apiUrl}/api/test`)
    testError.value = null
    const response = await fetch(`${apiUrl}/api/test`)
    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(`API test failed with status ${response.status}: ${errorText}`)
    }
    testResponse.value = await response.json()
    console.info('API test response:', testResponse.value)
  } catch (err) {
    console.error('API test error:', err)
    testError.value = err instanceof Error ? err.message : 'Unknown error'
    testResponse.value = null
  }
}

// Create a random text job
async function createRandomTextJob() {
  try {
    console.info('Creating random text job with length:', textLength.value)
    jobError.value = null
    const response = await fetch(`${apiUrl}/api/jobs/random-text`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ length: textLength.value }),
    })
    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(`Job creation failed with status ${response.status}: ${errorText}`)
    }
    jobResponse.value = await response.json()
    console.info('Job created:', jobResponse.value)
  } catch (err) {
    console.error('Job creation error:', err)
    jobError.value = err instanceof Error ? err.message : 'Unknown error'
    jobResponse.value = null
  }
}

// Connect to WebSocket for job updates
function connectWebSocket() {
  if (ws) {
    return
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${new URL(apiUrl).host}/api/ws/jobs`

  console.info('Connecting to WebSocket:', wsUrl)

  try {
    ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      console.info('WebSocket connected')
      wsConnected.value = true
    }

    ws.onmessage = (event) => {
      try {
        const update = JSON.parse(event.data) as JobUpdate
        console.info('WebSocket message received:', update)
        jobUpdates.value.unshift(update)
      } catch (err) {
        console.error('Error parsing WebSocket message:', err)
      }
    }

    ws.onclose = () => {
      console.info('WebSocket disconnected')
      wsConnected.value = false
      ws = null
    }

    ws.onerror = (event) => {
      console.error('WebSocket error:', event)
      wsConnected.value = false
    }
  } catch (err) {
    console.error('Error connecting to WebSocket:', err)
  }
}

// Disconnect from WebSocket
function disconnectWebSocket() {
  if (ws) {
    ws.close()
    ws = null
    wsConnected.value = false
  }
}

// Call testApi on component mount
onMounted(() => {
  testApi()
})

// Clean up WebSocket on component unmount
onUnmounted(() => {
  disconnectWebSocket()
})
</script>
