/**
 * Shared error codes used across api-gateway and frontend
 * Keep this file in sync between projects
 */

export const ERROR_CODES = {
  UNAUTHORIZED: 'UNAUTHORIZED',
  NOT_FOUND: 'NOT_FOUND',
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  BAD_REQUEST: 'BAD_REQUEST',
  FORBIDDEN: 'FORBIDDEN',
  CONFLICT: 'CONFLICT',
  INTERNAL_ERROR: 'INTERNAL_ERROR',
  TIMEOUT: 'TIMEOUT',
  SERVICE_UNAVAILABLE: 'SERVICE_UNAVAILABLE',
  GATEWAY_ERROR: 'GATEWAY_ERROR',
} as const;

export const HTTP_STATUS_CODES = {
  OK: 200,
  CREATED: 201,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  CONFLICT: 409,
  INTERNAL_ERROR: 500,
  SERVICE_UNAVAILABLE: 503,
  TIMEOUT: 504,
  GATEWAY_ERROR: 502,
} as const;

/**
 * Parse API error response from axios
 * @param error - Axios error object
 * @returns Formatted error with code and status
 */
export function parseAPIError(error: any) {
  const status = error.response?.status || 500;
  const data = error.response?.data || {};

  return {
    code: data.code || ERROR_CODES.INTERNAL_ERROR,
    message: data.error || 'An unexpected error occurred',
    status,
    details: data.details,
    requestId: data.requestId,
  };
}
