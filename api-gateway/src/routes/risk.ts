import { Router, Request, Response } from 'express';
import axios from 'axios';

const router = Router();
const GO_API_URL = 'http://localhost:8080';

// GET /api/risk - Get risk status
router.get('/', async (req: Request, res: Response) => {
  try {
    const response = await axios.get(`${GO_API_URL}/api/risk`);
    res.json(response.data);
  } catch (error: any) {
    console.error('Get risk status error:', error.message);
    res.status(500).json({ error: 'Failed to fetch risk status' });
  }
});

export default router;
