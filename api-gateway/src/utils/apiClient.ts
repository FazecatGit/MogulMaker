import axios, { AxiosInstance, AxiosError } from 'axios';

// Response cache with TTL (time-to-live)
interface CacheEntry {
  data: any;
  timestamp: number;
}

class APIClient {
  private client: AxiosInstance;
  private cache: Map<string, CacheEntry> = new Map();
  private readonly CACHE_DURATION = 60 * 1000;
  private readonly MAX_RETRIES = 3;
  private readonly RETRY_DELAY = 1000;

  constructor() {
    this.client = axios.create({
      baseURL: 'http://localhost:8080',
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add request interceptor for retry logic
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        const config = error.config;

        // Don't retry if no config or already retried max times
        if (!config || !config.retryCount) {
          config.retryCount = 0;
        }

        if (config.retryCount < this.MAX_RETRIES) {
          config.retryCount += 1;
          // Exponential backoff
          const delay = this.RETRY_DELAY * Math.pow(2, config.retryCount - 1);
          await new Promise((resolve) => setTimeout(resolve, delay));
          return this.client(config);
        }

        return Promise.reject(error);
      }
    );
  }

  /**
   * GET with caching
   * @param url - Endpoint
   * @param cacheDuration - Optional cache TTL in ms (default: 60s)
   */
  async get<T>(url: string, cacheDuration = this.CACHE_DURATION): Promise<T> {
    const cached = this.cache.get(url);
    if (cached && Date.now() - cached.timestamp < cacheDuration) {
      console.log(`[Cache HIT] ${url}`);
      return cached.data;
    }

    try {
      console.log(`[Cache MISS] ${url} - Fetching from Go API`);
      const response = await this.client.get<T>(url);
      
      this.cache.set(url, {
        data: response.data,
        timestamp: Date.now(),
      });

      return response.data;
    } catch (error) {
      this.handleError(error as AxiosError, 'GET', url);
      throw error;
    }
  }

  async post<T>(url: string, data?: any): Promise<T> {
    try {
      const response = await this.client.post<T>(url, data);
      this.invalidateCache(url);
      return response.data;
    } catch (error) {
      this.handleError(error as AxiosError, 'POST', url);
      throw error;
    }
  }


  async delete<T>(url: string, data?: any): Promise<T> {
    try {
      const response = await this.client.delete<T>(url, { data });
      this.invalidateCache(url);
      return response.data;
    } catch (error) {
      this.handleError(error as AxiosError, 'DELETE', url);
      throw error;
    }
  }

  invalidateCache(url?: string) {
    if (url) {
      this.cache.delete(url);
      console.log(`[Cache INVALIDATED] ${url}`);
    } else {
      this.cache.clear();
      console.log(`[Cache CLEARED] All entries`);
    }
  }

  private handleError(error: AxiosError, method: string, url: string) {
    if (error.response) {
      console.error(
        `[ERROR] ${method} ${url}: Status ${error.response.status} - ${error.response.statusText}`
      );
    } else if (error.request) {
      console.error(`[ERROR] ${method} ${url}: No response from server`);
    } else {
      console.error(`[ERROR] ${method} ${url}: ${error.message}`);
    }
  }
}

export default new APIClient();
