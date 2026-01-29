import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';

const router = Router();

// GET /api/stats - Get statistics
router.get('/', async (req: Request, res: Response) => {
  try {
    const data = await apiClient.get('/api/stats');
    res.json(data);
  } catch (error: any) {
    console.error('Get stats error:', error.message);
    res.status(500).json({ error: 'Failed to fetch stats' });
  }
});

export default router;
