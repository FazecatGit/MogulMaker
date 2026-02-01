import axios from 'axios';

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_GATEWAY_URL || 'http://localhost:3000/api',
  timeout: 10000,
});

// Add JWT token to requests
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('authToken');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle errors
apiClient.interceptors.response.use(
  (response) => response.data,
  (error) => {
    const status = error.response?.status || 500;
    const message = error.response?.data?.error || 'An error occurred';

    // Handle 401 - clear token and redirect to login
    if (status === 401) {
      localStorage.removeItem('authToken');
      window.location.href = '/login';
    }

    // Create a structured error object
    const apiError = new Error(message);
    Object.assign(apiError, {
      status,
      code: error.response?.data?.code || 'UNKNOWN_ERROR',
      details: error.response?.data?.details,
      requestId: error.response?.data?.requestId,
    });

    throw apiError;
  }
);

export default apiClient;