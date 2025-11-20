import { apiClient, setAccessToken, unwrapResponse } from '../client';
import type {
  LoginRequest,
  LoginResponse,
  RefreshResponse,
  ApiResponse,
} from '../types';

export const authService = {
  /**
   * Authenticate user and get access + refresh tokens
   * Refresh token automatically stored in httpOnly cookie by backend
   */
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const response = await apiClient.post<ApiResponse<LoginResponse>>(
      '/api/v1/auth/login',
      credentials
    );
    
    const data = unwrapResponse(response.data);
    
    // Сохраняем access token в памяти
    if (data.access_token) {
      setAccessToken(data.access_token);
    }
    
    return data;
  },

  /**
   * Refresh access token using httpOnly cookie
   */
  async refresh(): Promise<RefreshResponse> {
    const response = await apiClient.post<ApiResponse<RefreshResponse>>(
      '/api/v1/auth/refresh',
      {}
    );
    
    const data = unwrapResponse(response.data);
    
    // Обновляем access token
    if (data.access_token) {
      setAccessToken(data.access_token);
    }
    
    return data;
  },

  /**
   * Logout user and invalidate refresh token
   */
  async logout(): Promise<void> {
    try {
      await apiClient.post<ApiResponse>('/api/v1/auth/logout');
    } finally {
      // Всегда очищаем токен локально, даже если запрос не удался
      setAccessToken(null);
    }
  },
};
