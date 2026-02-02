import axios from 'axios';

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_GATEWAY_URL || 'http://localhost:3000/api',
  timeout: 30000, // Increased timeout to 30s for scanning
});

// Add JWT token to requests (only in browser)
apiClient.interceptors.request.use((config) => {
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('authToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

// Handle responses and errors
apiClient.interceptors.response.use(
  (response) => {
    // Log successful responses in development
    if (process.env.NODE_ENV === 'development') {
      console.log('[apiClient] Success:', response.config.url, 'Status:', response.status);
    }
    return response.data;
  },
  (error) => {
    const status = error.response?.status || error.code || 500;
    const message = error.response?.data?.error || error.message || 'An error occurred';

    // Handle 401 - clear token and redirect to login
    if (status === 401 && typeof window !== 'undefined') {
      localStorage.removeItem('authToken');
      window.location.href = '/login';
    }

    // Log error in development
    if (process.env.NODE_ENV === 'development') {
      console.error('[apiClient] Error:', {
        url: error.config?.url,
        status,
        message,
        data: error.response?.data,
      });
    }

    // Create a structured error object
    const apiError = new Error(message);
    Object.assign(apiError, {
      status,
      code: error.response?.data?.code || error.code || 'UNKNOWN_ERROR',
      details: error.response?.data?.details,
      requestId: error.response?.data?.requestId,
    });

    throw apiError;
  }
);

export default apiClient;