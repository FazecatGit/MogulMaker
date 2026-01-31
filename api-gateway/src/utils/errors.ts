interface ErrorEntry {
  error: string;
  code: string;
  status: number;
  timestamp: number;
  requestId?: string;
  traceId?: string;
  details?: Record<string, any>;
}

class APIError extends Error implements ErrorEntry {
  code: string;
  status: number;
  timestamp: number;
  requestId?: string | undefined;
    traceId?: string | undefined;
    details?: Record<string, any> | undefined;
    cause?: string | undefined;
  
  get error(): string {
    return this.message;
  }
  
constructor(
  message: string, 
  statusCode: number, 
  code: string,
  options?: {
    requestId?: string;
    traceId?: string;
    details?: Record<string, any>;
    cause?: string;
  }
) {
  super(message);
  this.code = code;
  this.status = statusCode;
  this.timestamp = Date.now();
  
  if (options) {
    this.requestId = options.requestId;
    this.traceId = options.traceId;
    this.details = options.details;
    this.cause = options.cause;
        }
    }
}

class ValidationError extends APIError {
    constructor(
      message: string, 
      code: string = 'VALIDATION_ERROR',
      options?: {
        requestId?: string;
        traceId?: string;
        details?: Record<string, any>;
        cause?: string;
      }
    ) {
      super(message, 400, code, options); 
      this.name = 'ValidationError';
    }
}

class NotFoundError extends APIError {
    constructor(message: string, 
        code: string = 'NOT_FOUND',
        options?: {
            requestId?: string;
            traceId?: string;
            details?: Record<string, any>;
            cause?: string;
          }
    ) {
    super(message, 404, code, options);
    this.name = 'NotFoundError';
    }
}

class UnauthorizedError extends APIError {
    constructor(message: string, 
        code: string = 'UNAUTHORIZED',
        options?: {
            requestId?: string;
            traceId?: string;
            details?: Record<string, any>;
            cause?: string;
        }
    ) {
    super(message, 401, code, options);
    this.name = 'UnauthorizedError';
    }
}

class ServerError extends APIError {
    constructor(message: string, code: string = 'SERVER_ERROR', options?: {
        requestId?: string;
        traceId?: string;
        details?: Record<string, any>;
        cause?: string;
    }) {
    super(message, 500, code, options);
    this.name = 'ServerError';
    }
}

class GoAPIError extends APIError {
    constructor(message: string, statusCode: number, code: string, options?: {
        requestId?: string;
        traceId?: string;
        details?: Record<string, any>;
        cause?: string;
    }) {
    super(message, statusCode, code, options);
    this.name = 'GoAPIError';
    }
}

class TimeoutError extends APIError {
    constructor(message: string, code: string = 'TIMEOUT_ERROR', options?: {
        requestId?: string;
        traceId?: string;
        details?: Record<string, any>;
        cause?: string;
    }) {
    super(message, 504, code, options);
    this.name = 'TimeoutError';
    }
}

class GatewayError extends APIError {
    constructor(message: string, code: string = 'GATEWAY_ERROR', options?: {
        requestId?: string;
        traceId?: string;
        details?: Record<string, any>;
        cause?: string;
    }) {
    super(message, 502, code, options);
    this.name = 'GatewayError';
    }
}

class ConflictError extends APIError {
    constructor(message: string, code: string = 'CONFLICT_ERROR', options?: {
        requestId?: string;
        traceId?: string;
        details?: Record<string, any>;
        cause?: string;
    }) {
    super(message, 409, code, options);
    this.name = 'ConflictError';
    }
}

class BadRequestError extends APIError {
    constructor(message: string, code: string = 'BAD_REQUEST', options?: {
        requestId?: string;
        traceId?: string;
        details?: Record<string, any>;
        cause?: string;
    }) {
    super(message, 400, code, options);
    this.name = 'BadRequestError';
    }
}

class ForbiddenError extends APIError {
    constructor(message: string, code: string = 'FORBIDDEN_ERROR', options?: {
        requestId?: string;
        traceId?: string;
        details?: Record<string, any>;
        cause?: string;
    }) {
    super(message, 403, code, options);
    this.name = 'ForbiddenError';
    }
}

class ServiceUnavailableError extends APIError {
    constructor(message: string, code: string = 'SERVICE_UNAVAILABLE', options?: {
        requestId?: string;
        traceId?: string;
        details?: Record<string, any>;
        cause?: string;
    }) {
    super(message, 503, code, options);
    this.name = 'ServiceUnavailableError';
    }
}

class ParsingError extends APIError {
    constructor(message: string, code: string = 'PARSING_ERROR', options?: {
        requestId?: string;
        traceId?: string;
        details?: Record<string, any>;
        cause?: string;
    }) {
    super(message, 500, code, options);
    this.name = 'ParsingError';
    }
}

class DataError extends APIError {
    constructor(message: string, code: string = 'DATA_ERROR', options?: {
        requestId?: string;
        traceId?: string;
        details?: Record<string, any>;
        cause?: string;
    }) {
    super(message, 422, code, options);
    this.name = 'DataError';
    }
}

class RateLimitError extends APIError {
    constructor(message: string, code: string = 'RATE_LIMIT_EXCEEDED', options?: {
        requestId?: string;
        traceId?: string;
        details?: Record<string, any>;
        cause?: string;
    }) {
    super(message, 429, code, options);
    this.name = 'RateLimitError';
    }
}

function formatError(error: APIError): ErrorEntry {
  return {
    error: error.message,
    code: error.code,
    status: error.status,
    timestamp: error.timestamp,
    requestId: error.requestId,
    traceId: error.traceId,
    details: error.details,
  };
}

// Generic handler for unknown errors
function isAPIError(error: unknown): error is APIError {
  return error instanceof APIError;
}

function handleError(error: unknown): ErrorEntry {
  if (isAPIError(error)) {
    return formatError(error);
  }
  return {
    error: 'Internal server error',
    code: 'INTERNAL_ERROR',
    status: 500,
    timestamp: Date.now(),
  };
}

export {
  APIError,
  ValidationError,
  NotFoundError,
  UnauthorizedError,
  ServerError,
  GoAPIError,
  TimeoutError,
  GatewayError,
  ConflictError,
  BadRequestError,
  ForbiddenError,
  ServiceUnavailableError,
  ParsingError,
  DataError,
  RateLimitError,
  formatError,
  isAPIError,
  handleError,
};