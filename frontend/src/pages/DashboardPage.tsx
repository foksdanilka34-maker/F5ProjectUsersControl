import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  ArrowRight,
  Bell,
  ChevronDown,
  FolderKanban,
  LogOut,
  Plus,
  Users,
  TrendingUp,
  CheckCircle,
  Clock,
  AlertTriangle,
  BarChart3,
  Wifi,
  WifiOff,
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { useWebSocket } from '../context/WebSocketContext';
import { roleLabels, projectStatusLabels, projectStatusColors } from '../lib/constants';
import { getProjects, type Project } from '../services/projectService';
import { getDashboardStats, type DashboardStats } from '../services/analyticsService';
import Avatar from '../components/Avatar';

const navTabs = ['Обзор', 'Проекты', 'Сотрудники'];

export default function DashboardPage() {
  const [activeTab, setActiveTab] = useState('Обзор');
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);
  const [projects, setProjects] = useState<Project[]>([]);
  const [loadingProjects, setLoadingProjects] = useState(true);
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loadingStats, setLoadingStats] = useState(true);
  const userMenuRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const { isConnected, subscribe } = useWebSocket();

  const userRole = user?.role || 'user';
  const userName = user?.full_name || user?.login || 'Пользователь';
  
  // Проверка - менеджеры, директора, админы и разработчики видят все проекты, остальные только свои
  const isManagerOrHigher = ['admin', 'director', 'manager', 'developer'].includes(userRole);

  const loadStats = useCallback(async () => {
    try {
      const data = await getDashboardStats();
      setStats(data);
    } catch (error) {
      console.error('Failed to load stats:', error);
    } finally {
      setLoadingStats(false);
    }
  }, []);

  useEffect(() => {
    const loadProjects = async () => {
      try {
        // Обычные сотрудники видят только проекты, в которых они участвуют
        const params = isManagerOrHigher ? {} : { member_id: user?.id };
        const data = await getProjects(params);
        setProjects(data);
      } catch (error) {
        console.error('Failed to load projects:', error);
      } finally {
        setLoadingProjects(false);
      }
    };
    
    loadProjects();
    loadStats();
  }, [isManagerOrHigher, user?.id, loadStats]);

  // WebSocket: обновляем статистику при изменении задач
  useEffect(() => {
    const reloadStats = () => {
      console.log('[Dashboard] Task event received, reloading stats...');
      loadStats();
    };

    const unsubCreated = subscribe('task:created', reloadStats);
    const unsubUpdated = subscribe('task:updated', reloadStats);
    const unsubDeleted = subscribe('task:deleted', reloadStats);
    const unsubMoved = subscribe('task:moved', reloadStats);
    const unsubAssigned = subscribe('task:assigned', reloadStats);

    return () => {
      unsubCreated();
      unsubUpdated();
      unsubDeleted();
      unsubMoved();
      unsubAssigned();
    };
  }, [subscribe, loadStats]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(event.target as Node)) {
        setIsUserMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const permissions = useMemo(
    () => ({
      canCreateProject: ['admin', 'manager', 'developer'].includes(userRole),
      canCreateEmployee: ['admin', 'hr'].includes(userRole),
    }),
    [userRole],
  );

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const recentProjects = projects.slice(0, 5);

  // Calculate percentages for visual bars - защита от NaN
  const safePercent = (value: number) => {
    if (!Number.isFinite(value) || Number.isNaN(value)) return 0;
    return Math.max(0, Math.min(100, Math.round(value)));
  };
  
  const taskCompletionPercent = safePercent(stats ? (stats.completed_tasks / Math.max(stats.total_tasks, 1)) * 100 : 0);
  const projectActivePercent = safePercent(stats ? (stats.active_projects / Math.max(stats.total_projects, 1)) * 100 : 0);
  // Бэкенд уже отдает проценты, не умножаем на 100
  const avgCompletionPercent = safePercent(stats?.avg_completion_rate || 0);
  const avgOnTimePercent = safePercent(stats?.avg_on_time_rate || 0);

  return (
    <div className="relative min-h-screen bg-white text-gray-900">
      {/* Background decorations */}
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -left-24 top-12 h-80 w-80 rounded-full bg-emerald-200/40 blur-3xl" />
        <div className="absolute -right-16 top-36 h-64 w-64 rounded-full bg-lime-200/40 blur-3xl" />
        <div className="absolute left-1/3 top-1/2 h-96 w-96 rounded-full bg-emerald-100/30 blur-3xl" />
      </div>

      <div className="relative z-10">
        {/* Header */}
        <header className="sticky top-0 z-20 border-b border-emerald-50 bg-white/80 backdrop-blur-xl">
          <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
            <div className="flex items-center gap-8">
              <div className="flex items-center gap-2">
                <span className="text-lg font-semibold tracking-tight text-emerald-600">КОМАНДА</span>
                <span className="rounded-xl bg-gradient-to-r from-lime-400 to-emerald-400 px-3 py-1 text-sm font-bold text-gray-900">
                  F5
                </span>
              </div>
              <nav className="hidden gap-4 md:flex">
                {navTabs.map((tab) => (
                  <button
                    key={tab}
                    type="button"
                    onClick={() => {
                      setActiveTab(tab);
                      if (tab === 'Сотрудники') {
                        navigate('/admin/employees');
                      } else if (tab === 'Проекты') {
                        navigate('/projects');
                      } else if (tab === 'Обзор') {
                        navigate('/');
                      }
                    }}
                    className={`rounded-full px-4 py-2 text-sm font-medium transition-all ${
                      activeTab === tab
                        ? 'bg-gray-900 text-white'
                        : 'text-gray-500 hover:text-emerald-600'
                    }`}
                  >
                    {tab}
                  </button>
                ))}
              </nav>
            </div>

            <div className="flex items-center gap-3">
              {/* WebSocket Status Indicator */}
              <span
                className={`inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-medium transition-colors ${
                  isConnected
                    ? 'bg-emerald-50 text-emerald-600 border border-emerald-200'
                    : 'bg-gray-100 text-gray-500 border border-gray-200'
                }`}
                title={isConnected ? 'Данные обновляются в реальном времени' : 'Нет подключения к серверу'}
              >
                {isConnected ? (
                  <Wifi className="h-3.5 w-3.5" />
                ) : (
                  <WifiOff className="h-3.5 w-3.5" />
                )}
                {isConnected ? 'Live' : 'Offline'}
              </span>

              <button
                type="button"
                className="rounded-full border border-gray-200 bg-white p-2 text-gray-500 hover:text-emerald-600"
              >
                <Bell className="h-5 w-5" />
              </button>
              
              {/* User Menu */}
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
                  <div className="text-left">
                    <p className="text-sm font-medium text-gray-700">{userName}</p>
                    <p className="text-[11px] text-gray-400">{roleLabels[userRole] || userRole}</p>
                  </div>
                  <ChevronDown className={`h-4 w-4 text-gray-400 transition-transform ${isUserMenuOpen ? 'rotate-180' : ''}`} />
                </button>
                
                {isUserMenuOpen && (
                  <div className="absolute right-0 mt-2 w-48 rounded-xl border border-gray-100 bg-white py-2 shadow-lg">
                    <Link
                      to="/profile"
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

        {/* Main Content */}
        <main className="mx-auto max-w-7xl px-6 py-8">
          {/* Welcome + Quick Stats */}
          <div className="grid gap-6 lg:grid-cols-3">
            <div className="lg:col-span-2 rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
              <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-widest text-emerald-500">Главная</p>
                  <h1 className="mt-2 text-2xl font-bold text-gray-900">Привет, {userName}!</h1>
                  <p className="mt-1 text-sm text-gray-500">Управляйте проектами и командой</p>
                </div>
                <div className="flex flex-wrap gap-2">
                  {permissions.canCreateProject && (
                    <Link
                      to="/projects/new"
                      className="inline-flex items-center gap-2 rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600"
                    >
                      <Plus className="h-4 w-4" />
                      Новый проект
                    </Link>
                  )}
                  {permissions.canCreateEmployee && (
                    <Link
                      to="/admin/employees"
                      className="inline-flex items-center gap-2 rounded-xl border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                    >
                      <Users className="h-4 w-4" />
                      Сотрудники
                    </Link>
                  )}
                </div>
              </div>
            </div>
            
            {/* Mini Stats */}
            <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm">
              <p className="text-xs font-semibold uppercase tracking-wide text-gray-400 mb-3">Сводка</p>
              {loadingStats ? (
                <div className="flex items-center justify-center py-8">
                  <div className="h-6 w-6 animate-spin rounded-full border-2 border-emerald-500 border-t-transparent" />
                </div>
              ) : stats ? (
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-500">Проектов</span>
                    <span className="font-bold text-gray-900">{stats.total_projects}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-500">Сотрудников</span>
                    <span className="font-bold text-gray-900">{stats.total_employees}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-500">Задач</span>
                    <span className="font-bold text-gray-900">{stats.total_tasks}</span>
                  </div>
                </div>
              ) : (
                <p className="text-sm text-gray-400 text-center py-4">Нет данных</p>
              )}
            </div>
          </div>

          {/* Analytics Charts */}
          <div className="mt-6 grid gap-6 lg:grid-cols-2 xl:grid-cols-4">
            {/* Tasks Progress */}
            <div className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
              <div className="flex items-center gap-3">
                <div className="rounded-xl bg-emerald-50 p-2.5">
                  <CheckCircle className="h-5 w-5 text-emerald-600" />
                </div>
                <div>
                  <p className="text-xs text-gray-500">Выполнено задач</p>
                  <p className="text-xl font-bold text-gray-900">
                    {loadingStats ? '—' : `${stats?.completed_tasks || 0} / ${stats?.total_tasks || 0}`}
                  </p>
                </div>
              </div>
              <div className="mt-4">
                <div className="flex justify-between text-xs mb-1">
                  <span className="text-gray-400">Прогресс</span>
                  <span className="font-medium text-emerald-600">{taskCompletionPercent}%</span>
                </div>
                <div className="h-2 rounded-full bg-gray-100">
                  <div
                    className="h-2 rounded-full bg-gradient-to-r from-emerald-400 to-lime-400 transition-all duration-500"
                    style={{ width: `${taskCompletionPercent}%` }}
                  />
                </div>
              </div>
            </div>

            {/* Active Projects */}
            <div className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
              <div className="flex items-center gap-3">
                <div className="rounded-xl bg-blue-50 p-2.5">
                  <FolderKanban className="h-5 w-5 text-blue-600" />
                </div>
                <div>
                  <p className="text-xs text-gray-500">Активные проекты</p>
                  <p className="text-xl font-bold text-gray-900">
                    {loadingStats ? '—' : `${stats?.active_projects || 0} / ${stats?.total_projects || 0}`}
                  </p>
                </div>
              </div>
              <div className="mt-4">
                <div className="flex justify-between text-xs mb-1">
                  <span className="text-gray-400">Активность</span>
                  <span className="font-medium text-blue-600">{projectActivePercent}%</span>
                </div>
                <div className="h-2 rounded-full bg-gray-100">
                  <div
                    className="h-2 rounded-full bg-gradient-to-r from-blue-400 to-cyan-400 transition-all duration-500"
                    style={{ width: `${projectActivePercent}%` }}
                  />
                </div>
              </div>
            </div>

            {/* Overdue Tasks */}
            <div className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
              <div className="flex items-center gap-3">
                <div className="rounded-xl bg-amber-50 p-2.5">
                  <Clock className="h-5 w-5 text-amber-600" />
                </div>
                <div>
                  <p className="text-xs text-gray-500">Просроченных</p>
                  <p className="text-xl font-bold text-gray-900">
                    {loadingStats ? '—' : stats?.overdue_tasks || 0}
                  </p>
                </div>
              </div>
              <div className="mt-4 space-y-2">
                <div className="flex items-center justify-between text-xs">
                  <span className="text-gray-500">Завершено с опозданием</span>
                  <span className="font-medium text-amber-600">{stats?.completed_late || 0}</span>
                </div>
                {stats && stats.overdue_tasks > 0 ? (
                  <div className="flex items-center gap-2 text-xs text-amber-600">
                    <AlertTriangle className="h-3.5 w-3.5" />
                    <span>Требуют внимания</span>
                  </div>
                ) : (
                  <div className="flex items-center gap-2 text-xs text-emerald-600">
                    <CheckCircle className="h-3.5 w-3.5" />
                    <span>Всё в срок</span>
                  </div>
                )}
              </div>
            </div>

            {/* Team Performance */}
            <div className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
              <div className="flex items-center gap-3">
                <div className="rounded-xl bg-violet-50 p-2.5">
                  <TrendingUp className="h-5 w-5 text-violet-600" />
                </div>
                <div>
                  <p className="text-xs text-gray-500">Эффективность</p>
                  <p className="text-xl font-bold text-gray-900">
                    {loadingStats ? '—' : `${avgCompletionPercent}%`}
                  </p>
                </div>
              </div>
              <div className="mt-4">
                <div className="flex justify-between text-xs mb-1">
                  <span className="text-gray-400">В срок</span>
                  <span className="font-medium text-violet-600">
                    {loadingStats ? '—' : `${avgOnTimePercent}%`}
                  </span>
                </div>
                <div className="h-2 rounded-full bg-gray-100">
                  <div
                    className="h-2 rounded-full bg-gradient-to-r from-violet-400 to-purple-400 transition-all duration-500"
                    style={{ width: `${avgOnTimePercent}%` }}
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Two Column Layout */}
          <div className="mt-6 grid gap-6 lg:grid-cols-3">
            {/* Recent Projects */}
            <div className="lg:col-span-2 rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <BarChart3 className="h-5 w-5 text-emerald-500" />
                  <h2 className="font-semibold text-gray-900">Последние проекты</h2>
                </div>
                <Link to="/projects" className="text-sm font-medium text-emerald-600 hover:text-emerald-700">
                  Все →
                </Link>
              </div>

              {loadingProjects ? (
                <div className="flex items-center justify-center py-8">
                  <div className="h-6 w-6 animate-spin rounded-full border-2 border-emerald-500 border-t-transparent" />
                </div>
              ) : recentProjects.length > 0 ? (
                <div className="space-y-2">
                  {recentProjects.map((project) => (
                    <Link
                      to={`/projects/${project.id}`}
                      key={project.id}
                      className="flex items-center justify-between rounded-xl border border-gray-100 p-3 hover:border-emerald-100 hover:bg-emerald-50/30 transition-colors"
                    >
                      <div className="flex items-center gap-3">
                        <div className="h-9 w-9 rounded-lg bg-gradient-to-r from-emerald-400 to-lime-400 flex items-center justify-center text-white font-semibold text-sm">
                          {project.name.charAt(0).toUpperCase()}
                        </div>
                        <div>
                          <h3 className="text-sm font-medium text-gray-900">{project.name}</h3>
                          <p className="text-xs text-gray-500 line-clamp-1">{project.description || 'Без описания'}</p>
                        </div>
                      </div>
                      <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${projectStatusColors[project.status] || 'bg-gray-100 text-gray-600'}`}>
                        {projectStatusLabels[project.status] || project.status}
                      </span>
                    </Link>
                  ))}
                </div>
              ) : (
                <div className="rounded-xl border border-dashed border-gray-200 p-8 text-center">
                  <FolderKanban className="mx-auto h-10 w-10 text-gray-300" />
                  <p className="mt-3 text-sm text-gray-500">Проектов пока нет</p>
                  {permissions.canCreateProject && (
                    <Link
                      to="/projects/new"
                      className="mt-3 inline-flex items-center gap-2 text-sm font-medium text-emerald-600 hover:text-emerald-700"
                    >
                      <Plus className="h-4 w-4" />
                      Создать проект
                    </Link>
                  )}
                </div>
              )}
            </div>

            {/* Quick Actions */}
            <div className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
              <h2 className="font-semibold text-gray-900 mb-4">Быстрые действия</h2>
              <div className="space-y-2">
                <Link
                  to="/projects"
                  className="flex items-center gap-3 rounded-xl border border-gray-100 p-3 hover:border-emerald-100 hover:bg-emerald-50/50 transition-colors group"
                >
                  <div className="rounded-lg bg-emerald-50 p-2 text-emerald-600 group-hover:bg-emerald-100">
                    <FolderKanban className="h-4 w-4" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-900">Проекты</p>
                    <p className="text-xs text-gray-500">{projects.length} всего</p>
                  </div>
                  <ArrowRight className="ml-auto h-4 w-4 text-gray-300 group-hover:text-emerald-500" />
                </Link>

                <Link
                  to="/admin/employees"
                  className="flex items-center gap-3 rounded-xl border border-gray-100 p-3 hover:border-emerald-100 hover:bg-emerald-50/50 transition-colors group"
                >
                  <div className="rounded-lg bg-blue-50 p-2 text-blue-600 group-hover:bg-blue-100">
                    <Users className="h-4 w-4" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-900">Сотрудники</p>
                    <p className="text-xs text-gray-500">{stats?.total_employees || 0} человек</p>
                  </div>
                  <ArrowRight className="ml-auto h-4 w-4 text-gray-300 group-hover:text-emerald-500" />
                </Link>

                {permissions.canCreateProject && (
                  <Link
                    to="/projects/new"
                    className="flex items-center gap-3 rounded-xl border border-dashed border-emerald-200 bg-emerald-50/50 p-3 hover:bg-emerald-50 transition-colors group"
                  >
                    <div className="rounded-lg bg-emerald-100 p-2 text-emerald-600">
                      <Plus className="h-4 w-4" />
                    </div>
                    <div>
                      <p className="text-sm font-medium text-emerald-700">Новый проект</p>
                      <p className="text-xs text-emerald-600">Создать</p>
                    </div>
                  </Link>
                )}
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
