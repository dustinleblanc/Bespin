import { Controller, Post, Body, Get, Logger } from '@nestjs/common';
import { JobsService } from './jobs.service';
import { RandomTextJobDto } from './dto/random-text-job.dto';

@Controller('jobs')
export class JobsController {
  private readonly logger = new Logger(JobsController.name);

  constructor(private readonly jobsService: JobsService) {}

  @Get('test')
  test() {
    this.logger.log('Test endpoint called');
    return {
      message: 'API is working!',
      timestamp: new Date().toISOString(),
    };
  }

  @Post('random-text')
  async createRandomTextJob(@Body() randomTextJobDto: RandomTextJobDto) {
    this.logger.log('Random text job endpoint called');
    this.logger.log(`Creating random text job with length: ${randomTextJobDto.length}`);

    const job = await this.jobsService.createRandomTextJob(randomTextJobDto.length);

    this.logger.log(`Job created with ID: ${job.id}`);
    return { jobId: job.id };
  }
}
