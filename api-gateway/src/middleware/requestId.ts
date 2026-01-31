import { Request, Response, NextFunction } from 'express';
import { randomUUID } from 'crypto'; 

declare global {
  namespace Express {
    interface Request {
      requestId?: string;
    }
  }
}

function requestIdMiddleware(req: Request, res: Response, next: NextFunction): void {
  const requestId = `req-${Date.now()}-${randomUUID().slice(0, 12)}`;
    req.requestId = requestId;
    next();
}

export default requestIdMiddleware;