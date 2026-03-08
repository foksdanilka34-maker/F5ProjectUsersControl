import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  ArrowRight,
  FolderKanban,
  Plus,
  Users,
  TrendingUp,
  CheckCircle,
  Clock,
  AlertTriangle,
  BarChart3,
  Trophy,
  Medal,
  AlertOctagon,
} from 'lucide-react';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts';
import Avatar from '../components/Avatar';
import { useAuth } from '../context/AuthContext';
import { useWebSocket } from '../context/WebSocketContext';
import { projectStatusLabels, projectStatusColors } from '../lib/constants';
import { getProjects, type Project } from '../services/projectService';
import { getDashboardStats, type DashboardStats } from '../services/analyticsService';
import { listProfiles } from '../services/employeeService';
import type { ProfileDTO } from '../services/types';

export default function DashboardPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loadingProjects, setLoadingProjects] = useState(true);
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loadingStats, setLoadingStats] = useState(true);
  const [profilesMap, setProfilesMap] = useState<Map<number, ProfileDTO>>(new Map());
  const { user } = useAuth();
  const { subscribe } = useWebSocket();

  const userRole = user?.role || 'user';
  const userName = user?.full_name || user?.login || 'Пользователь';

  const isManagerOrHigher = ['admin', 'director', 'manager', 'developer'].includes(userRole);

  const loadStats = useCallback(async () => {
    try {
      console.log('[Dashboard] Loading stats...');
      const data = await getDashboardStats();
      console.log('[Dashboard] Stats loaded:', data);
      setStats(data);
    } catch (error) {
      console.error('Failed to load stats:', error);
    } finally {
      setLoadingStats(false);
    }
  }, []);

  const loadStatsRef = useRef(loadStats);
  loadStatsRef.current = loadStats;

  useEffect(() => {
    const loadProjects = async () => {
      try {

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

    listProfiles({ pageSize: 200 })
      .then((res) => {
        const map = new Map<number, ProfileDTO>();
        res.profiles.forEach((p) => map.set(p.id, p));
        setProfilesMap(map);
      })
      .catch((err) => console.error('Failed to load profiles:', err));
  }, [isManagerOrHigher, user?.id, loadStats]);

  useEffect(() => {

    const reloadStats = () => {
      console.log('[Dashboard] Task event received, reloading stats...');
      loadStatsRef.current();
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
  }, [subscribe]); // Убираем loadStats из зависимостей, используем ref

  const permissions = useMemo(
    () => ({
      canCreateProject: ['admin', 'manager', 'developer'].includes(userRole),
      canCreateEmployee: ['admin', 'hr'].includes(userRole),
    }),
    [userRole],
  );

  const recentProjects = projects.slice(0, 5);

  const safePercent = (value: number) => {
    if (!Number.isFinite(value) || Number.isNaN(value)) return 0;
    return Math.max(0, Math.min(100, Math.round(value)));
  };
  
  const taskCompletionPercent = safePercent(stats ? (stats.completed_tasks / Math.max(stats.total_tasks, 1)) * 100 : 0);
  const projectActivePercent = safePercent(stats ? (stats.active_projects / Math.max(stats.total_projects, 1)) * 100 : 0);

  const avgCompletionPercent = safePercent(stats?.avg_completion_rate || 0);
  const avgOnTimePercent = safePercent(stats?.avg_on_time_rate || 0);

  return (
    <div className="relative">
      {}
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -left-24 top-12 h-80 w-80 rounded-full bg-emerald-200/40 blur-3xl" />
        <div className="absolute -right-16 top-36 h-64 w-64 rounded-full bg-lime-200/40 blur-3xl" />
        <div className="absolute left-1/3 top-1/2 h-96 w-96 rounded-full bg-emerald-100/30 blur-3xl" />
      </div>

      <div className="relative z-10">
        {}
        <main className="mx-auto max-w-7xl px-6 py-8">
          {}
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
            
            {}
            <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm flex flex-col">
              <p className="text-xs font-semibold uppercase tracking-wide text-gray-400 mb-2">Задачи</p>
              {loadingStats ? (
                <div className="flex items-center justify-center py-8">
                  <div className="h-6 w-6 animate-spin rounded-full border-2 border-emerald-500 border-t-transparent" />
                </div>
              ) : stats && stats.total_tasks > 0 ? (
                (() => {
                  const inProgress = stats.total_tasks - stats.completed_tasks - stats.overdue_tasks;
                  const donutData = [
                    { name: 'Вовремя', value: stats.completed_on_time || 0, color: '#10b981' },
                    { name: 'С опозданием', value: stats.completed_late || 0, color: '#f59e0b' },
                    { name: 'В работе', value: Math.max(0, inProgress), color: '#3b82f6' },
                    { name: 'Просрочено', value: stats.overdue_tasks || 0, color: '#ef4444' },
                  ].filter((d) => d.value > 0);
                  return (
                    <div className="flex-1 flex flex-col items-center justify-center">
                      <div className="relative w-full" style={{ height: 150 }}>
                        <ResponsiveContainer width="100%" height="100%">
                          <PieChart>
                            <Pie
                              data={donutData}
                              cx="50%"
                              cy="50%"
                              innerRadius={40}
                              outerRadius={60}
                              paddingAngle={3}
                              dataKey="value"
                              stroke="none"
                            >
                              {donutData.map((entry, i) => (
                                <Cell key={i} fill={entry.color} />
                              ))}
                            </Pie>
                            <Tooltip
                              formatter={(value: number, name: string) => [`${value} задач`, name]}
                              contentStyle={{ borderRadius: '12px', border: '1px solid #e5e7eb', fontSize: '12px' }}
                            />
                          </PieChart>
                        </ResponsiveContainer>
                        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                          <div className="text-center">
                            <p className="text-lg font-bold text-gray-900">{stats.total_tasks}</p>
                            <p className="text-[10px] text-gray-400">задач</p>
                          </div>
                        </div>
                      </div>
                      <div className="grid grid-cols-2 gap-x-4 gap-y-1 mt-2 w-full px-2">
                        {donutData.map((d) => (
                          <div key={d.name} className="flex items-center gap-1.5 text-xs text-gray-600">
                            <span className="h-2 w-2 rounded-full flex-shrink-0" style={{ backgroundColor: d.color }} />
                            <span className="truncate">{d.name}</span>
                            <span className="ml-auto font-medium">{d.value}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  );
                })()
              ) : (
                <p className="text-sm text-gray-400 text-center py-4">Нет данных</p>
              )}
            </div>
          </div>

          {}
          <div className="mt-6 grid gap-6 lg:grid-cols-2 xl:grid-cols-4">
            {}
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

            {}
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

            {}
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

            {}
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

          {}
          <div className="mt-6 grid gap-6 lg:grid-cols-3">
            {}
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
                        <div className="h-9 w-9 rounded-lg bg-emerald-50 flex items-center justify-center">
                          <FolderKanban className="h-4.5 w-4.5 text-emerald-600" />
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

            {}
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

          {/* Top employees leaderboard */}
          {stats && stats.top_employees && stats.top_employees.length > 0 && (
            <div className="mt-6 grid gap-6 lg:grid-cols-2">
              <div className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
                <div className="flex items-center gap-2 mb-4">
                  <Trophy className="h-5 w-5 text-amber-500" />
                  <h2 className="font-semibold text-gray-900">Лучшие сотрудники</h2>
                </div>
                <div className="space-y-3">
                  {stats.top_employees.slice(0, 5).map((emp, index) => {
                    const profile = profilesMap.get(emp.employee_id);
                    const name = profile
                      ? `${profile.first_name} ${profile.last_name}`
                      : `Сотрудник #${emp.employee_id}`;
                    const medalColors = ['text-amber-500', 'text-gray-400', 'text-orange-400'];
                    return (
                      <Link
                        key={emp.employee_id}
                        to={`/profile/${emp.employee_id}`}
                        className="flex items-center gap-3 rounded-xl border border-gray-50 p-3 hover:border-emerald-100 hover:bg-emerald-50/30 transition-colors"
                      >
                        <div className="flex-shrink-0 w-6 text-center">
                          {index < 3 ? (
                            <Medal className={`h-5 w-5 ${medalColors[index]}`} />
                          ) : (
                            <span className="text-sm font-bold text-gray-300">{index + 1}</span>
                          )}
                        </div>
                        <Avatar
                          src={profile?.avatar_url}
                          name={name}
                          size="sm"
                        />
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium text-gray-900 truncate">{name}</p>
                          <p className="text-xs text-gray-500">{emp.tasks_completed} задач выполнено</p>
                        </div>
                        <div className="flex-shrink-0">
                          <div className="relative h-10 w-10">
                            <svg className="h-10 w-10 -rotate-90" viewBox="0 0 36 36">
                              <circle cx="18" cy="18" r="14" fill="none" stroke="#e5e7eb" strokeWidth="3" />
                              <circle
                                cx="18" cy="18" r="14" fill="none"
                                stroke="#10b981"
                                strokeWidth="3"
                                strokeDasharray={`${Math.round(emp.completion_rate * 0.88)} 88`}
                                strokeLinecap="round"
                              />
                            </svg>
                            <span className="absolute inset-0 flex items-center justify-center text-[10px] font-bold text-gray-700">
                              {Math.round(emp.completion_rate)}%
                            </span>
                          </div>
                        </div>
                      </Link>
                    );
                  })}
                </div>
              </div>

              {/* Problematic projects warning */}
              {stats.problematic_projects && stats.problematic_projects.length > 0 && (
                <div className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
                  <div className="flex items-center gap-2 mb-4">
                    <AlertOctagon className="h-5 w-5 text-red-500" />
                    <h2 className="font-semibold text-gray-900">Проблемные проекты</h2>
                  </div>
                  <div className="space-y-3">
                    {stats.problematic_projects.slice(0, 5).map((pp) => {
                      const project = projects.find((p) => p.id === pp.project_id);
                      const name = project?.name || `Проект #${pp.project_id}`;
                      const isHealthy = pp.health_status === 'HEALTH_STATUS_HEALTHY';
                      const isAtRisk = pp.health_status === 'HEALTH_STATUS_AT_RISK';
                      const healthColor = isHealthy
                        ? 'bg-emerald-500'
                        : isAtRisk
                        ? 'bg-amber-500'
                        : 'bg-red-500';
                      const healthLabel = isHealthy
                        ? 'Нормально'
                        : isAtRisk
                        ? 'Риск'
                        : 'Критично';
                      const onTime = Math.round(pp.on_time_rate || 0);
                      return (
                        <Link
                          key={pp.project_id}
                          to={`/projects/${pp.project_id}`}
                          className="block rounded-xl border border-gray-50 p-3 hover:border-red-100 hover:bg-red-50/20 transition-colors"
                        >
                          <div className="flex items-center justify-between mb-2">
                            <div className="flex items-center gap-2">
                              <FolderKanban className="h-4 w-4 text-gray-400" />
                              <p className="text-sm font-medium text-gray-900 truncate">{name}</p>
                            </div>
                            <div className="flex items-center gap-1.5">
                              <span className={`h-2 w-2 rounded-full ${healthColor}`} />
                              <span className="text-xs text-gray-500">{healthLabel}</span>
                            </div>
                          </div>
                          <div className="flex items-center gap-2">
                            <div className="flex-1 h-1.5 rounded-full bg-gray-100">
                              <div
                                className={`h-1.5 rounded-full transition-all ${
                                  onTime >= 70 ? 'bg-emerald-400' : onTime >= 40 ? 'bg-amber-400' : 'bg-red-400'
                                }`}
                                style={{ width: `${onTime}%` }}
                              />
                            </div>
                            <span className="text-xs font-medium text-gray-500">{onTime}% в срок</span>
                          </div>
                        </Link>
                      );
                    })}
                  </div>
                </div>
              )}
            </div>
          )}
        </main>
      </div>
    </div>
  );
}


