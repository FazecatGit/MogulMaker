import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';

const router = Router();

router.get('/', async (req: Request, res: Response) => {
  try {
    const data = await apiClient.get('/api/scout');
    res.json(data);
    } catch (error: any) { 
    console.error('Scout fetch error:', error.message);
    res.status(500).json({ error: 'Failed to fetch scout data' });
  }
});

export default router;