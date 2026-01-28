import { Router, Request, Response } from 'express';
import axios from 'axios';

const router = Router();
const GO_API_URL = 'http://localhost:8080';

// GET /api/trades - Get all trades
router.get('/', async (req: Request, res: Response) => {
  try {
    const response = await axios.get(`${GO_API_URL}/api/trades`);
    res.json(response.data);
  } catch (error: any) {
    console.error('Get trades error:', error.message);
    res.status(500).json({ error: 'Failed to fetch trades' });
  }
});

// POST /api/trades - Execute trade (protected)
router.post('/', async (req: Request, res: Response) => {
  try {
    const token = req.headers.authorization?.split(' ')[1];

    if (!token) {
      res.status(401).json({ error: 'Authorization required' });
      return;
    }

    const response = await axios.post(
      `${GO_API_URL}/api/trades`,
      req.body,
      {
        headers: { Authorization: `Bearer ${token}` }
      }
    );
    res.json(response.data);
  } catch (error: any) {
    console.error('Execute trade error:', error.message);
    res.status(500).json({ error: 'Failed to execute trade' });
  }
});

// POST /api/trades/sell-all - Sell all trades (protected)
router.post('/sell-all', async (req: Request, res: Response) => {
  try {
    const token = req.headers.authorization?.split(' ')[1];

    if (!token) {
      res.status(401).json({ error: 'Authorization required' });
      return;
    }

    const response = await axios.post(
      `${GO_API_URL}/api/trades/sell-all`,
      req.body,
      {
        headers: { Authorization: `Bearer ${token}` }
      }
    );
    res.json(response.data);
  } catch (error: any) {
    console.error('Sell all error:', error.message);
    res.status(500).json({ error: 'Failed to sell all trades' });
  }
});

export default router;
