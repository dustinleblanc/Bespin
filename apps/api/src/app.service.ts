import { Injectable } from '@nestjs/common';

/**
 * AppService - Provides basic application functionality
 *
 * In NestJS, services are responsible for business logic and data access.
 * Services are decorated with @Injectable() to enable dependency injection.
 *
 * This service provides a simple method to return a greeting message.
 * In a real application, services would contain more complex business logic,
 * database operations, external API calls, etc.
 */
@Injectable()
export class AppService {
  /**
   * Returns a greeting message
   *
   * This is a simple method that returns a string.
   * It's used by the AppController to handle the root endpoint.
   */
  getHello(): string {
    return 'API server is running';
  }
}
