import { useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import { apiClient } from '../lib/apiClient';
import { login as loginRequest, logout as logoutRequest, refreshSession as refreshRequest, type UserInfo } from '../services/authService';
import { AuthContext, type AuthContextValue } from './auth-context';

const ACCESS_TOKEN_STORAGE_KEY = 'f5_access_token';

const readStoredToken = () => {
  if (typeof window === 'undefined') {
    return null;
  }
  return window.localStorage.getItem(ACCESS_TOKEN_STORAGE_KEY);
};

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [accessToken, updateAccessToken] = useState<string | null>(() => readStoredToken());
  const [user, setUser] = useState<UserInfo | null>(null);
  const [refreshPending, setRefreshPendingState] = useState(false);
  const [initialized, setInitialized] = useState(false);
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

  const handleRefresh = useCallback(async (): Promise<string | null> => {
    try {
      const response = await refreshRequest();
      return response.access_token;
    } catch {
      return null;
    }
  }, []);

  useEffect(() => {
    apiClient.configure({
      getAccessToken: () => accessToken,
      setAccessToken: (token) => updateAccessToken(token),
      onUnauthorized: () => {
        updateAccessToken(null);
        setUser(null);
      },
      onRefresh: handleRefresh,
    });
  }, [accessToken, handleRefresh]);

  const setAccessToken = useCallback((token: string | null) => {
    updateAccessToken(token);
  }, []);

  const clearSession = useCallback(() => {
    updateAccessToken(null);
    setUser(null);
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
        setUser(response.user);
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

      if (response.user) {
        setUser(response.user);
      }
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
    if (!hydrated || refreshAttemptedRef.current) {
      return;
    }
    refreshAttemptedRef.current = true;

    const storedToken = readStoredToken();
    if (storedToken) {
      refreshSession()
        .catch(() => {

          clearSession();
        })
        .finally(() => {
          setInitialized(true);
        });
    } else {

      setInitialized(true);
    }
  }, [hydrated, refreshSession, clearSession]);

  // Proactive token refresh every 13 minutes (access token lives 15 min)
  useEffect(() => {
    if (!accessToken) return;

    const REFRESH_INTERVAL = 13 * 60 * 1000; // 13 minutes
    const intervalId = setInterval(() => {
      refreshSession().catch(() => {
        // silent fail — reactive refresh on 401 will handle it
      });
    }, REFRESH_INTERVAL);

    return () => clearInterval(intervalId);
  }, [accessToken, refreshSession]);

  const value = useMemo<AuthContextValue>(() => ({
    accessToken,
    user,
    isAuthenticated: Boolean(accessToken),
    refreshPending,
    initialized,
    setAccessToken,
    setRefreshPending,
    clearSession,
    loginWithCredentials,
    refreshSession,
    logout,
  }), [accessToken, user, initialized, clearSession, logout, loginWithCredentials, refreshPending, refreshSession, setAccessToken, setRefreshPending]);

  if (!hydrated) {
    return null;
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}


