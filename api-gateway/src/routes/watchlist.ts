import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';
import authMiddleware from '../middleware/auth';

const router = Router();

// GET /api/watchlist - Proxy to Go API
router.get('/', async (req: Request, res: Response) => {
  try {
    const data = await apiClient.get('/api/watchlist');
    res.json(data);
  } catch (error: any) {
    console.error('Watchlist fetch error:', error.message);
    res.status(500).json({ error: 'Failed to fetch watchlist' });
  }
});

// POST /api/watchlist - Add symbol to watchlist
router.post('/',authMiddleware ,async (req: Request, res: Response) => {
  try {
    const { symbol, reason } = req.body;

    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }

    const data = await apiClient.post('/api/watchlist', { 
      symbol, 
      reason: reason || '' 
    });
    res.json(data);
  } catch (error: any) {
    console.error('Add to watchlist error:', error.message);
    res.status(500).json({ error: 'Failed to add to watchlist' });
  }
});

// DELETE /api/watchlist - Remove symbol from watchlist
router.delete('/',authMiddleware ,async (req: Request, res: Response) => {
  try {
    const { symbol } = req.body;

    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }

    const data = await apiClient.delete('/api/watchlist', { symbol });
    res.json(data);
  } catch (error: any) {
    console.error('Remove from watchlist error:', error.message);
    res.status(500).json({ error: 'Failed to remove from watchlist' });
  }
});

export default router;
