import { Request, Response, NextFunction } from 'express';
import { RateLimitError } from '../utils/errors'; 

interface RateLimitConfig {
  windowMs: number; // Time window in milliseconds
  maxRequests: number; // Max requests per window
}

class RateLimiter {
  private requests: Map<string, number[]> = new Map();
  private windowMs: number;
  private maxRequests: number;

  constructor(config: RateLimitConfig) {
    this.windowMs = config.windowMs;
    this.maxRequests = config.maxRequests;
  }

  check(identifier: string): boolean {
    const now = Date.now();
    const timestamps = this.requests.get(identifier) || [];
    const windowStart = now - this.windowMs;
    const recentTimestamps = timestamps.filter(ts => ts > windowStart);

    if(recentTimestamps.length===0){
        this.requests.delete(identifier);
    }else{
    this.requests.set(identifier, recentTimestamps);
    }
   
    if (recentTimestamps.length >= this.maxRequests) {
      return false;
    }


    recentTimestamps.push(now);
    this.requests.set(identifier, recentTimestamps);

    return true;
  }

  middleware() {
    return (req: Request, res: Response, next: NextFunction) => {
      const identifier = req.ip || req.requestId || 'unknown';
      
      if (!this.check(identifier)) {
        throw new RateLimitError('Rate limit exceeded');
      }
      next();
    };
  }
}

export default RateLimiter;