import { Request, Response, NextFunction } from 'express';
import { ValidationError } from '../utils/errors';

function validateRequest(schema: any) {
  return (req: Request, res: Response, next: NextFunction) => {
    const { error, value } = schema.validate(req.body);
    
    if (error) {
      const validationError = new ValidationError(
        'Invalid request data',
        'VALIDATION_ERROR',
        { details: error.details }
      );
      return next(validationError);
    }
    
    req.body = value;
    next();
  };
}

export default validateRequest;