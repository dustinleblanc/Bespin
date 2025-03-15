import { Process, Processor } from '@nestjs/bull';
import { Logger } from '@nestjs/common';
import { Job } from 'bull';
import { RandomTextJobData } from '../types';

@Processor('default')
export class JobsProcessor {
  private readonly logger = new Logger(JobsProcessor.name);

  constructor() {
    this.logger.log('Job processor initialized');
  }

  @Process('random-text')
  async processRandomTextJob(job: Job<RandomTextJobData>): Promise<string> {
    this.logger.log(`Processing random text job ${job.id} with length: ${job.data.length}`);

    try {
      // Generate random text
      const result = await this.generateRandomText(job.data.length);

      this.logger.log(`Completed job ${job.id} with result length: ${result.length}`);
      return result;
    } catch (error) {
      this.logger.error(`Error processing job ${job.id}:`, error);
      throw error;
    }
  }

  private async generateRandomText(length: number): Promise<string> {
    this.logger.log(`Generating random text of length: ${length}`);

    // Simulate processing time
    await new Promise((resolve) => setTimeout(resolve, 2000));

    // Generate random words
    const words = [
      'cloud', 'computing', 'platform', 'service', 'data',
      'storage', 'network', 'server', 'virtual', 'container',
      'function', 'application', 'microservice', 'kubernetes', 'docker',
      'infrastructure', 'code', 'deployment', 'scaling', 'monitoring'
    ];

    let result = '';
    for (let i = 0; i < length; i++) {
      const randomIndex = Math.floor(Math.random() * words.length);
      result += words[randomIndex] + ' ';
    }

    this.logger.log(`Generated random text of length: ${result.length} characters`);
    return result.trim();
  }
}
