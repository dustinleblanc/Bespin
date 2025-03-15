import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';
import { ConfigService } from '@nestjs/config';
import { Logger } from '@nestjs/common';
import { Request, Response, NextFunction } from 'express';

/**
 * Bootstrap function to initialize and configure the NestJS application
 *
 * This is the entry point of the NestJS application. It:
 * 1. Creates a new NestJS application instance
 * 2. Configures middleware, CORS, and other settings
 * 3. Starts the HTTP server to listen for incoming requests
 */
async function bootstrap() {
  // Create a logger instance for bootstrap-related logs
  const logger = new Logger('Bootstrap');

  // Create a new NestJS application with the AppModule as the root module
  const app = await NestFactory.create(AppModule);

  // Get the ConfigService to access environment variables
  const configService = app.get(ConfigService);

  // Set global prefix for all routes - all endpoints will be prefixed with '/api'
  app.setGlobalPrefix('api');

  // Add request logging middleware to log all incoming HTTP requests
  app.use((req: Request, res: Response, next: NextFunction) => {
    logger.log(`[${req.method}] ${req.url}`);
    next();
  });

  // Enable CORS to allow cross-origin requests (important for development)
  app.enableCors({
    origin: '*', // Allow any origin for development
    methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
    allowedHeaders: ['Content-Type', 'Authorization'],
    credentials: true,
  });

  // Get the port from environment variables or use 3001 as default
  const port = configService.get<number>('PORT', 3001);

  // Start the HTTP server
  await app.listen(port);

  // Log application startup information
  logger.log(`Application is running on: ${await app.getUrl()}`);
  logger.log(`Accepting requests from: Any origin (CORS enabled)`);
  logger.log(`Socket.IO initialized with transports: polling, websocket`);
  logger.log(`All routes are prefixed with: /api`);
}

// Execute the bootstrap function to start the application
bootstrap();
