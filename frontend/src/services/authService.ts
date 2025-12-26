import { apiClient } from '../lib/apiClient';

export type LoginRequest = {
  login: string;
  password: string;
};

export type UserInfo = {
  id: number;
  login: string;
  full_name: string;
  role: string;
  avatar_url?: string;
};

export type LoginResponse = {
  access_token: string;
  user: UserInfo;
};

export type RefreshResponse = {
  access_token: string;
  user?: UserInfo;
};

export type MeResponse = UserInfo;

const AUTH_PREFIX = '/auth';

export function login(request: LoginRequest) {
  return apiClient.request<LoginResponse>(`${AUTH_PREFIX}/login`, {
    method: 'POST',
    body: JSON.stringify(request),
    skipAuth: true,
  });
}

// Refresh session using HttpOnly cookie (sent automatically)
export function refreshSession() {
  return apiClient.request<RefreshResponse>(`${AUTH_PREFIX}/refresh`, {
    method: 'POST',
    skipAuth: true,
    skipRetry: true, // Don't retry refresh on 401
  });
}

export function logout() {
  return apiClient.request<void>(`${AUTH_PREFIX}/logout`, {
    method: 'POST',
  });
}

export function getMe() {
  return apiClient.request<MeResponse>(`${AUTH_PREFIX}/me`);
}
