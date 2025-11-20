import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import { authService, setAccessToken, type UserRole } from '../api';
import { jwtDecode } from 'jwt-decode';

interface JWTPayload {
  user_id: string;
  role: UserRole;
  exp: number;
}

interface AuthContextType {
  user: { userId: string; role: UserRole } | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<{ userId: string; role: UserRole } | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Функция для принудительной очистки состояния аутентификации
  const clearAuthState = () => {
    setUser(null);
    setAccessToken(null);
  };

  // Делаем функцию доступной глобально для использования в interceptors
  useEffect(() => {
    (window as any).clearAuthState = clearAuthState;
  }, []);

  // При загрузке приложения пытаемся восстановить сессию через refresh token (в cookie)
  useEffect(() => {
    const initAuth = async () => {
      try {
        // Пытаемся обновить токен (refresh token в httpOnly cookie)
        const response = await authService.refresh();
        
        if (response.access_token) {
          const decoded = jwtDecode<JWTPayload>(response.access_token);
          setUser({ userId: decoded.user_id, role: decoded.role });
        }
      } catch (error) {
        // Нет валидной сессии - это нормально при первом запуске
        console.log('No active session');
      } finally {
        setIsLoading(false);
      }
    };

    initAuth();
  }, []);

  const login = async (username: string, password: string) => {
    try {
      const response = await authService.login({ login: username, password });
      
      // Декодируем JWT чтобы получить user_id и role
      const decoded = jwtDecode<JWTPayload>(response.access_token);
      
      setUser({ userId: decoded.user_id, role: decoded.role });
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    }
  };

  const logout = async () => {
    try {
      await authService.logout();
    } catch (error) {
      console.error('Logout failed:', error);
    } finally {
      setUser(null);
      setAccessToken(null);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
