import { Request, Response, NextFunction } from 'express';
import { APIError, formatError, handleError } from '../utils/errors';

function errorHandler(err: Error, req: Request, res: Response, next: NextFunction): void {
  if (err instanceof APIError) {
    const errorResponse = formatError(err);
    errorResponse.requestId = req.requestId;
    res.status(err.status).json(errorResponse);
    return;
  }
  
  // Use the handleError function from utils/errors
  const fallbackError = handleError(err);
  fallbackError.requestId = req.requestId;
  res.status(500).json(fallbackError);
}

export default errorHandler;