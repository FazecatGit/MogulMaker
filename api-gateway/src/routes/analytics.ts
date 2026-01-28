import { Router, Request, Response } from 'express';
import axios from 'axios';

const router = Router();
const GO_API_URL = 'http://localhost:8080';

// GET /api/portfolio-summary - Portfolio summary
router.get('/portfolio-summary', async (req: Request, res: Response) => {
  try {
    const response = await axios.get(`${GO_API_URL}/api/portfolio-summary`);
    res.json(response.data);
  } catch (error: any) {
    console.error('Portfolio summary error:', error.message);
    res.status(500).json({ error: 'Failed to fetch portfolio summary' });
  }
});

// GET /api/risk-adjustments - Risk adjustments
router.get('/risk-adjustments', async (req: Request, res: Response) => {
  try {
    const response = await axios.get(`${GO_API_URL}/api/risk-adjustments`);
    res.json(response.data);
  } catch (error: any) {
    console.error('Risk adjustments error:', error.message);
    res.status(500).json({ error: 'Failed to fetch risk adjustments' });
  }
});

// GET /api/performance-metrics - Performance metrics
router.get('/performance-metrics', async (req: Request, res: Response) => {
  try {
    const response = await axios.get(`${GO_API_URL}/api/performance-metrics`);
    res.json(response.data);
  } catch (error: any) {
    console.error('Performance metrics error:', error.message);
    res.status(500).json({ error: 'Failed to fetch performance metrics' });
  }
});

// GET /api/risk-alerts - Risk alerts
router.get('/risk-alerts', async (req: Request, res: Response) => {
  try {
    const response = await axios.get(`${GO_API_URL}/api/risk-alerts`);
    res.json(response.data);
  } catch (error: any) {
    console.error('Risk alerts error:', error.message);
    res.status(500).json({ error: 'Failed to fetch risk alerts' });
  }
});

export default router;
