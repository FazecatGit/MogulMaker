import { Router, Request, Response } from 'express';
import axios from 'axios'; 

const router = Router();
const GO_API_URL = 'http://localhost:8080';

router.get('/', async (req: Request, res: Response) => {
  try {
    const response = await axios.get(`${GO_API_URL}/api/scout`);
    res.json(response.data);
    } catch (error: any) { 
    console.error('Scout fetch error:', error.message);
    res.status(500).json({ error: 'Failed to fetch scout data' });
  }
});

export default router;