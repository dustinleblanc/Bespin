import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { JobsModule } from './jobs/jobs.module';

/**
 * AppModule - The root module of the NestJS application
 *
 * In NestJS, modules are used to organize the application structure.
 * The AppModule is the entry point of the application and imports all other modules.
 *
 * This module:
 * 1. Imports the ConfigModule to handle environment variables
 * 2. Imports the JobsModule which contains job processing functionality
 * 3. Declares the AppController and AppService for basic API functionality
 */
@Module({
  imports: [
    // ConfigModule handles environment variables and configuration
    // Setting isGlobal to true makes the ConfigService available throughout the application
    ConfigModule.forRoot({
      isGlobal: true,
    }),

    // JobsModule contains all the job processing functionality
    // Including controllers, services, and WebSocket gateways
    JobsModule,
  ],

  // Controllers handle HTTP requests and define API endpoints
  controllers: [AppController],

  // Providers are services, repositories, factories, helpers, etc.
  providers: [AppService],
})
export class AppModule {}
