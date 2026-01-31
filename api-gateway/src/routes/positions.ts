import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';
import authMiddleware from '../middleware/auth';
import { symbol } from 'zod';
import logger from '../utils/logger';

const router = Router();

// GET /api/positions - Get all positions
router.get('/', async (req: Request, res: Response) => {
  try {
    const data = await apiClient.get('/api/positions');
    res.json(data);
  } catch (error: any) {
    console.error('Get positions error:', error.message);
    res.status(500).json({ error: 'Failed to fetch positions' });
  }
});

// GET /api/positions/{symbol} - Get position by symbol
router.get('/:symbol', async (req: Request, res: Response) => {
  try {
    const { symbol } = req.params;
    
    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }

    const data = await apiClient.get(`/api/positions/${symbol}`);
    res.json(data);
  } catch (error: any) {
    console.error('Get position error:', error.message);
    res.status(500).json({ error: 'Failed to fetch position' });
  }
});

// DELETE /api/positions/{symbol} - Close position (protected)
router.delete('/:symbol', authMiddleware, async (req: Request, res: Response, next) => {
  try {
    const { symbol } = req.params;

    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }

    logger.info('Closing position', { symbol });
    const data = await apiClient.delete(`/api/positions/${symbol}`);
    logger.info('Position closed successfully', { symbol });
    res.json(data);
  } catch (error) {
    next(error);
  }
});
export default router;
