interface CacheEntry<T> {
  data: T;
  timestamp: number;
  ttl: number;
}

class Cache {
  private cache: Map<string, CacheEntry<any>> = new Map();
  private maxSize: number = 10000;

  constructor() {
    this.cleanupCache();
    this.maxCacheSize();
  }

  set<T>(key: string, data: T, ttlSeconds: number = 60): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl: ttlSeconds * 1000
    });
    console.log(`[Cache SET] ${key} (TTL: ${ttlSeconds}s)`);
  }

  get<T>(key: string): T | null {
    const cached = this.cache.get(key);
    if (!cached) {
      console.log(`[Cache MISS] ${key}`);
      return null;
    }

    const isExpired = Date.now() - cached.timestamp > cached.ttl;
    if (isExpired) {
      this.cache.delete(key);
      console.log(`[Cache EXPIRED] ${key}`);
      return null;
    }

    console.log(`[Cache HIT] ${key}`);
    return cached.data as T;
  }

  clear(key?: string): void {
    if (key) {
      this.cache.delete(key);
      console.log(`[Cache CLEARED] ${key}`);
    } else {
      this.cache.clear();
      console.log(`[Cache CLEARED] All entries`);
    }
  }

  size(): number {
    return this.cache.size;
  }

  private cleanupCache(intervalMs: number = 60000) {
  setInterval(() => {
    console.log('[Cache CLEANUP] Running cleanup task');
    const keysToDelete: string[] = [];
    const now = Date.now();
    this.cache.forEach((entry, key) => {
      if (now - entry.timestamp > entry.ttl) {
        keysToDelete.push(key);
      }
    });
    keysToDelete.forEach(key => {
      this.clear(key);
    });
    }, intervalMs);
}
    private maxCacheSize(intervalMs: number = 300000) {
    setInterval(() => {
        while (this.size() > this.maxSize) {
            let oldestKey: string | null = null;
            let oldestTimestamp = Infinity;

            this.cache.forEach((entry, key) => {
                if (entry.timestamp < oldestTimestamp) {
                    oldestTimestamp = entry.timestamp;
                    oldestKey = key;
                }
            });

            if (oldestKey) {
                this.clear(oldestKey);
            }
        }

    }, intervalMs);
    }
}

export default new Cache();