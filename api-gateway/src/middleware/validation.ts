import { Request, Response, NextFunction } from 'express';
import { ValidationError } from '../utils/errors';
import { z } from 'zod';

function validateRequest(schema: z.ZodSchema) {
  return (req: Request, res: Response, next: NextFunction) => {
    try {
      const validated = schema.parse(req.body);
      req.body = validated;
      next();
    } catch (error: any) {
      if (error instanceof z.ZodError) {
        const details = error.issues.map(err => ({
          path: err.path.join('.'),
          message: err.message,
        }));
        const validationError = new ValidationError(
          'Invalid request data',
          'VALIDATION_ERROR',
          { details }
        );
        return next(validationError);
      }
      next(error);
    }
  };
}

export default validateRequest;