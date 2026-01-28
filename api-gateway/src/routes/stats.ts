import { Router, Request, Response } from 'express';
import axios from 'axios';

const router = Router();
const GO_API_URL = 'http://localhost:8080';

// GET /api/stats - Get statistics
router.get('/', async (req: Request, res: Response) => {
  try {
    const response = await axios.get(`${GO_API_URL}/api/stats`);
    res.json(response.data);
  } catch (error: any) {
    console.error('Get stats error:', error.message);
    res.status(500).json({ error: 'Failed to fetch stats' });
  }
});

export default router;
