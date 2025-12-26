import { API_BASE_URL, API_TIMEOUT_MS } from '../config/api';

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

export type RequestOptions = {
  method?: HttpMethod;
  headers?: Record<string, string>;
  body?: BodyInit | null;
  signal?: AbortSignal;
  timeoutMs?: number;
  skipAuth?: boolean;
  skipRetry?: boolean; // Skip 401 retry to avoid infinite loops
};

export class ApiError<T = unknown> extends Error {
  readonly status: number;
  readonly payload: T;

  constructor(message: string, status: number, payload: T) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.payload = payload;
  }
}

type TokenProvider = () => string | null;
type TokenSetter = (token: string | null) => void;
type UnauthorizedHandler = () => void;
type RefreshHandler = () => Promise<string | null>;

type ClientConfig = {
  baseUrl?: string;
  getAccessToken?: TokenProvider;
  setAccessToken?: TokenSetter;
  onUnauthorized?: UnauthorizedHandler;
  onRefresh?: RefreshHandler;
};

export class ApiClient {
  private baseUrl: string;
  private getAccessToken?: TokenProvider;
  private setAccessToken?: TokenSetter;
  private onUnauthorized?: UnauthorizedHandler;
  private onRefresh?: RefreshHandler;
  private refreshPromise: Promise<string | null> | null = null;

  constructor(config: ClientConfig = {}) {
    this.baseUrl = config.baseUrl ?? API_BASE_URL;
    this.getAccessToken = config.getAccessToken;
    this.setAccessToken = config.setAccessToken;
    this.onUnauthorized = config.onUnauthorized;
    this.onRefresh = config.onRefresh;
  }

  configure(config: ClientConfig) {
    if (config.baseUrl) this.baseUrl = config.baseUrl;
    if (config.getAccessToken) this.getAccessToken = config.getAccessToken;
    if (config.setAccessToken) this.setAccessToken = config.setAccessToken;
    if (config.onUnauthorized) this.onUnauthorized = config.onUnauthorized;
    if (config.onRefresh) this.onRefresh = config.onRefresh;
  }

  async request<T>(path: string, options: RequestOptions = {}): Promise<T> {
    const response = await this.executeRequest<T>(path, options);
    return response;
  }

  private async executeRequest<T>(path: string, options: RequestOptions): Promise<T> {
    const url = this.composeUrl(path);
    const headers = new Headers(options.headers);

    const body = options.body ?? null;
    if (body && !(body instanceof FormData) && !headers.has('Content-Type')) {
      headers.set('Content-Type', 'application/json');
    }

    if (!options.skipAuth && this.getAccessToken) {
      const token = this.getAccessToken();
      if (token) {
        headers.set('Authorization', `Bearer ${token}`);
      }
    }

    const controller = new AbortController();
    const timeout = window.setTimeout(() => controller.abort(), options.timeoutMs ?? API_TIMEOUT_MS);

    try {
      const response = await fetch(url, {
        method: options.method ?? 'GET',
        headers,
        body,
        credentials: 'include', // Important for cookies
        signal: options.signal ?? controller.signal,
      });

      const payload = await this.parsePayload(response);

      if (!response.ok) {
        // Handle 401 with auto-refresh
        if (response.status === 401 && !options.skipRetry && this.onRefresh) {
          const newToken = await this.handleTokenRefresh();
          if (newToken) {
            // Retry the original request with new token
            return this.executeRequest<T>(path, { ...options, skipRetry: true });
          }
          // Refresh failed, trigger unauthorized handler
          if (this.onUnauthorized) {
            this.onUnauthorized();
          }
        } else if (response.status === 401 && this.onUnauthorized) {
          this.onUnauthorized();
        }
        throw new ApiError('Request failed', response.status, payload);
      }
      return payload as T;
    } finally {
      window.clearTimeout(timeout);
    }
  }

  private async handleTokenRefresh(): Promise<string | null> {
    // Deduplicate concurrent refresh requests
    if (this.refreshPromise) {
      return this.refreshPromise;
    }

    if (!this.onRefresh) {
      return null;
    }

    this.refreshPromise = this.onRefresh()
      .then((token) => {
        if (token && this.setAccessToken) {
          this.setAccessToken(token);
        }
        return token;
      })
      .catch(() => {
        return null;
      })
      .finally(() => {
        this.refreshPromise = null;
      });

    return this.refreshPromise;
  }

  private composeUrl(path: string) {
    if (path.startsWith('http://') || path.startsWith('https://')) {
      return path;
    }
    return `${this.baseUrl.replace(/\/$/, '')}/${path.replace(/^\//, '')}`;
  }

  private async parsePayload(response: Response) {
    const contentType = response.headers.get('content-type') ?? '';
    if (contentType.includes('application/json')) {
      return response.json().catch(() => null);
    }
    return response.text();
  }
}

export const apiClient = new ApiClient();
