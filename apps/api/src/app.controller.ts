import { Controller, Get } from '@nestjs/common';
import { AppService } from './app.service';

@Controller()
export class AppController {
  constructor(private readonly appService: AppService) {}

  @Get()
  getHello(): string {
    console.log('Root endpoint called');
    return this.appService.getHello();
  }

  @Get('test')
  getTest() {
    console.log('Test endpoint called');
    return {
      message: 'API is working!',
      timestamp: new Date().toISOString(),
    };
  }
}
