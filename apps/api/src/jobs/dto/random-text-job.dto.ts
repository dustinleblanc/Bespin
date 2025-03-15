import { IsNumber, IsOptional, Min } from 'class-validator';

export class RandomTextJobDto {
  @IsNumber()
  @IsOptional()
  @Min(1)
  length: number = 100;
}
