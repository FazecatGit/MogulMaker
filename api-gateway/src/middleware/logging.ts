import { Request, Response, NextFunction } from 'express';
import logger from '../utils/logger';

function loggingMiddleware(req: Request, res: Response, next: NextFunction): void {
  const startTime = Date.now();
  const requestId = req.requestId;

  logger.info(`${req.method} ${req.path}`, {
    requestId,
    ip: req.ip,
    userAgent: req.headers['user-agent'],
  });

  res.on('finish', () => {
    const duration = Date.now() - startTime;
    logger.info(`${req.method} ${req.path} - ${res.statusCode}`, {
      requestId,
      status: res.statusCode,
      duration: `${duration}ms`,
    });
  });

  next();
}

export default loggingMiddleware;