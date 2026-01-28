import { Router, Request, Response } from 'express';
import axios from 'axios';

const router = Router();
const GO_API_URL = 'http://localhost:8080';

// GET /api/analysis/symbol - Symbol analysis
router.get('/symbol', async (req: Request, res: Response) => {
  try {
    const { symbol } = req.query;

    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }

    const response = await axios.get(
      `${GO_API_URL}/api/analysis/symbol?symbol=${symbol}`
    );
    res.json(response.data);
  } catch (error: any) {
    console.error('Symbol analysis error:', error.message);
    res.status(500).json({ error: 'Failed to fetch symbol analysis' });
  }
});

// GET /api/analysis/report - Analysis report
router.get('/report', async (req: Request, res: Response) => {
  try {
    const { symbol } = req.query;

    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }

    const response = await axios.get(
      `${GO_API_URL}/api/analysis/report?symbol=${symbol}`
    );
    res.json(response.data);
  } catch (error: any) {
    console.error('Analysis report error:', error.message);
    res.status(500).json({ error: 'Failed to fetch analysis report' });
  }
});

export default router;
