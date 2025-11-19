import { API_BASE_URL, API_TIMEOUT_MS } from '../config/api';

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

export type RequestOptions = {
  method?: HttpMethod;
  headers?: Record<string, string>;
  body?: BodyInit | null;
  signal?: AbortSignal;
  timeoutMs?: number;
  skipAuth?: boolean;
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
type UnauthorizedHandler = () => void;

type ClientConfig = {
  baseUrl?: string;
  getAccessToken?: TokenProvider;
  onUnauthorized?: UnauthorizedHandler;
};

export class ApiClient {
  private baseUrl: string;
  private getAccessToken?: TokenProvider;
  private onUnauthorized?: UnauthorizedHandler;

  constructor(config: ClientConfig = {}) {
    this.baseUrl = config.baseUrl ?? API_BASE_URL;
    this.getAccessToken = config.getAccessToken;
    this.onUnauthorized = config.onUnauthorized;
  }

  configure(config: ClientConfig) {
    if (config.baseUrl) this.baseUrl = config.baseUrl;
    if (config.getAccessToken) this.getAccessToken = config.getAccessToken;
    if (config.onUnauthorized) this.onUnauthorized = config.onUnauthorized;
  }

  async request<T>(path: string, options: RequestOptions = {}): Promise<T> {
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
        credentials: 'include',
        signal: options.signal ?? controller.signal,
      });

      const payload = await this.parsePayload(response);
      if (!response.ok) {
        if (response.status === 401 && this.onUnauthorized) {
          this.onUnauthorized();
        }
        throw new ApiError('Request failed', response.status, payload);
      }
      return payload as T;
    } finally {
      window.clearTimeout(timeout);
    }
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
