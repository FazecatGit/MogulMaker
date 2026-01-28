import express, { Express, Request, Response } from 'express';
import cors from 'cors';
import watchlistRoutes from './routes/watchlist';
import backtestRoutes from './routes/backtest';
import scoutRoutes from './routes/scout';
import positionsRoutes from './routes/positions';
import riskRoutes from './routes/risk';
import statsRoutes from './routes/stats';
import tradesRoutes from './routes/trades';
import analyticsRoutes from './routes/analytics';
import analysisRoutes from './routes/analysis';
import tokenRoutes from './routes/token';

const app: Express = express();
const port = process.env.PORT || 3000;

// Middleware
app.use(cors());
app.use(express.json());

// Health check endpoint
app.get('/health', (req: Request, res: Response) => {
  res.json({ status: 'API Gateway is running', timestamp: new Date() });
});

// Routes
app.use('/api/watchlist', watchlistRoutes);
app.use('/api/backtest', backtestRoutes);
app.use('/api/scout', scoutRoutes);
app.use('/api/positions', positionsRoutes);
app.use('/api/risk', riskRoutes);
app.use('/api/stats', statsRoutes);
app.use('/api/trades', tradesRoutes);
app.use('/api/portfolio-summary', analyticsRoutes);
app.use('/api/risk-adjustments', analyticsRoutes);
app.use('/api/performance-metrics', analyticsRoutes);
app.use('/api/risk-alerts', analyticsRoutes);
app.use('/api/analysis', analysisRoutes);
app.use('/api/token', tokenRoutes);

// Error handling middleware
app.use((err: any, req: Request, res: Response, next: any) => {
  console.error('Error:', err);
  res.status(500).json({ error: 'Internal server error', message: err.message });
});

// Start server
app.listen(port, () => {
  console.log(`API Gateway listening on port ${port}`);
  console.log(`Go API running on http://localhost:8080`);
});
