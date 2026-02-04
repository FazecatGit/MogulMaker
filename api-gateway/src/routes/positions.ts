import { Router, Request, Response, NextFunction } from 'express';
import apiClient from '../utils/apiClient';
import authMiddleware from '../middleware/auth';
import { symbol } from 'zod';
import logger from '../utils/logger';

const router = Router();

// GET /api/positions - Get all positions
router.get('/', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const data = await apiClient.get('/api/positions');
    res.json(data);
  } catch (error) {
    next(error);
  }
});

// GET /api/positions/{symbol} - Get position by symbol
router.get('/:symbol', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { symbol } = req.params;
    
    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }

    const data = await apiClient.get(`/api/positions/${symbol}`);
    res.json(data);
  } catch (error) {
    next(error);
  }
});

// DELETE /api/positions/{symbol} - Close position
router.delete('/:symbol', async (req: Request, res: Response, next) => {
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
