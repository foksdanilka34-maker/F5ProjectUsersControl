import { useState } from 'react';
import { Eye, EyeOff, Lock, Mail } from 'lucide-react';

export default function LoginForm() {
  const [showPassword, setShowPassword] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [focusedField, setFocusedField] = useState<string | null>(null);

  const handleSubmit = () => {
    if (email && password) {
      console.log('Login attempt:', { email, password });
      alert('Вход выполнен!');
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleSubmit();
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
          <p className="text-gray-400 text-sm mt-3">Войдите в личный кабинет</p>
        </div>

        {/* Login Card */}
        <div className="bg-gray-800/50 backdrop-blur-xl rounded-2xl shadow-2xl p-8 border border-gray-700/50 relative overflow-hidden">
          {/* Card gradient accent */}
          <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-emerald-500/10 to-transparent rounded-full blur-3xl"></div>
          
          <div className="space-y-6 relative">
            <h2 className="text-2xl font-bold text-white mb-6">Вход в систему</h2>

            {/* Email Field */}
            <div className="relative">
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Электронная почта
              </label>
              <div className="relative">
                <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  onFocus={() => setFocusedField('email')}
                  onBlur={() => setFocusedField(null)}
                  onKeyPress={handleKeyPress}
                  className={`w-full pl-12 pr-4 py-3.5 bg-gray-900/50 border rounded-xl text-white placeholder-gray-500 transition-all duration-300 focus:outline-none ${
                    focusedField === 'email'
                      ? 'border-emerald-400 shadow-lg shadow-emerald-500/20'
                      : 'border-gray-600 hover:border-gray-500'
                  }`}
                  placeholder="your@email.com"
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
                  onKeyPress={handleKeyPress}
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
              onClick={handleSubmit}
              className="w-full bg-gradient-to-r from-lime-400 to-emerald-400 hover:from-lime-500 hover:to-emerald-500 text-gray-900 font-bold py-4 px-6 rounded-xl transition-all duration-300 transform hover:scale-[1.02] hover:shadow-xl hover:shadow-emerald-500/30 active:scale-[0.98]"
            >
              Войти
            </button>
          </div>
        </div>

        {/* Footer Text */}
        <p className="text-center text-gray-500 text-xs mt-6">
          © 2025 Команда F5. Все права защищены.
        </p>
      </div>
    </div>
  );
}