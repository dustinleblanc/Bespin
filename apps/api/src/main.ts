import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';
import { ConfigService } from '@nestjs/config';
import { Logger } from '@nestjs/common';
import { Request, Response, NextFunction } from 'express';

async function bootstrap() {
  const logger = new Logger('Bootstrap');
  const app = await NestFactory.create(AppModule);
  const configService = app.get(ConfigService);

  // Set global prefix for all routes
  app.setGlobalPrefix('api');

  // Add request logging middleware
  app.use((req: Request, res: Response, next: NextFunction) => {
    logger.log(`[${req.method}] ${req.url}`);
    next();
  });

  // Enable CORS for our frontend
  app.enableCors({
    origin: '*', // Allow any origin for development
    methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
    allowedHeaders: ['Content-Type', 'Authorization'],
    credentials: true,
  });

  const port = configService.get<number>('PORT', 3001);
  await app.listen(port);

  logger.log(`Application is running on: ${await app.getUrl()}`);
  logger.log(`Accepting requests from: Any origin (CORS enabled)`);
  logger.log(`Socket.IO initialized with transports: polling, websocket`);
  logger.log(`All routes are prefixed with: /api`);
}
bootstrap();
