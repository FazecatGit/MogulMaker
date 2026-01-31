import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';
import authMiddleware from '../middleware/auth';

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
router.delete('/:symbol', authMiddleware ,async (req: Request, res: Response) => {
  try {
    const { symbol } = req.params;
    const token = req.headers.authorization?.split(' ')[1];

    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }

    const data = await apiClient.delete(`/api/positions/${symbol}`);
    res.json(data);
  } catch (error: any) {
    console.error('Close position error:', error.message);
    res.status(500).json({ error: 'Failed to close position' });
  }
});

export default router;
