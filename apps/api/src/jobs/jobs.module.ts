import { Module } from '@nestjs/common';
import { BullModule } from '@nestjs/bull';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { JobsController } from './jobs.controller';
import { JobsService } from './jobs.service';
import { JobsGateway } from './jobs.gateway';
import { JobsProcessor } from './jobs.processor';

/**
 * JobsModule - Handles all job processing functionality
 *
 * This module encapsulates all the components related to job processing:
 * 1. BullModule - Provides the job queue functionality using Redis
 * 2. JobsController - Handles HTTP requests for job creation
 * 3. JobsService - Contains business logic for job management
 * 4. JobsGateway - Handles WebSocket connections for real-time updates
 * 5. JobsProcessor - Processes jobs in the background
 */
@Module({
  imports: [
    // Configure Bull with Redis connection details from environment variables
    BullModule.forRootAsync({
      imports: [ConfigModule],
      useFactory: (configService: ConfigService) => ({
        redis: {
          host: configService.get('REDIS_HOST', 'redis'),
          port: parseInt(configService.get('REDIS_PORT', '6379')),
        },
      }),
      inject: [ConfigService],
    }),

    // Register the 'default' queue that will be used for job processing
    BullModule.registerQueue({
      name: 'default',
    }),
  ],

  // JobsController handles HTTP endpoints for job creation
  controllers: [JobsController],

  // Register all providers (services, gateways, processors)
  providers: [JobsService, JobsGateway, JobsProcessor],

  // Export JobsService to make it available to other modules if needed
  exports: [JobsService],
})
export class JobsModule {}
