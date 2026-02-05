import axios from 'axios';

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_GATEWAY_URL || 'http://localhost:3000/api',
  timeout: 120000, // Increased timeout to 120s for long-running scans
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
    } else if (status === 500) {
      message = 'Server error occurred';
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