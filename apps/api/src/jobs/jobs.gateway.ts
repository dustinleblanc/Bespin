import { WebSocketGateway, WebSocketServer } from '@nestjs/websockets';
import { Server } from 'socket.io';
import { OnModuleInit } from '@nestjs/common';
import { InjectQueue } from '@nestjs/bull';
import { Queue, Job } from 'bull';
import { RandomTextJobData } from '../types';
import { Logger } from '@nestjs/common';

@WebSocketGateway({
  cors: {
    origin: '*',
    methods: ['GET', 'POST', 'OPTIONS'],
    credentials: true,
  },
  transports: ['polling', 'websocket'],
})
export class JobsGateway implements OnModuleInit {
  @WebSocketServer()
  server: Server;

  private readonly logger = new Logger(JobsGateway.name);

  constructor(
    @InjectQueue('default') private readonly jobQueue: Queue<RandomTextJobData>,
  ) {
    this.logger.log('JobsGateway constructor called');
  }

  onModuleInit() {
    this.logger.log('WebSocket gateway initialized');
    this.logger.log(`Connected clients: ${this.server ? Object.keys(this.server.sockets.sockets).length : 'Server not initialized'}`);

    // Set up job completion listener
    this.jobQueue.on('completed', (job: Job<RandomTextJobData>, result: string) => {
      this.logger.log(`Job ${job.id} completed with result length: ${result.length}`);
      this.logger.log(`Connected clients at completion: ${Object.keys(this.server.sockets.sockets).length}`);

      try {
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

      socket.on('disconnect', () => {
        this.logger.log(`Client disconnected: ${socket.id}`);
        this.logger.log(`Remaining connected clients: ${Object.keys(this.server.sockets.sockets).length}`);
      });

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
