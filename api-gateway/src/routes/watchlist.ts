import { Router, Request, Response } from 'express';
import axios from 'axios';

const router = Router();
const GO_API_URL = 'http://localhost:8080';

// GET /api/watchlist - Proxy to Go API
router.get('/', async (req: Request, res: Response) => {
  try {
    const response = await axios.get(`${GO_API_URL}/api/watchlist`);
    res.json(response.data);
  } catch (error: any) {
    console.error('Watchlist fetch error:', error.message);
    res.status(500).json({ error: 'Failed to fetch watchlist' });
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

    const response = await axios.post(
      `${GO_API_URL}/api/watchlist`,
      { symbol, reason: reason || '' }
    );
    res.json(response.data);
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

    const response = await axios.delete(
      `${GO_API_URL}/api/watchlist`,
      { data: { symbol } }
    );
    res.json(response.data);
  } catch (error: any) {
    console.error('Remove from watchlist error:', error.message);
    res.status(500).json({ error: 'Failed to remove from watchlist' });
  }
});

export default router;
