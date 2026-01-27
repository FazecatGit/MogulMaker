import { Router, Request, Response } from 'express';
import axios from 'axios';

const router = Router();
const GO_API_URL = 'http://localhost:8080';

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

    const response = await axios.get(
      `${GO_API_URL}/api/backtest?${queryParams.toString()}`
    );
    res.json(response.data);
  } catch (error: any) {
    console.error('Backtest error:', error.message);
    res.status(500).json({ error: 'Failed to run backtest' });
  }
});

// GET /api/backtest/results - Get cached results
router.get('/results', async (req: Request, res: Response) => {
  try {
    const response = await axios.get(`${GO_API_URL}/api/backtest/results`);
    res.json(response.data);
  } catch (error: any) {
    console.error('Get results error:', error.message);
    res.status(500).json({ error: 'Failed to fetch results' });
  }
});

export default router;
