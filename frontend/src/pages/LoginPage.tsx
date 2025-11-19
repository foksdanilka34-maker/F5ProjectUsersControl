import { useEffect, useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { AlertTriangle, CheckCircle2, Eye, EyeOff, Loader2, Lock, Mail } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import { ApiError } from '../lib/apiClient';

export default function LoginForm() {
  const [showPassword, setShowPassword] = useState(false);
  const [login, setLogin] = useState('');
  const [password, setPassword] = useState('');
  const [focusedField, setFocusedField] = useState<string | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const navigate = useNavigate();
  const { loginWithCredentials, isAuthenticated, refreshPending } = useAuth();

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/', { replace: true });
    }
  }, [isAuthenticated, navigate]);

  const handleSubmit = async (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault();
    setFormError(null);
    setSuccessMessage(null);

    if (!login || !password) {
      setFormError('Введите логин и пароль.');
      return;
    }

    setIsSubmitting(true);
    try {
      await loginWithCredentials({ login, password });
      setSuccessMessage('Авторизация успешна, перенаправляем на дашборд...');
      navigate('/', { replace: true });
    } catch (error) {
      if (error instanceof ApiError) {
        const payloadMessage =
          error.payload && typeof error.payload === 'object' && 'message' in error.payload
            ? String((error.payload as Record<string, unknown>).message)
            : undefined;
        setFormError(`Ошибка ${error.status}: ${payloadMessage ?? 'не удалось выполнить вход'}`);
      } else if (error instanceof Error) {
        setFormError(error.message);
      } else {
        setFormError('Не удалось выполнить вход. Попробуйте ещё раз.');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-emerald-950 to-gray-900 flex items-center justify-center p-4">
      {/* Animated background gradient orbs */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/4 -left-20 w-96 h-96 bg-emerald-500 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-pulse"></div>
        <div className="absolute top-1/3 -right-20 w-96 h-96 bg-green-500 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-pulse delay-1000"></div>
        <div className="absolute -bottom-32 left-1/3 w-96 h-96 bg-lime-500 rounded-full mix-blend-multiply filter blur-3xl opacity-10 animate-pulse delay-500"></div>
      </div>

      <div className="w-full max-w-md relative">
        {/* Logo/Brand */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center gap-3 mb-2">
            <div className="text-white text-3xl font-bold tracking-tight">КОМАНДА</div>
            <div className="bg-gradient-to-r from-lime-400 to-emerald-400 text-gray-900 px-4 py-2 rounded-lg font-bold text-xl">
              F5
            </div>
          </div>
          <p className="text-gray-400 text-sm mt-3">Войдите в личный кабинет</p>
        </div>

        {/* Login Card */}
        <div className="bg-gray-800/50 backdrop-blur-xl rounded-2xl shadow-2xl p-8 border border-gray-700/50 relative overflow-hidden">
          {/* Card gradient accent */}
          <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-emerald-500/10 to-transparent rounded-full blur-3xl"></div>
          
          <form className="space-y-6 relative" onSubmit={handleSubmit}>
            <h2 className="text-2xl font-bold text-white mb-6">Вход в систему</h2>
            {formError && (
              <div className="rounded-xl border border-rose-500/40 bg-rose-500/10 px-4 py-3 text-sm text-rose-100 flex items-start gap-2">
                <AlertTriangle className="h-5 w-5" />
                <span>{formError}</span>
              </div>
            )}
            {successMessage && (
              <div className="rounded-xl border border-emerald-500/40 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-100 flex items-start gap-2">
                <CheckCircle2 className="h-5 w-5" />
                <span>{successMessage}</span>
              </div>
            )}

            {/* Login Field */}
            <div className="relative">
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Логин
              </label>
              <div className="relative">
                <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type="text"
                  value={login}
                  onChange={(e) => setLogin(e.target.value)}
                  onFocus={() => setFocusedField('login')}
                  onBlur={() => setFocusedField(null)}
                  className={`w-full pl-12 pr-4 py-3.5 bg-gray-900/50 border rounded-xl text-white placeholder-gray-500 transition-all duration-300 focus:outline-none ${
                    focusedField === 'login'
                      ? 'border-emerald-400 shadow-lg shadow-emerald-500/20'
                      : 'border-gray-600 hover:border-gray-500'
                  }`}
                  placeholder="Введите ваш логин"
                />
              </div>
            </div>

            {/* Password Field */}
            <div className="relative">
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Пароль
              </label>
              <div className="relative">
                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  onFocus={() => setFocusedField('password')}
                  onBlur={() => setFocusedField(null)}
                  className={`w-full pl-12 pr-12 py-3.5 bg-gray-900/50 border rounded-xl text-white placeholder-gray-500 transition-all duration-300 focus:outline-none ${
                    focusedField === 'password'
                      ? 'border-emerald-400 shadow-lg shadow-emerald-500/20'
                      : 'border-gray-600 hover:border-gray-500'
                  }`}
                  placeholder="••••••••"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400 hover:text-emerald-400 transition-colors"
                >
                  {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isSubmitting || refreshPending || !login || !password}
              className={`w-full rounded-xl py-4 px-6 font-bold transition-all duration-300 flex items-center justify-center gap-2 ${
                isSubmitting || refreshPending
                  ? 'bg-gray-600 text-gray-200 cursor-not-allowed'
                  : 'bg-gradient-to-r from-lime-400 to-emerald-400 text-gray-900 hover:from-lime-500 hover:to-emerald-500'
              }`}
            >
              {isSubmitting || refreshPending ? (
                <>
                  <Loader2 className="h-5 w-5 animate-spin" /> Выполняем вход...
                </>
              ) : (
                'Войти'
              )}
            </button>
            <p className="text-xs text-gray-400 text-center">
              POST /api/v1/auth/login отдаёт access_token, refresh реализован через HttpOnly cookie (POST /auth/refresh).
            </p>
          </form>
        </div>

        {/* Footer Text */}
        <p className="text-center text-gray-500 text-xs mt-6">
          © 2025 Команда F5. Все права защищены.
        </p>
      </div>
    </div>
  );
}