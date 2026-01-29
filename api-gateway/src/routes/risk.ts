import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';

const router = Router();

// GET /api/risk - Get risk status
router.get('/', async (req: Request, res: Response) => {
  try {
    const data = await apiClient.get('/api/risk');
    res.json(data);
  } catch (error: any) {
    console.error('Get risk status error:', error.message);
    res.status(500).json({ error: 'Failed to fetch risk status' });
  }
});

export default router;
