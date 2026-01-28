import { Router, Request, Response } from 'express';
import axios from 'axios';

const router = Router();
const GO_API_URL = 'http://localhost:8080';

// POST /api/token - Generate JWT token
router.post('/', async (req: Request, res: Response) => {
  try {
    if (!req.body) {
      res.status(400).json({ error: 'Request body is required' });
      return;
    }

    const response = await axios.post(
      `${GO_API_URL}/api/token`,
      req.body
    );
    res.json(response.data);
  } catch (error: any) {
    console.error('Token generation error:', error.message);
    res.status(500).json({ error: 'Failed to generate token' });
  }
});

export default router;
