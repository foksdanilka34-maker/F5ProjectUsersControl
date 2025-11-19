import { apiClient } from '../lib/apiClient';

export type LoginRequest = {
  login: string;
  password: string;
};

export type LoginResponse = {
  access_token: string;
  refresh_token?: string;
  expires_in?: number;
};

export type RefreshResponse = {
  access_token: string;
  refresh_token?: string;
  expires_in?: number;
};

const AUTH_PREFIX = '/auth';

export function login(request: LoginRequest) {
  return apiClient.request<LoginResponse>(`${AUTH_PREFIX}/login`, {
    method: 'POST',
    body: JSON.stringify(request),
    skipAuth: true,
  });
}

export function refreshSession() {
  return apiClient.request<RefreshResponse>(`${AUTH_PREFIX}/refresh`, {
    method: 'POST',
    skipAuth: true,
  });
}

export function logout() {
  return apiClient.request<void>(`${AUTH_PREFIX}/logout`, {
    method: 'POST',
  });
}
