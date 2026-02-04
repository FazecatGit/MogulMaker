import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';

const router = Router();

// GET /api/backtest - Proxy backtest request
router.get('/', async (req: Request, res: Response) => {
  try {
    const { symbol, start_date, end_date } = req.query;

    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }

    // Default to last 60 days if dates not provided
    const endDate = end_date ? (end_date as string) : new Date().toISOString().split('T')[0];
    const startDate = start_date ? (start_date as string) : new Date(Date.now() - 60 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];

    const queryParams = new URLSearchParams();
    queryParams.append('symbol', symbol as string);
    queryParams.append('start_date', startDate);
    queryParams.append('end_date', endDate);

    const data = await apiClient.get(
      `/backtest?${queryParams.toString()}`
    );
    res.json(data);
  } catch (error: any) {
    console.error('Backtest error:', error.message);
    res.status(500).json({ error: 'Failed to run backtest' });
  }
});

// GET /api/backtest/results - Get cached results
router.get('/results', async (req: Request, res: Response) => {
  try {
    const { id } = req.query;

    if (!id) {
      res.status(400).json({ error: 'Backtest ID is required' });
      return;
    }

    const data = await apiClient.get(`api/backtest/results?id=${id}`);
    res.json(data);
  } catch (error: any) {
    console.error('Get results error:', error.message);
    res.status(500).json({ error: 'Failed to fetch results' });
  }
});

// GET /api/backtest/status - Get backtest status
router.get('/status', async (req: Request, res: Response) => {
  try {
    const { id } = req.query;

    if (!id) {
      res.status(400).json({ error: 'Backtest ID is required' });
      return;
    }

    const data = await apiClient.get(`/api/backtest/status?id=${id}`);
    res.json(data);
  } catch (error: any) {
    console.error('Backtest status error:', error.message);
    res.status(500).json({ error: 'Failed to fetch backtest status' });
  }
});

export default router;
