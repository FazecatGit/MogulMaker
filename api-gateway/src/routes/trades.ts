import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';
import authMiddleware from '../middleware/auth';
import logger from '../utils/logger';

const router = Router();

// GET /api/trades - Get all trades
router.get('/', async (req: Request, res: Response, next) => {
  try {
    const data = await apiClient.get('/api/trades');
    res.json(data);
  } catch (error) {
    next(error);
  }
});

// GET /api/trades/statistics - Get trade statistics
router.get('/statistics', async (req: Request, res: Response, next) => {
  try {
    const data = await apiClient.get('/api/trades/statistics');
    res.json(data);
  } catch (error) {
    next(error);
  }
});

// POST /api/trades - Execute trade
router.post('/', async (req: Request, res: Response, next) => {
  try {
    logger.info('Executing trade', { symbol: req.body.symbol, quantity: req.body.quantity });
    const data = await apiClient.post('/api/trades', req.body);
    logger.info('Trade executed successfully', { orderId: (data as any).id });
    res.json(data);
  } catch (error) {
    next(error);
  }
});

// POST /api/trades/sell-all - Sell all trades
router.post('/sell-all', async (req: Request, res: Response, next) => {
  try {
    logger.info('Selling all trades');
    const data = await apiClient.post('/api/trades/sell-all', req.body);
    logger.info('All trades sold successfully', { count: (data as any).count });
    res.json(data);
  } catch (error) {
    next(error);
  }
});

export default router;