import { useState, type FormEvent } from 'react';
import { Eye, EyeOff, Loader2, Lock, User } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';

export default function LoginPage() {
  const [showPassword, setShowPassword] = useState(false);
  const [login, setLogin] = useState('');
  const [password, setPassword] = useState('');
  const [formError, setFormError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { loginWithCredentials } = useAuth();

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setFormError(null);

    if (!login || !password) {
      setFormError('Введите логин и пароль');
      return;
    }

    setIsSubmitting(true);
    try {
      await loginWithCredentials({ login, password });

    } catch (error) {
      if (error instanceof Error) {
        setFormError('Неверный логин или пароль');
      } else {
        setFormError('Не удалось выполнить вход');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-white via-emerald-50/50 to-white flex items-center justify-center p-4">
      {}
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -left-24 top-12 h-80 w-80 rounded-full bg-emerald-200/40 blur-3xl" />
        <div className="absolute -right-16 top-36 h-64 w-64 rounded-full bg-lime-200/40 blur-3xl" />
        <div className="absolute left-1/3 bottom-0 h-96 w-96 rounded-full bg-emerald-100/30 blur-3xl" />
      </div>

      <div className="w-full max-w-md relative z-10">
        {}
        <div className="text-center mb-8">
          <div className="inline-flex items-center gap-2 mb-3">
            <span className="text-2xl font-semibold tracking-tight text-emerald-600">КОМАНДА</span>
            <span className="rounded-xl bg-gradient-to-r from-lime-400 to-emerald-400 px-3 py-1.5 text-lg font-bold text-gray-900">
              F5
            </span>
          </div>
          <p className="text-gray-500 text-sm">Войдите в систему</p>
        </div>

        {}
        <div className="rounded-3xl border border-gray-100 bg-white/95 p-8 shadow-[0_30px_80px_rgba(6,95,70,0.08)]">
          <form className="space-y-6" onSubmit={handleSubmit}>
            {formError && (
              <div className="rounded-2xl border border-red-100 bg-red-50 px-4 py-3 text-sm text-red-700">
                {formError}
              </div>
            )}

            {}
            <div>
              <label className="block text-sm font-medium text-gray-600 mb-2">
                Логин
              </label>
              <div className="relative">
                <User className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type="text"
                  value={login}
                  onChange={(e) => setLogin(e.target.value)}
                  className="w-full pl-12 pr-4 py-3 bg-white border border-gray-200 rounded-2xl text-gray-900 placeholder-gray-400 transition-colors focus:outline-none focus:border-emerald-400"
                  placeholder="Введите логин"
                  autoComplete="username"
                />
              </div>
            </div>

            {}
            <div>
              <label className="block text-sm font-medium text-gray-600 mb-2">
                Пароль
              </label>
              <div className="relative">
                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full pl-12 pr-12 py-3 bg-white border border-gray-200 rounded-2xl text-gray-900 placeholder-gray-400 transition-colors focus:outline-none focus:border-emerald-400"
                  placeholder="Введите пароль"
                  autoComplete="current-password"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400 hover:text-emerald-500 transition-colors"
                >
                  {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
            </div>

            {}
            <button
              type="submit"
              disabled={isSubmitting || !login || !password}
              className="w-full rounded-2xl py-3.5 px-6 font-semibold transition-all bg-gradient-to-r from-lime-400 to-emerald-400 text-gray-900 hover:shadow-[0_15px_40px_rgba(132,204,22,0.35)] disabled:opacity-60 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="h-5 w-5 animate-spin" />
                  Вход...
                </>
              ) : (
                'Войти'
              )}
            </button>
          </form>
        </div>

        {}
        <p className="text-center text-gray-400 text-xs mt-6">
          © {new Date().getFullYear()} Команда F5
        </p>
      </div>
    </div>
  );
}


