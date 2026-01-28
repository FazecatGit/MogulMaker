import { Router, Request, Response } from 'express';
import axios from 'axios';

const router = Router();
const GO_API_URL = 'http://localhost:8080';

// GET /api/positions - Get all positions
router.get('/', async (req: Request, res: Response) => {
  try {
    const response = await axios.get(`${GO_API_URL}/api/positions`);
    res.json(response.data);
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

    const response = await axios.get(`${GO_API_URL}/api/positions/${symbol}`);
    res.json(response.data);
  } catch (error: any) {
    console.error('Get position error:', error.message);
    res.status(500).json({ error: 'Failed to fetch position' });
  }
});

// DELETE /api/positions/{symbol} - Close position (protected)
router.delete('/:symbol', async (req: Request, res: Response) => {
  try {
    const { symbol } = req.params;
    const token = req.headers.authorization?.split(' ')[1];

    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }

    const response = await axios.delete(
      `${GO_API_URL}/api/positions/${symbol}`,
      {
        headers: token ? { Authorization: `Bearer ${token}` } : {}
      }
    );
    res.json(response.data);
  } catch (error: any) {
    console.error('Close position error:', error.message);
    res.status(500).json({ error: 'Failed to close position' });
  }
});

export default router;
