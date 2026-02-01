import axios from 'axios';
import { parseAPIError } from '../../../shared/errorCodes';

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
    const parsedError = parseAPIError(error);
    
    // Handle 401 - clear token and redirect to login
    if (parsedError.status === 401) {
      localStorage.removeItem('authToken');
      window.location.href = '/login';
    }
    
    // Throw parsed error with structured format
    const errorToThrow = new Error(parsedError.message);
    Object.assign(errorToThrow, parsedError);
    throw errorToThrow;
  }
);

export default apiClient;