import { WebSocketGateway, WebSocketServer } from '@nestjs/websockets';
import { Server } from 'socket.io';
import { OnModuleInit } from '@nestjs/common';
import { InjectQueue } from '@nestjs/bull';
import { Queue, Job } from 'bull';
import { RandomTextJobData } from '../types';
import { Logger } from '@nestjs/common';

/**
 * JobsGateway - Handles WebSocket connections for real-time job updates
 *
 * This gateway:
 * 1. Sets up a WebSocket server using Socket.IO
 * 2. Listens for job completion events from the Bull queue
 * 3. Broadcasts job results to connected clients in real-time
 * 4. Handles client connections, disconnections, and errors
 *
 * The @WebSocketGateway decorator configures the WebSocket server with CORS settings
 * and transport options.
 */
@WebSocketGateway({
  cors: {
    origin: '*',
    methods: ['GET', 'POST', 'OPTIONS'],
    credentials: true,
  },
  transports: ['polling', 'websocket'],
})
export class JobsGateway implements OnModuleInit {
  /**
   * WebSocket server instance
   *
   * The @WebSocketServer decorator injects the Socket.IO server instance.
   * This server is used to emit events to connected clients.
   */
  @WebSocketServer()
  server: Server;

  // Create a logger instance for this gateway
  private readonly logger = new Logger(JobsGateway.name);

  /**
   * Constructor with dependency injection
   *
   * The Bull queue is injected to allow the gateway to listen for job events.
   *
   * @param jobQueue - The Bull queue for job processing
   */
  constructor(
    @InjectQueue('default') private readonly jobQueue: Queue<RandomTextJobData>,
  ) {
    this.logger.log('JobsGateway constructor called');
  }

  /**
   * Initialize the gateway when the module is initialized
   *
   * This method is called when the NestJS module is initialized.
   * It sets up event listeners for the Bull queue and Socket.IO server.
   */
  onModuleInit() {
    this.logger.log('WebSocket gateway initialized');
    this.logger.log(`Connected clients: ${this.server ? Object.keys(this.server.sockets.sockets).length : 'Server not initialized'}`);

    // Set up job completion listener
    this.jobQueue.on('completed', (job: Job<RandomTextJobData>, result: string) => {
      this.logger.log(`Job ${job.id} completed with result length: ${result.length}`);
      this.logger.log(`Connected clients at completion: ${Object.keys(this.server.sockets.sockets).length}`);

      try {
        // Emit the job completion event to all connected clients
        // The event name includes the job ID so clients can listen for specific jobs
        this.server.emit(`job-completed:${job.id}`, result);
        this.logger.log(`Emitted job-completed:${job.id} event`);
      } catch (error) {
        this.logger.error('Error emitting job completion event:', error);
      }
    });

    // Set up job failed listener
    this.jobQueue.on('failed', (job: Job<RandomTextJobData>, error: Error) => {
      this.logger.error(`Job ${job.id} failed:`, error);
    });

    // Socket connection handling
    this.server.on('connection', (socket) => {
      this.logger.log(`Client connected with ID: ${socket.id}`);
      this.logger.log(`Total connected clients: ${Object.keys(this.server.sockets.sockets).length}`);

      // Handle client disconnection
      socket.on('disconnect', () => {
        this.logger.log(`Client disconnected: ${socket.id}`);
        this.logger.log(`Remaining connected clients: ${Object.keys(this.server.sockets.sockets).length}`);
      });

      // Handle socket errors
      socket.on('error', (error: Error) => {
        this.logger.error('Socket error:', error);
      });

      // Log all incoming messages for debugging
      socket.onAny((event, ...args) => {
        this.logger.log(`Received event: ${event} with args: ${JSON.stringify(args)}`);
      });
    });
  }
}
