import { Process, Processor } from '@nestjs/bull';
import { Logger } from '@nestjs/common';
import { Job } from 'bull';
import { RandomTextJobData } from '../types';

/**
 * JobsProcessor - Processes jobs from the Bull queue
 *
 * This class is responsible for processing jobs that are added to the Bull queue.
 * It uses the @Processor decorator to register itself as a processor for the 'default' queue.
 *
 * The processor:
 * 1. Listens for jobs in the queue
 * 2. Processes jobs when they are available
 * 3. Returns results that can be used by other components (like the WebSocket gateway)
 */
@Processor('default')
export class JobsProcessor {
  // Create a logger instance for this processor
  private readonly logger = new Logger(JobsProcessor.name);

  /**
   * Constructor
   *
   * Initializes the processor and logs that it has been created.
   */
  constructor() {
    this.logger.log('Job processor initialized');
  }

  /**
   * Process random text generation jobs
   *
   * This method is decorated with @Process('random-text') to indicate that
   * it should process jobs of type 'random-text'.
   *
   * When a job is added to the queue with the 'random-text' type, this method
   * will be called to process it.
   *
   * @param job - The Bull job containing the job data
   * @returns A Promise that resolves to the generated random text
   */
  @Process('random-text')
  async processRandomTextJob(job: Job<RandomTextJobData>): Promise<string> {
    this.logger.log(`Processing random text job ${job.id} with length: ${job.data.length}`);

    try {
      // Generate random text based on the job data
      const result = await this.generateRandomText(job.data.length);

      this.logger.log(`Completed job ${job.id} with result length: ${result.length}`);
      return result;
    } catch (error) {
      // Log and rethrow any errors that occur during processing
      this.logger.error(`Error processing job ${job.id}:`, error);
      throw error;
    }
  }

  /**
   * Generate random text
   *
   * This private method generates a string of random words based on the specified length.
   * It simulates processing time with a delay to demonstrate asynchronous job processing.
   *
   * @param length - The number of words to generate
   * @returns A Promise that resolves to the generated random text
   */
  private async generateRandomText(length: number): Promise<string> {
    this.logger.log(`Generating random text of length: ${length}`);

    // Simulate processing time with a delay
    await new Promise((resolve) => setTimeout(resolve, 2000));

    // List of words to use for random text generation
    const words = [
      'cloud', 'computing', 'platform', 'service', 'data',
      'storage', 'network', 'server', 'virtual', 'container',
      'function', 'application', 'microservice', 'kubernetes', 'docker',
      'infrastructure', 'code', 'deployment', 'scaling', 'monitoring'
    ];

    // Generate the random text by selecting random words from the list
    let result = '';
    for (let i = 0; i < length; i++) {
      const randomIndex = Math.floor(Math.random() * words.length);
      result += words[randomIndex] + ' ';
    }

    this.logger.log(`Generated random text of length: ${result.length} characters`);
    return result.trim();
  }
}
