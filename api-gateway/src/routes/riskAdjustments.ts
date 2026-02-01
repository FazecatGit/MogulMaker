import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';

const router = Router();

// GET /api/risk-adjustments - Get risk adjustments and account balance
router.get('/', async (req: Request, res: Response) => {
  try {
    const data = await apiClient.get('/api/risk-adjustments');
    res.json(data);
  } catch (error: any) {
    console.error('Risk adjustments error:', error.message);
    res.status(500).json({ error: 'Failed to fetch risk adjustments' });
  }
});

export default router;
