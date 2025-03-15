import { IsNumber, IsOptional, Min } from 'class-validator';

/**
 * RandomTextJobDto - Data Transfer Object for random text job creation
 *
 * DTOs (Data Transfer Objects) are used to define the structure of data
 * that is sent to and from the API. They help with validation and type safety.
 *
 * This DTO is used for the POST /api/jobs/random-text endpoint.
 * It defines the parameters needed to create a random text generation job.
 */
export class RandomTextJobDto {
  /**
   * length - The number of words to generate
   *
   * This property uses class-validator decorators for validation:
   * - @IsNumber() - Ensures the value is a number
   * - @IsOptional() - Makes the field optional (will use default if not provided)
   * - @Min(1) - Ensures the value is at least 1
   *
   * The default value is 100 words.
   */
  @IsNumber()
  @IsOptional()
  @Min(1)
  length: number = 100;
}
