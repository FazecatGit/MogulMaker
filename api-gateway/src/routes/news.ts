import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';
import logger from '../utils/logger';

const router = Router();

// GET /api/news - Fetch news for trading positions and watchlist
router.get('/', async (req: Request, res: Response) => {
  try {
    logger.info('Fetching news for positions and watchlist');
    const data = await apiClient.get<any>('/api/news');
    logger.info('News fetched successfully', { count: (data as any)?.count || 0, symbols: (data as any)?.symbols_tracked || 0 });
    res.json(data);
  } catch (error: any) {
    logger.error('News fetch error', {
      message: error.message,
      status: error.response?.status,
      data: error.response?.data,
      stack: error.stack,
    });
    res.status(error.response?.status || 500).json({
      error: 'Failed to fetch news',
      details: error.response?.data || error.message,
    });
  }
});

export default router;
