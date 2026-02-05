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

// PUT /api/watchlist/refresh-scores - Recalculate all watchlist scores (must be before other routes)
router.put('/refresh-scores', async (req: Request, res: Response) => {
  try {
    logger.info('Refreshing all watchlist scores');
    const response = await apiClient.put(`/api/watchlist/refresh-scores`);
    apiClient.invalidateCache('/api/watchlist');
    logger.info('Watchlist scores refreshed', { response });
    res.json(response);
  } catch (error: any) {
    const status = error.response?.status || 500;
    const errorData = error.response?.data || {};
    const errorMessage = errorData?.error || errorData?.message || error.message || 'Failed to refresh scores';
    
    logger.error('Refresh scores error', {
      message: error.message,
      status: status,
      data: errorData,
    });
    
    res.status(status).json({
      error: errorMessage,
      details: errorData?.details || error.message,
    });
  }
});

// POST /api/watchlist - Add symbol to watchlist
router.post('/', async (req: Request, res: Response) => {
  try {
    const { symbol, reason, score } = req.body;

    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }
    logger.info('Adding to watchlist', { symbol, reason, score });
    const data = await apiClient.post('/api/watchlist', { 
      symbol, 
      score: score || 50,
      reason: reason || '' 
    });
    logger.info('Symbol added to watchlist', { symbol });
    res.json(data);
  } catch (error: any) {
    logger.error('Add to watchlist error', {
      message: error.message,
      status: error.response?.status,
      data: error.response?.data,
    });
    res.status(error.response?.status || 500).json({
      error: error.response?.data?.error || 'Failed to add to watchlist',
      details: error.response?.data?.details,
    });
  }
});

// DELETE /api/watchlist - Remove symbol from watchlist
router.delete('/', async (req: Request, res: Response) => {
  try {
    // Get symbol from query parameter first, then from body as fallback
    const symbol = (req.query.symbol as string) || req.body?.symbol;

    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }
    logger.info('Removing from watchlist', { symbol });
    // Call Go backend DELETE endpoint
    const response = await apiClient.delete(`/api/watchlist?symbol=${encodeURIComponent(symbol)}`);
    logger.info('Symbol removed from watchlist', { symbol, response });
    res.json(response);
  } catch (error: any) {
    const status = error.response?.status || 500;
    const errorData = error.response?.data || {};
    const errorMessage = errorData?.error || errorData?.message || error.message || 'Failed to remove from watchlist';
    
    logger.error('Remove from watchlist error', {
      message: error.message,
      status: status,
      data: errorData,
      symbol: (req.query.symbol as string) || req.body?.symbol,
    });
    
    res.status(status).json({
      error: errorMessage,
      details: errorData?.details || error.message,
    });
  }
});

// GET /api/watchlist/analyze - Analyze individual stock
router.get('/analyze', async (req: Request, res: Response) => {
  try {
    const symbol = req.query.symbol as string;

    if (!symbol) {
      res.status(400).json({ error: 'Symbol parameter is required' });
      return;
    }
    logger.info('Analyzing stock', { symbol });
    const data = await apiClient.get(`/api/watchlist/analyze?symbol=${symbol}`);
    logger.info('Stock analyzed successfully', { symbol });
    res.json(data);
  } catch (error: any) {
    logger.error('Stock analysis error', {
      message: error.message,
      status: error.response?.status,
      data: error.response?.data,
      symbol: req.query.symbol,
    });
    res.status(error.response?.status || 500).json({
      error: error.response?.data?.error || 'Failed to analyze stock',
      details: error.response?.data?.details,
    });
  }
});

export default router;
