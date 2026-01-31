import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';
import authMiddleware from '../middleware/auth';
import logger from '../utils/logger';

const router = Router();

// POST /api/token - Generate JWT token
router.post('/', authMiddleware ,async (req: Request, res: Response) => {
  try {
    if (!req.body) {
      res.status(400).json({ error: 'Request body is required' });
      return;
    }
    logger.info('Generating token with payload:', req.body);
    const data = await apiClient.post('/api/token', req.body);
    logger.info('Token generated successfully');
    res.json(data);
  } catch (error: any) {
    console.error('Token generation error:', error.message);
    res.status(500).json({ error: 'Failed to generate token' });
  }
});

export default router;
