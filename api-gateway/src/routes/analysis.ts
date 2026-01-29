import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';

const router = Router();

// GET /api/analysis/symbol - Symbol analysis
router.get('/symbol', async (req: Request, res: Response) => {
  try {
    const { symbol } = req.query;

    if (!symbol) {
      res.status(400).json({ error: 'Symbol is required' });
      return;
    }

    const data = await apiClient.get(
      `/api/analysis/symbol?symbol=${symbol}`
    );
    res.json(data);
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

    const data = await apiClient.get(
      `/api/analysis/report?symbol=${symbol}`
    );
    res.json(data);
  } catch (error: any) {
    console.error('Analysis report error:', error.message);
    res.status(500).json({ error: 'Failed to fetch analysis report' });
  }
});

export default router;
