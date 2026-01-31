import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';

const router = Router();

// POST /api/token - Generate JWT token
router.post('/', authMiddleware ,async (req: Request, res: Response) => {
  try {
    if (!req.body) {
      res.status(400).json({ error: 'Request body is required' });
      return;
    }

    const data = await apiClient.post('/api/token', req.body);
    res.json(data);
  } catch (error: any) {
    console.error('Token generation error:', error.message);
    res.status(500).json({ error: 'Failed to generate token' });
  }
});

export default router;
