// middleware/logging.ts
import { Request, Response, NextFunction } from 'express';

function loggingMiddleware(req: Request, res: Response, next: NextFunction): void {
  const startTime = Date.now();
  
  const requestId = req.requestId;

  console.log(`[${requestId}] ${req.method} ${req.path}`);
  
  // Capture response time
  res.on('finish', () => {
    const duration = Date.now() - startTime;
    console.log(`[${requestId}] ${res.statusCode} ${duration}ms`);
  });
  
  next();
}

export default loggingMiddleware;