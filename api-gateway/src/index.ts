import express, { Express, Request, Response } from 'express';
import cors from 'cors';
import watchlistRoutes from './routes/watchlist';
import backtestRoutes from './routes/backtest';

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

// Error handling middleware
app.use((err: any, req: Request, res: Response) => {
  console.error('Error:', err);
  res.status(500).json({ error: 'Internal server error', message: err.message });
});

// Start server
app.listen(port, () => {
  console.log(`API Gateway listening on port ${port}`);
  console.log(`Go API running on http://localhost:8080`);
});
