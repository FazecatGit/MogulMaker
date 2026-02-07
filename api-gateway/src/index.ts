import dotenv from 'dotenv';
import express, { Express, Request, Response } from 'express';
import cors from 'cors';
import compression from 'compression';
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
import executeTradeRoutes from './routes/execute-trade';
import newsRoutes from './routes/news';
import settingsRoutes from './routes/settings';
import requestIdMiddleware from './middleware/requestId';
import errorHandler from './middleware/errorHandler';
import RateLimiter from './middleware/ratelimit';

// Load environment variables from .env file
dotenv.config();

const app: Express = express();
const port = process.env.PORT || 3000;

// Middleware - OrderMatters!
// 1. Enable gzip compression for responses
app.use(compression({ level: 6 }));

// 2. CORS
app.use(cors());

// 3. Body parser with larger limits
app.use(express.json({ limit: '50mb' }));
app.use(express.urlencoded({ limit: '50mb', extended: true }));

// 4. Rate limiter middleware - Much more lenient for a trading app
const rateLimiter = new RateLimiter({
  windowMs: 1000, // 1 second window
  maxRequests: 1000 // 1000 requests per second = 3600000 per hour (very generous)
});
app.use(rateLimiter.middleware());

// Middleware to assign request IDs
app.use(requestIdMiddleware);

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
app.use('/api/news', newsRoutes);
app.use('/api/portfolio-summary', analyticsRoutes);
app.use('/api/analysis', analysisRoutes);
app.use('/api/token', tokenRoutes);
app.use('/api/settings', settingsRoutes);
app.use('/api', executeTradeRoutes);

// Error handling middleware
app.use(errorHandler);

// Start server
app.listen(port, () => {
  console.log(`API Gateway listening on port ${port}`);
  console.log(`Go API running on http://localhost:8080`);
});
