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

// GET /api/risk-alerts - Fetch risk alerts from backend
router.get('/alerts', async (req: Request, res: Response) => {
  try {
    const { limit } = req.query;
    
    const queryParams = new URLSearchParams();
    if (limit) {
      queryParams.append('limit', limit as string);
    }

    const queryString = queryParams.toString();
    const endpoint = queryString ? `/api/risk-alerts?${queryString}` : '/api/risk-alerts';
    
    const data = await apiClient.get(endpoint);
    res.json(data);
  } catch (error: any) {
    console.error('Risk alerts error:', error.message);
    res.status(500).json({ error: 'Failed to fetch risk alerts' });
  }
});

// GET /api/risk-adjustments - Get risk adjustments and account balance
router.get('/adjustments', async (req: Request, res: Response) => {
  try {
    const data = await apiClient.get('/api/risk-adjustments');
    res.json(data);
  } catch (error: any) {
    console.error('Risk adjustments error:', error.message);
    res.status(500).json({ error: 'Failed to fetch risk adjustments' });
  }
});

export default router;
