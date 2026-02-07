module.exports = {
  ci: {
    collect: {
      startServerCommand: 'npm run build && npm run start',
      startServerReadyPattern: 'ready on',
      startServerReadyTimeout: 30000,
      url: [
        'http://localhost:3001/',
        'http://localhost:3001/dashboard',
        'http://localhost:3001/positions',
        'http://localhost:3001/trades',
        'http://localhost:3001/watchlist',
        'http://localhost:3001/backtest',
      ],
      numberOfRuns: 3,
      settings: {
        preset: 'desktop',
        // Optionally add mobile configuration
        // formFactor: 'mobile',
        // screenEmulation: {
        //   mobile: true,
        //   width: 375,
        //   height: 667,
        //   deviceScaleFactor: 2,
        // },
      },
    },
    assert: {
      preset: 'lighthouse:recommended',
      assertions: {
        'categories:performance': ['error', { minScore: 0.8 }],
        'categories:accessibility': ['error', { minScore: 0.9 }],
        'categories:best-practices': ['error', { minScore: 0.9 }],
        'categories:seo': ['error', { minScore: 0.9 }],
        
        // Custom performance budgets
        'resource-summary:script:size': ['warn', { maxNumericValue: 500000 }], // 500 KB
        'resource-summary:stylesheet:size': ['warn', { maxNumericValue: 100000 }], // 100 KB
        'resource-summary:image:size': ['warn', { maxNumericValue: 300000 }], // 300 KB
        
        // Core Web Vitals
        'first-contentful-paint': ['warn', { maxNumericValue: 2000 }], // 2s
        'largest-contentful-paint': ['warn', { maxNumericValue: 2500 }], // 2.5s
        'cumulative-layout-shift': ['warn', { maxNumericValue: 0.1 }],
        'total-blocking-time': ['warn', { maxNumericValue: 300 }], // 300ms
      },
    },
    upload: {
      target: 'temporary-public-storage',
      // For permanent storage, configure:
      // target: 'lhci',
      // serverBaseUrl: 'https://your-lhci-server.com',
      // token: 'your-build-token',
    },
  },
};
