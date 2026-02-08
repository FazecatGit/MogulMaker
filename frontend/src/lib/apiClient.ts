import axios from 'axios';

// Rate limiting configuration - relaxed for development, strict for production
const RATE_LIMIT_CONFIG = {
  // In development (localhost), no rate limiting. In production, enforce limits.
  maxRequests: process.env.NODE_ENV === 'development' ? 1000 : 20, // Very high for dev, reasonable for prod
  windowMs: 1000, // 1 second window
  retryAfter: 60000, // Wait 60s if rate limited (prod only)
};

let requestTimestamps: number[] = [];
let rateLimitWaitUntil = 0;

/**
 * Check if we should rate limit this request
 */
function shouldRateLimit(): boolean {
  // Skip rate limiting entirely in development
  if (process.env.NODE_ENV === 'development') {
    return false;
  }

  const now = Date.now();
  
  // If we're in rate limit backoff, wait
  if (now < rateLimitWaitUntil) {
    return true;
  }
  
  // Remove old timestamps outside the window
  requestTimestamps = requestTimestamps.filter(ts => now - ts < RATE_LIMIT_CONFIG.windowMs);
  
  // Check if we've exceeded the limit
  if (requestTimestamps.length >= RATE_LIMIT_CONFIG.maxRequests) {
    rateLimitWaitUntil = now + RATE_LIMIT_CONFIG.retryAfter;
    return true;
  }
  
  // Record this request
  requestTimestamps.push(now);
  return false;
}

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_GATEWAY_URL || 'http://localhost:3000/api',
  timeout: 30000, // 30s timeout
});

// Request interceptor: check rate limits and add token
apiClient.interceptors.request.use((config) => {
  if (shouldRateLimit()) {
    const waitTime = Math.ceil((rateLimitWaitUntil - Date.now()) / 1000);
    const error = new Error(`Rate limited. Please wait ${waitTime}s before trying again.`);
    Object.assign(error, { status: 429, code: 'RATE_LIMITED' });
    return Promise.reject(error);
  }

  // Add JWT token to requests (only in browser)
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('authToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

// Response interceptor: handle errors with exponential backoff
let retryCount = 0;
const MAX_RETRIES = 3;

apiClient.interceptors.response.use(
  (response) => {
    // Reset retry count on success
    retryCount = 0;
    
    // Log successful responses in development
    if (process.env.NODE_ENV === 'development') {
      console.log('[apiClient] Success:', response.config.url, 'Status:', response.status);
    }
    return response.data;
  },
  async (error) => {
    const status = error.response?.status || error.code || 500;
    const errorData = error.response?.data || {};
    
    // Extract error message with multiple fallbacks
    let message = 'An error occurred';
    if (typeof errorData === 'object' && errorData.error) {
      message = errorData.error;
    } else if (typeof errorData === 'string') {
      message = errorData;
    } else if (error.message) {
      message = error.message;
    } else if (status === 400) {
      message = 'Invalid request';
    } else if (status === 404) {
      message = 'Resource not found';
    } else if (status === 429) {
      message = 'Too many requests. Please slow down.';
    } else if (status === 500) {
      message = 'Server error occurred';
    }

    // Retry logic for rate limiting (429) or server errors (5xx)
    if ((status === 429 || status >= 500) && retryCount < MAX_RETRIES) {
      retryCount++;
      // Exponential backoff: 1s, 2s, 4s
      const delayMs = Math.pow(2, retryCount - 1) * 1000;

      if (process.env.NODE_ENV === 'development') {
        console.warn(`[apiClient] Retry attempt ${retryCount}/${MAX_RETRIES} after ${delayMs}ms for ${error.config.url}`);
      }

      // Wait before retrying
      await new Promise(resolve => setTimeout(resolve, delayMs));

      // Reset rate limit trackers
      requestTimestamps = [];
      rateLimitWaitUntil = 0;

      return apiClient.request(error.config);
    }

    // Log full error info for debugging
    if (process.env.NODE_ENV === 'development') {
      console.error('[apiClient] Full Error Response:', {
        url: error.config?.url,
        status,
        message,
        errorData,
        errorMessage: error.message,
        errorResponse: error.response,
      });
    }

    // Handle 401 - clear token (but don't redirect, no login page exists)
    if (status === 401 && typeof window !== 'undefined') {
      localStorage.removeItem('authToken');
    }

    // Create a structured error object
    const apiError = new Error(message);
    Object.assign(apiError, {
      status,
      code: errorData?.code || error.code || 'UNKNOWN_ERROR',
      details: errorData?.details,
      requestId: errorData?.requestId,
    });

    throw apiError;
  }
);

export default apiClient;