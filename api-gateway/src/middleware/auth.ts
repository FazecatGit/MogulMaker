import jwt from 'jsonwebtoken';
import { UnauthorizedError } from '../utils/errors';
import { Request, Response, NextFunction } from 'express';

interface Claims {
  userId: string;
  email: string;
  iat: number;
  exp: number;
  iss: string;
}

class JWTManager {
  private secretKey: string;

  constructor() {
    this.secretKey = process.env.JWT_SECRET_KEY || 'your-secret-key-change-this-in-production';
  }

  validateToken(tokenString: string): Claims {
    try {
      const claims = jwt.verify(tokenString, this.secretKey) as Claims;
      return claims;
    } catch (error: any) {
      throw new UnauthorizedError(`Invalid token: ${error.message}`);
    }
  }
}

const jwtManager = new JWTManager();

function authMiddleware(req: Request, res: Response, next: NextFunction): void {
  const token = req.headers.authorization?.split(' ')[1];
  
  if (!token) {
    throw new UnauthorizedError('Missing authorization token');
  }

  const claims = jwtManager.validateToken(token);
  (req as any).user = claims;
  next();
}

export default authMiddleware;
export { jwtManager, JWTManager };