import { Router, Request, Response } from 'express';
import axios from 'axios';

const router = Router();

router.get('/', async (req: Request, res: Response) => {
  try {
    const limit = req.query.limit || '15';
    const minScore = req.query.min_score || '50';
    const offset = req.query.offset || '0';
    
    console.log(`[Scout Route] Forwarding request with limit=${limit}, min_score=${minScore}, offset=${offset}`);
    
    const response = await axios.get('http://localhost:8080/api/scout', {
      params: {
        limit,
        min_score: minScore,
        offset
      },
      timeout: 30000
    });
    
    res.json(response.data);
  } catch (error: any) { 
    console.error('Scout fetch error:', error.message);
    if (error.response) {
      console.error('Backend response:', error.response.status, error.response.data);
    }
    res.status(500).json({ error: 'Failed to fetch scout data' });
  }
});

export default router;