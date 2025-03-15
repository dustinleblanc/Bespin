import { Controller, Post, Body, Get, Logger } from '@nestjs/common';
import { JobsService } from './jobs.service';
import { RandomTextJobDto } from './dto/random-text-job.dto';

/**
 * JobsController - Handles HTTP endpoints for job management
 *
 * This controller provides endpoints for:
 * 1. Testing the jobs service
 * 2. Creating new random text generation jobs
 *
 * The controller uses the JobsService to handle the business logic
 * and returns appropriate responses to the client.
 */
@Controller('jobs')
export class JobsController {
  // Create a logger instance for this controller
  private readonly logger = new Logger(JobsController.name);

  /**
   * Constructor with dependency injection
   *
   * The JobsService is injected to handle job creation and management.
   */
  constructor(private readonly jobsService: JobsService) {}

  /**
   * Test endpoint - GET /api/jobs/test
   *
   * A simple endpoint to verify that the jobs service is working.
   * Returns a JSON object with a message and timestamp.
   */
  @Get('test')
  test() {
    this.logger.log('Test endpoint called');
    return {
      message: 'API is working!',
      timestamp: new Date().toISOString(),
    };
  }

  /**
   * Create random text job endpoint - POST /api/jobs/random-text
   *
   * This endpoint accepts a request to create a new random text generation job.
   * It uses the JobsService to add the job to the queue and returns the job ID.
   *
   * The client can then listen for job completion events via WebSocket.
   *
   * @param randomTextJobDto - The DTO containing job parameters (length)
   * @returns An object containing the job ID
   */
  @Post('random-text')
  async createRandomTextJob(@Body() randomTextJobDto: RandomTextJobDto) {
    this.logger.log('Random text job endpoint called');
    this.logger.log(`Creating random text job with length: ${randomTextJobDto.length}`);

    // Create the job using the JobsService
    const job = await this.jobsService.createRandomTextJob(randomTextJobDto.length);

    this.logger.log(`Job created with ID: ${job.id}`);
    return { jobId: job.id };
  }
}
