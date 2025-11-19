import { createContext } from 'react';

export type AuthContextValue = {
  accessToken: string | null;
  isAuthenticated: boolean;
  refreshPending: boolean;
  setAccessToken: (token: string | null) => void;
  setRefreshPending: (pending: boolean) => void;
  clearSession: () => void;
  loginWithCredentials: (payload: { login: string; password: string }) => Promise<void>;
  refreshSession: () => Promise<void>;
  logout: () => Promise<void>;
};

export const AuthContext = createContext<AuthContextValue | undefined>(undefined);
