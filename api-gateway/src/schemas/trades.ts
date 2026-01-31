import { z } from 'zod';

export const CreateTradeSchema = z.object({
  symbol: z.string().min(1, 'Symbol required'),
  quantity: z.number().int().positive('Quantity must be positive'),
  orderType: z.enum(['market', 'limit', 'stop']).optional(),
  price: z.number().optional(),
});

export const SellAllTradesSchema = z.object({
  reason: z.string().optional(),
});

export type CreateTradeRequest = z.infer<typeof CreateTradeSchema>;
export type SellAllRequest = z.infer<typeof SellAllTradesSchema>;
