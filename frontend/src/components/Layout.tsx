import { useEffect, useRef, useState } from 'react';
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { Bell, ChevronDown, LogOut, Users } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { roleLabels } from '../lib/constants';
import Avatar from './Avatar';

const NAV_TABS = [
  { label: 'Обзор', path: '/' },
  { label: 'Проекты', path: '/projects' },
  { label: 'Сотрудники', path: '/admin/employees' },
] as const;

function getActiveTab(pathname: string): string {
  if (pathname.startsWith('/projects')) return '/projects';
  if (pathname.startsWith('/admin/employees')) return '/admin/employees';
  if (pathname.startsWith('/profile')) return '/admin/employees';
  return '/';
}

export default function Layout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);
  const userMenuRef = useRef<HTMLDivElement>(null);

  const userRole = user?.role || 'user';
  const userName = user?.full_name || user?.login || 'Пользователь';
  const activeTab = getActiveTab(location.pathname);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(event.target as Node)) {
        setIsUserMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="relative min-h-screen bg-white text-gray-900">
      {/* Header */}
      <header className="sticky top-0 z-20 border-b border-emerald-50 bg-white/80 backdrop-blur-xl">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-8">
            <Link to="/" className="flex items-center gap-2">
              <span className="text-lg font-semibold tracking-tight text-emerald-600">КОМАНДА</span>
              <span className="rounded-xl bg-gradient-to-r from-lime-400 to-emerald-400 px-3 py-1 text-sm font-bold text-gray-900">
                F5
              </span>
            </Link>
            <nav className="hidden gap-4 md:flex">
              {NAV_TABS.map((tab) => (
                <button
                  key={tab.path}
                  type="button"
                  onClick={() => navigate(tab.path)}
                  className={`rounded-full px-4 py-2 text-sm font-medium transition-all ${
                    activeTab === tab.path
                      ? 'bg-gray-900 text-white'
                      : 'text-gray-500 hover:text-emerald-600'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </nav>
          </div>

          <div className="flex items-center gap-3">
            <button
              type="button"
              className="rounded-full border border-gray-200 bg-white p-2 text-gray-500 hover:text-emerald-600"
            >
              <Bell className="h-5 w-5" />
            </button>

            {/* User menu */}
            <div className="relative" ref={userMenuRef}>
              <button
                type="button"
                onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
                className="flex items-center gap-2 rounded-full border border-gray-200 bg-white px-3 py-1.5 text-sm hover:border-emerald-200"
              >
                <Avatar
                  src={user?.avatar_url}
                  name={userName}
                  size="sm"
                />
                <div className="text-left hidden sm:block">
                  <p className="text-sm font-medium text-gray-700">{userName}</p>
                  <p className="text-[11px] text-gray-400">{roleLabels[userRole] || userRole}</p>
                </div>
                <ChevronDown className={`h-4 w-4 text-gray-400 transition-transform ${isUserMenuOpen ? 'rotate-180' : ''}`} />
              </button>

              {isUserMenuOpen && (
                <div className="absolute right-0 mt-2 w-48 rounded-xl border border-gray-100 bg-white py-2 shadow-lg">
                  <Link
                    to="/profile"
                    onClick={() => setIsUserMenuOpen(false)}
                    className="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                  >
                    <Users className="h-4 w-4" />
                    Мой профиль
                  </Link>
                  <button
                    type="button"
                    onClick={handleLogout}
                    className="flex w-full items-center gap-2 px-4 py-2 text-sm text-red-600 hover:bg-red-50"
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

      {/* Page content */}
      <Outlet />
    </div>
  );
}
