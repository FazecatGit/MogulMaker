import { Router, Request, Response, NextFunction } from 'express';
import apiClient from '../utils/apiClient';
import logger from '../utils/logger';

const router = Router();

interface SettingsPayload {
  trading?: {
    maxDailyLoss: number;
    maxPositionRisk: number;
    maxOpenPositions: number;
    tradingHoursOnly: boolean;
    autoStopLoss: boolean;
    autoProfitTaking: boolean;
  };
  notifications?: {
    emailAlerts: boolean;
    tradeExecutionNotifications: boolean;
    riskAlerts: boolean;
    dailySummary: boolean;
    newsAlerts: boolean;
  };
  api?: {
    alpacaKey?: string;
    alpacaSecret?: string;
    finnhubKey?: string;
  };
}

// GET /api/settings - Get all settings
router.get('/', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const data = await apiClient.get('/api/settings');
    res.json(data);
  } catch (error) {
    next(error);
  }
});

// GET /api/settings/{key} - Get specific setting
router.get('/:key', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { key } = req.params;

    if (!key) {
      res.status(400).json({ error: 'Setting key is required' });
      return;
    }

    const data = await apiClient.get(`/api/settings/${key}`);
    res.json(data);
  } catch (error) {
    next(error);
  }
});

// POST /api/settings - Update settings
router.post('/', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const payload: SettingsPayload = req.body;

    if (!payload || Object.keys(payload).length === 0) {
      res.status(400).json({ error: 'Settings payload is required' });
      return;
    }

    logger.info('Updating settings', { payload });
    const data = await apiClient.post('/api/settings', payload);
    logger.info('Settings updated successfully');
    res.json(data);
  } catch (error) {
    next(error);
  }
});

// POST /api/settings/validate - Validate API credentials
router.post('/validate/credentials', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { alpacaKey, alpacaSecret, finnhubKey } = req.body;

    if (!alpacaKey || !alpacaSecret) {
      res.status(400).json({ error: 'Alpaca credentials are required for validation' });
      return;
    }

    logger.info('Validating API credentials');
    const data = await apiClient.post('/api/settings/validate/credentials', {
      alpacaKey,
      alpacaSecret,
      finnhubKey,
    });
    res.json(data);
  } catch (error) {
    next(error);
  }
});

export default router;
