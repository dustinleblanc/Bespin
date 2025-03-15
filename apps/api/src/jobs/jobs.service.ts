import { Injectable, Logger } from '@nestjs/common';
import { InjectQueue } from '@nestjs/bull';
import { Queue, Job } from 'bull';
import { RandomTextJobData } from '../types';

/**
 * JobsService - Handles job creation and management
 *
 * This service is responsible for:
 * 1. Creating new jobs and adding them to the Bull queue
 * 2. Handling job-related business logic
 * 3. Error handling for the job queue
 *
 * It uses Bull, a Redis-based queue for Node.js, to manage job processing.
 */
@Injectable()
export class JobsService {
  // Create a logger instance for this service
  private readonly logger = new Logger(JobsService.name);

  /**
   * Constructor with dependency injection
   *
   * The Bull queue is injected using the @InjectQueue decorator.
   * This allows the service to add jobs to the queue.
   *
   * @param jobQueue - The Bull queue for job processing
   */
  constructor(
    @InjectQueue('default') private readonly jobQueue: Queue<RandomTextJobData>,
  ) {
    // Set up error handling for the job queue
    this.jobQueue.on('error', (error: Error) => {
      this.logger.error('Bull queue error:', error);
    });
  }

  /**
   * Create a new random text generation job
   *
   * This method adds a new job to the Bull queue with the specified parameters.
   * The job will be processed by the JobsProcessor in the background.
   *
   * @param length - The number of words to generate
   * @returns A Promise that resolves to the created job
   */
  async createRandomTextJob(length: number): Promise<Job<RandomTextJobData>> {
    this.logger.log(`Creating random text job with length: ${length}`);

    // Add the job to the queue with the 'random-text' type
    // The job data includes the length parameter
    return this.jobQueue.add('random-text', { length });
  }
}
