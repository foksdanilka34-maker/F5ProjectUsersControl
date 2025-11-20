import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios';
import type { ApiResponse } from './types';

// Base URL для API - можно вынести в .env
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Создаем axios instance
export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Важно! Для отправки httpOnly cookies (refresh_token)
});

// Храним access token в памяти (не в localStorage для безопасности)
let accessToken: string | null = null;

export const setAccessToken = (token: string | null) => {
  accessToken = token;
};

export const getAccessToken = () => accessToken;

// Request interceptor - добавляем access token в заголовок
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    if (accessToken && config.headers) {
      config.headers.Authorization = `Bearer ${accessToken}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor - обработка ошибок
apiClient.interceptors.response.use(
  (response) => {
    // Успешный ответ
    return response;
  },
  async (error: AxiosError<ApiResponse>) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // Проверяем, является ли запрос попыткой обновления токена
    const isRefreshRequest = originalRequest.url?.includes('/auth/refresh');
    // Проверяем, является ли запрос попыткой входа (чтобы не ретраить 401 при неверном пароле)
    const isLoginRequest = originalRequest.url?.includes('/auth/login');

    // Если 401 и это не повторный запрос И это не запрос на обновление токена И это не логин
    if (error.response?.status === 401 && !originalRequest._retry && !isRefreshRequest && !isLoginRequest) {
      originalRequest._retry = true;

      try {
        // Пытаемся обновить токен через refresh endpoint
        // Refresh token автоматически отправится через cookie благодаря withCredentials
        const response = await axios.post<ApiResponse<{ access_token: string; expires_in: number }>>(
          `${API_BASE_URL}/api/v1/auth/refresh`,
          {},
          { withCredentials: true }
        );

        if (response.data.success && response.data.data) {
          const newAccessToken = response.data.data.access_token;
          setAccessToken(newAccessToken);

          // Обновляем заголовок в исходном запросе
          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
          }

          // Повторяем исходный запрос с новым токеном
          return apiClient(originalRequest);
        }
      } catch (refreshError) {
        // Не удалось обновить токен - очищаем состояние и перенаправляем на логин
        console.error('Token refresh failed:', refreshError);
        setAccessToken(null);
        
        // Очищаем состояние аутентификации в AuthContext
        (window as any).clearAuthState?.();
        
        // Перенаправляем на страницу логина
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }
        
        return Promise.reject(refreshError);
      }
    }

    // Для всех остальных ошибок
    const errorMessage = error.response?.data?.error || error.message || 'Произошла ошибка';
    console.error('API Error:', errorMessage);
    
    return Promise.reject(error);
  }
);

// Вспомогательная функция для извлечения данных из ответа
export const unwrapResponse = <T>(response: ApiResponse<T>): T => {
  if (!response.success) {
    throw new Error(response.error || 'Request failed');
  }
  return response.data as T;
};
