import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';
import logger from '../utils/logger';

const router = Router();

// GET /api/watchlist - Proxy to Go API
router.get('/', async (req: Request, res: Response) => {
  try {
    logger.info('Fetching watchlist from Go API');
    const data = await apiClient.get<any>('/api/watchlist');
    logger.info('Watchlist fetched successfully', { count: (data as any)?.watchlist?.length || 0 });
    res.json(data);
  } catch (error: any) {
    logger.error('Watchlist fetch error', {
      message: error.message,
      status: error.response?.status,
      data: error.response?.data,
      stack: error.stack,
    });
    res.status(error.response?.status || 500).json({
      error: 'Failed to fetch watchlist',
      details: error.response?.data || error.message,
    });
  }
});

// POST /api/watchlist - Add symbol to watchlist
router.post('/', async (req: Request, res: Response) => {
  try {
    const { symbol, reason } = req.body;

    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }
    logger.info('Adding to watchlist', { symbol, reason });
    const data = await apiClient.post('/api/watchlist', { 
      symbol, 
      reason: reason || '' 
    });
    logger.info('Symbol added to watchlist', { symbol });
    res.json(data);
  } catch (error: any) {
    console.error('Add to watchlist error:', error.message);
    res.status(500).json({ error: 'Failed to add to watchlist' });
  }
});

// DELETE /api/watchlist - Remove symbol from watchlist
router.delete('/', async (req: Request, res: Response) => {
  try {
    const { symbol } = req.body;

    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }
    logger.info('Removing from watchlist', { symbol });
    const data = await apiClient.delete('/api/watchlist', { symbol });
    logger.info('Symbol removed from watchlist', { symbol });
    res.json(data);
  } catch (error: any) {
    console.error('Remove from watchlist error:', error.message);
    res.status(500).json({ error: 'Failed to remove from watchlist' });
  }
});

export default router;
