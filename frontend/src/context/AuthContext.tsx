import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import { apiClient } from '../lib/apiClient';
import { login as loginRequest, logout as logoutRequest, refreshSession as refreshRequest } from '../services/authService';
import { AuthContext, type AuthContextValue } from './auth-context';

const ACCESS_TOKEN_STORAGE_KEY = 'f5_access_token';

const readStoredToken = () => {
  if (typeof window === 'undefined') {
    return null;
  }
  return window.localStorage.getItem(ACCESS_TOKEN_STORAGE_KEY);
};

export function AuthProvider({ children }: { children: ReactNode }) {
  const [accessToken, updateAccessToken] = useState<string | null>(() => readStoredToken());
  const [refreshPending, setRefreshPendingState] = useState(false);
  const hydrated = typeof window !== 'undefined';
  const refreshAttemptedRef = useRef(false);

  useEffect(() => {
    if (!hydrated) return;
    if (accessToken) {
      window.localStorage.setItem(ACCESS_TOKEN_STORAGE_KEY, accessToken);
    } else {
      window.localStorage.removeItem(ACCESS_TOKEN_STORAGE_KEY);
    }
  }, [accessToken, hydrated]);

  useEffect(() => {
    apiClient.configure({
      getAccessToken: () => accessToken,
      onUnauthorized: () => {
        updateAccessToken(null);
      },
    });
  }, [accessToken]);

  const setAccessToken = useCallback((token: string | null) => {
    updateAccessToken(token);
  }, []);

  const clearSession = useCallback(() => {
    updateAccessToken(null);
    setRefreshPendingState(false);
  }, []);

  const setRefreshPending = useCallback((pending: boolean) => {
    setRefreshPendingState(pending);
  }, []);

  const loginWithCredentials = useCallback(
    async (payload: { login: string; password: string }) => {
      setRefreshPendingState(true);
      try {
        const response = await loginRequest(payload);
        updateAccessToken(response.access_token);
      } finally {
        setRefreshPendingState(false);
      }
    },
    [],
  );

  const refreshSession = useCallback(async () => {
    setRefreshPendingState(true);
    try {
      const response = await refreshRequest();
      updateAccessToken(response.access_token);
    } catch (error) {
      clearSession();
      throw error;
    } finally {
      setRefreshPendingState(false);
    }
  }, [clearSession]);

  const logout = useCallback(async () => {
    try {
      await logoutRequest();
    } finally {
      clearSession();
    }
  }, [clearSession]);

  useEffect(() => {
    if (!hydrated || accessToken || refreshAttemptedRef.current) {
      return;
    }
    refreshAttemptedRef.current = true;
    refreshSession().catch(() => {
      // Ignore initial refresh failures; user will log in manually
    });
  }, [hydrated, accessToken, refreshSession]);

  const value = useMemo<AuthContextValue>(() => ({
    accessToken,
    isAuthenticated: Boolean(accessToken),
    refreshPending,
    setAccessToken,
    setRefreshPending,
    clearSession,
    loginWithCredentials,
    refreshSession,
    logout,
  }), [accessToken, clearSession, logout, loginWithCredentials, refreshPending, refreshSession, setAccessToken, setRefreshPending]);

  if (!hydrated) {
    return null;
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
