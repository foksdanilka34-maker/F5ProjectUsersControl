import { createContext } from 'react';
import type { UserInfo } from '../services/authService';

export type AuthContextValue = {
  accessToken: string | null;
  user: UserInfo | null;
  isAuthenticated: boolean;
  refreshPending: boolean;
  initialized: boolean;
  setAccessToken: (token: string | null) => void;
  setRefreshPending: (pending: boolean) => void;
  clearSession: () => void;
  loginWithCredentials: (payload: { login: string; password: string }) => Promise<void>;
  refreshSession: () => Promise<void>;
  logout: () => Promise<void>;
};

export const AuthContext = createContext<AuthContextValue | undefined>(undefined);


