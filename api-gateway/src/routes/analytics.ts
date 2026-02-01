import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';

const router = Router();

// GET / - Portfolio summary (when mounted at /api/portfolio-summary)
router.get('/', async (req: Request, res: Response) => {
  try {
    const data = await apiClient.get('/api/portfolio-summary');
    res.json(data);
  } catch (error: any) {
    console.error('Portfolio summary error:', error.message);
    res.status(500).json({ error: 'Failed to fetch portfolio summary' });
  }
});

export default router;

