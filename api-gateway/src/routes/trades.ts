import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';
import authMiddleware from '../middleware/auth';

const router = Router();

// GET /api/trades - Get all trades
router.get('/', async (req: Request, res: Response) => {
  try {
    const data = await apiClient.get('/api/trades');
    res.json(data);
  } catch (error: any) {
    console.error('Get trades error:', error.message);
    res.status(500).json({ error: 'Failed to fetch trades' });
  }
});

// POST /api/trades - Execute trade (protected)
router.post('/',authMiddleware ,async (req: Request, res: Response) => {
  try {
    const token = req.headers.authorization?.split(' ')[1];

    if (!token) {
      res.status(401).json({ error: 'Authorization required' });
      return;
    }

    const data = await apiClient.post('/api/trades', req.body);
    res.json(data);
  } catch (error: any) {
    console.error('Execute trade error:', error.message);
    res.status(500).json({ error: 'Failed to execute trade' });
  }
});

// POST /api/trades/sell-all - Sell all trades (protected)
router.post('/sell-all', authMiddleware ,async (req: Request, res: Response) => {
  try {
    const token = req.headers.authorization?.split(' ')[1];

    if (!token) {
      res.status(401).json({ error: 'Authorization required' });
      return;
    }

    const data = await apiClient.post('/api/trades/sell-all', req.body);
    res.json(data);
  } catch (error: any) {
    console.error('Sell all error:', error.message);
    res.status(500).json({ error: 'Failed to sell all trades' });
  }
});

export default router;
