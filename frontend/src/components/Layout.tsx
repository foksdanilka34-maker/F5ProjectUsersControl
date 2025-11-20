import { useState } from 'react';
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { ChevronDown, LogOut } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';

const navTabs = [
  { label: 'Обзор', path: '/' },
  { label: 'Проекты', path: '/projects' },
  { label: 'Сотрудники', path: '/admin/employees' },
  // { label: 'Аналитика', path: '/analytics' }, // Пока не реализовано
];

export default function Layout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [isProfileOpen, setIsProfileOpen] = useState(false);

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  return (
    <div className="relative min-h-screen bg-white text-gray-900">
      {/* Background effects */}
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -left-24 top-12 h-80 w-80 rounded-full bg-emerald-200/40 blur-3xl" />
        <div className="absolute -right-16 top-36 h-64 w-64 rounded-full bg-lime-200/40 blur-3xl" />
        <div className="absolute left-1/3 top-1/2 h-96 w-96 rounded-full bg-emerald-100/30 blur-3xl" />
      </div>

      <div className="relative z-10">
        <header className="sticky top-0 z-20 border-b border-emerald-50 bg-white/80 backdrop-blur-xl">
          <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
            <div className="flex items-center gap-8">
              <Link to="/" className="flex items-center gap-2">
                <span className="text-lg font-semibold tracking-tight text-emerald-600">КОМАНДА</span>
                <span className="rounded-xl bg-linear-to-r from-lime-400 to-emerald-400 px-3 py-1 text-sm font-bold text-gray-900">
                  F5
                </span>
              </Link>
              <nav className="hidden gap-4 md:flex">
                {navTabs.map((tab) => {
                  const isActive = location.pathname === tab.path || 
                                   (tab.path !== '/' && location.pathname.startsWith(tab.path));
                  return (
                    <Link
                      key={tab.path}
                      to={tab.path}
                      className={`rounded-full px-4 py-2 text-sm font-medium transition-all ${
                        isActive
                          ? 'bg-gray-900 text-white'
                          : 'text-gray-500 hover:text-emerald-600'
                      }`}
                    >
                      {tab.label}
                    </Link>
                  );
                })}
              </nav>
            </div>

            <div className="flex items-center gap-3">
              <div className="relative">
                <button
                  type="button"
                  onClick={() => setIsProfileOpen(!isProfileOpen)}
                  className="flex items-center gap-2 rounded-full border border-gray-200 bg-white px-3 py-1.5 text-sm hover:bg-gray-50 transition-colors"
                >
                  <div className="h-8 w-8 rounded-full bg-linear-to-r from-emerald-500 to-lime-500 flex items-center justify-center text-white font-bold">
                    {user?.role?.[0]?.toUpperCase() || 'U'}
                  </div>
                  <div className="text-left hidden sm:block">
                    <p className="text-sm font-medium text-gray-700">Пользователь</p>
                    <p className="text-[11px] uppercase tracking-wide text-gray-400">{user?.role}</p>
                  </div>
                  <ChevronDown className="h-4 w-4 text-gray-400" />
                </button>

                {isProfileOpen && (
                  <div className="absolute right-0 mt-2 w-48 rounded-xl border border-gray-100 bg-white py-1 shadow-lg animate-in fade-in slide-in-from-top-2">
                    <button
                      onClick={handleLogout}
                      className="flex w-full items-center gap-2 px-4 py-2 text-left text-sm text-rose-600 hover:bg-rose-50 transition-colors"
                    >
                      <LogOut className="h-4 w-4" />
                      Выйти
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>
        </header>

        <main className="mx-auto max-w-7xl px-6 py-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
