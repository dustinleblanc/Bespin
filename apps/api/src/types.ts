/**
 * Common type definitions shared across the API
 *
 * This file contains TypeScript interfaces and types that are used throughout the application.
 * Centralizing types in a shared file helps maintain consistency and makes it easier to update
 * data structures across the application.
 */

/**
 * RandomTextJobData - Interface for random text generation job data
 *
 * This interface defines the structure of the data required for a random text generation job.
 * It's used by the Bull queue to ensure type safety when creating and processing jobs.
 *
 * Properties:
 * - length: The number of words to generate in the random text
 */
export interface RandomTextJobData {
  length: number;
}
