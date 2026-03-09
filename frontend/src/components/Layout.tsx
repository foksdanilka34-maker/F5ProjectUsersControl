import { useCallback, useEffect, useRef, useState } from 'react';
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { Bell, ChevronDown, LogOut, Users, CheckCircle, UserPlus, Eye } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { useWSEvent } from '../context/WebSocketContext';
import { roleLabels } from '../lib/constants';
import Avatar from './Avatar';

export interface AppNotification {
  id: string;
  kind: 'review_assigned' | 'review_approved' | 'task_assigned';
  message: string;
  taskId: number;
  projectId: number;
  taskTitle: string;
  timestamp: number;
  read: boolean;
}

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

function getTimeAgo(ts: number): string {
  const diff = Math.floor((Date.now() - ts) / 1000);
  if (diff < 60) return 'только что';
  if (diff < 3600) return `${Math.floor(diff / 60)} мин. назад`;
  if (diff < 86400) return `${Math.floor(diff / 3600)} ч. назад`;
  return `${Math.floor(diff / 86400)} д. назад`;
}

export default function Layout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);
  const [isNotifOpen, setIsNotifOpen] = useState(false);
  const [notifications, setNotifications] = useState<AppNotification[]>([]);
  const userMenuRef = useRef<HTMLDivElement>(null);
  const notifRef = useRef<HTMLDivElement>(null);

  const userRole = user?.role || 'user';
  const userName = user?.full_name || user?.login || 'Пользователь';
  const activeTab = getActiveTab(location.pathname);

  // Subscribe to WS notifications
  const handleNotification = useCallback(
    (payload: unknown) => {
      const data = payload as {
        target_user_id?: number;
        kind?: string;
        message?: string;
        task_id?: number;
        project_id?: number;
        task_title?: string;
      };
      // Only process notifications targeted at current user
      if (!user || data.target_user_id !== user.id) return;

      const notif: AppNotification = {
        id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
        kind: (data.kind as AppNotification['kind']) || 'task_assigned',
        message: data.message || 'Новое уведомление',
        taskId: data.task_id || 0,
        projectId: data.project_id || 0,
        taskTitle: data.task_title || '',
        timestamp: Date.now(),
        read: false,
      };

      setNotifications((prev) => [notif, ...prev].slice(0, 50));
    },
    [user],
  );

  useWSEvent('notification', handleNotification, [handleNotification]);

  const unreadCount = notifications.filter((n) => !n.read).length;

  const markAllRead = () => {
    setNotifications((prev) => prev.map((n) => ({ ...n, read: true })));
  };

  const clearAll = () => {
    setNotifications([]);
    setIsNotifOpen(false);
  };

  const handleNotifClick = (notif: AppNotification) => {
    setNotifications((prev) =>
      prev.map((n) => (n.id === notif.id ? { ...n, read: true } : n)),
    );
    setIsNotifOpen(false);
    if (notif.projectId) {
      navigate(`/projects/${notif.projectId}`);
    }
  };

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(event.target as Node)) {
        setIsUserMenuOpen(false);
      }
      if (notifRef.current && !notifRef.current.contains(event.target as Node)) {
        setIsNotifOpen(false);
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
            {/* Notification bell */}
            <div className="relative" ref={notifRef}>
              <button
                type="button"
                onClick={() => {
                  setIsNotifOpen(!isNotifOpen);
                  if (!isNotifOpen && unreadCount > 0) markAllRead();
                }}
                className="relative rounded-full border border-gray-200 bg-white p-2 text-gray-500 hover:text-emerald-600 transition-colors"
              >
                <Bell className="h-5 w-5" />
                {unreadCount > 0 && (
                  <span className="absolute -right-1 -top-1 flex h-5 min-w-[20px] items-center justify-center rounded-full bg-red-500 px-1 text-[11px] font-bold text-white">
                    {unreadCount > 9 ? '9+' : unreadCount}
                  </span>
                )}
              </button>

              {isNotifOpen && (
                <div className="absolute right-0 mt-2 w-80 rounded-xl border border-gray-100 bg-white shadow-xl overflow-hidden z-50">
                  <div className="flex items-center justify-between border-b border-gray-100 px-4 py-3">
                    <h3 className="text-sm font-semibold text-gray-800">Уведомления</h3>
                    {notifications.length > 0 && (
                      <button
                        type="button"
                        onClick={clearAll}
                        className="text-xs text-gray-400 hover:text-red-500 transition-colors"
                      >
                        Очистить
                      </button>
                    )}
                  </div>
                  <div className="max-h-80 overflow-y-auto">
                    {notifications.length === 0 ? (
                      <div className="px-4 py-8 text-center text-sm text-gray-400">
                        Нет уведомлений
                      </div>
                    ) : (
                      notifications.map((notif) => {
                        const icon =
                          notif.kind === 'review_assigned' ? (
                            <Eye className="h-4 w-4 text-indigo-500" />
                          ) : notif.kind === 'review_approved' ? (
                            <CheckCircle className="h-4 w-4 text-emerald-500" />
                          ) : (
                            <UserPlus className="h-4 w-4 text-blue-500" />
                          );

                        const timeAgo = getTimeAgo(notif.timestamp);

                        return (
                          <button
                            key={notif.id}
                            type="button"
                            onClick={() => handleNotifClick(notif)}
                            className={`flex w-full items-start gap-3 px-4 py-3 text-left transition-colors hover:bg-gray-50 ${
                              !notif.read ? 'bg-emerald-50/40' : ''
                            }`}
                          >
                            <div className="mt-0.5 flex-shrink-0 rounded-lg bg-gray-100 p-1.5">
                              {icon}
                            </div>
                            <div className="min-w-0 flex-1">
                              <p className="text-sm text-gray-700 line-clamp-2">{notif.message}</p>
                              <p className="mt-1 text-xs text-gray-400">{timeAgo}</p>
                            </div>
                            {!notif.read && (
                              <div className="mt-2 h-2 w-2 flex-shrink-0 rounded-full bg-emerald-500" />
                            )}
                          </button>
                        );
                      })
                    )}
                  </div>
                </div>
              )}
            </div>

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
