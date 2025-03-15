import { Controller, Get } from '@nestjs/common';
import { AppService } from './app.service';

/**
 * AppController - Handles basic API endpoints
 *
 * In NestJS, controllers are responsible for handling incoming requests and returning responses.
 * Controllers are decorated with @Controller() and contain methods decorated with HTTP method decorators
 * like @Get(), @Post(), etc.
 *
 * This controller provides:
 * 1. A root endpoint that returns a simple message
 * 2. A test endpoint that returns a JSON object with a message and timestamp
 */
@Controller()
export class AppController {
  /**
   * Constructor with dependency injection
   *
   * NestJS uses dependency injection to provide instances of services to controllers.
   * The 'private readonly' syntax is a TypeScript shorthand that creates and initializes a class property.
   */
  constructor(private readonly appService: AppService) {}

  /**
   * Root endpoint handler - GET /api
   *
   * This method handles GET requests to the root endpoint (/api).
   * It logs the request and returns a string from the AppService.
   */
  @Get()
  getHello(): string {
    console.log('Root endpoint called');
    return this.appService.getHello();
  }

  /**
   * Test endpoint handler - GET /api/test
   *
   * This method handles GET requests to the /api/test endpoint.
   * It logs the request and returns a JSON object with a message and timestamp.
   */
  @Get('test')
  getTest() {
    console.log('Test endpoint called');
    return {
      message: 'API is working!',
      timestamp: new Date().toISOString(),
    };
  }
}
