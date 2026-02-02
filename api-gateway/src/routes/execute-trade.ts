import { Router, Request, Response } from 'express';
import apiClient from '../utils/apiClient';

const router = Router();

interface ExecuteTradeRequest {
  symbol: string;
  side: 'buy' | 'sell';
  quantity: number;
}

router.post('/execute-trade', async (req: Request, res: Response) => {
  try {
    const { symbol, side, quantity } = req.body as ExecuteTradeRequest;

    if (!symbol || !side || !quantity) {
      return res.status(400).json({ error: 'Missing required fields: symbol, side, quantity' });
    }

    if (!['buy', 'sell'].includes(side)) {
      return res.status(400).json({ error: 'Side must be either "buy" or "sell"' });
    }

    if (quantity <= 0) {
      return res.status(400).json({ error: 'Quantity must be greater than 0' });
    }

    // Call backend trade execution endpoint
    const data = await apiClient.post('/api/execute-trade', {
      symbol: symbol.toUpperCase(),
      side,
      quantity,
    });

    res.json(data);
  } catch (error: any) {
    console.error('Trade execution error:', error.message);
    
    if (error.response?.status === 404) {
      return res.status(404).json({ error: 'Trade execution endpoint not found on backend' });
    }
    
    res.status(error.response?.status || 500).json({
      error: error.response?.data?.error || 'Failed to execute trade',
    });
  }
});

export default router;
