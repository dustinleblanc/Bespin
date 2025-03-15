import { Injectable, Logger } from '@nestjs/common';
import { InjectQueue } from '@nestjs/bull';
import { Queue, Job } from 'bull';
import { RandomTextJobData } from '../types';

@Injectable()
export class JobsService {
  private readonly logger = new Logger(JobsService.name);

  constructor(
    @InjectQueue('default') private readonly jobQueue: Queue<RandomTextJobData>,
  ) {
    // Redis error handling
    this.jobQueue.on('error', (error: Error) => {
      this.logger.error('Bull queue error:', error);
    });
  }

  async createRandomTextJob(length: number): Promise<Job<RandomTextJobData>> {
    this.logger.log(`Creating random text job with length: ${length}`);
    return this.jobQueue.add('random-text', { length });
  }
}
