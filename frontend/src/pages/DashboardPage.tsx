import { useMemo, useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import type { LucideIcon } from 'lucide-react';
import {
  AlertCircle,
  ArrowDownRight,
  ArrowUpRight,
  Briefcase,
  CheckCircle,
  Plus,
  Users,
} from 'lucide-react';
import {
  Area,
  AreaChart,
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { useAuth } from '../contexts/AuthContext';
import { analyticsService } from '../api/services/analytics.service';
import { projectService } from '../api/services/project.service';
import type { 
  DashboardStats, 
  ProductivityTrend, 
  CompletionRateTrend, 
  TopPerformer, 
  ProjectMetrics,
  Project
} from '../api/types';

const periodOptions = [
  { label: 'Неделя', value: 'weekly' },
  { label: 'Месяц', value: 'monthly' },
  { label: 'Квартал', value: 'quarterly' },
];

type TrendMode = 'productivity' | 'completion';

type MetricCardProps = {
  label: string;
  value: number;
  delta?: string;
  trend?: 'up' | 'down';
  icon: LucideIcon;
};

function MetricCard({ label, value, delta, trend, icon: Icon }: MetricCardProps) {
  return (
    <div className="rounded-2xl border border-gray-100 bg-white/80 p-6 shadow-[0_10px_40px_rgba(15,118,110,0.06)]">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-500">{label}</p>
          <p className="mt-2 text-3xl font-semibold text-gray-900">{value.toLocaleString('ru-RU')}</p>
        </div>
        <div className="rounded-2xl bg-emerald-50 p-3 text-emerald-500">
          <Icon className="h-5 w-5" />
        </div>
      </div>
      {delta && (
        <div className="mt-4 flex items-center gap-2 text-sm">
          {trend === 'up' ? (
            <ArrowUpRight className="h-4 w-4 text-emerald-500" />
          ) : (
            <ArrowDownRight className="h-4 w-4 text-rose-500" />
          )}
          <span className={trend === 'up' ? 'text-emerald-600' : 'text-rose-600'}>{delta}</span>
        </div>
      )}
    </div>
  );
}

function QuickFilterChip({
  label,
  isActive,
  onClick,
}: {
  label: string;
  isActive: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`rounded-full border px-4 py-2 text-sm font-medium transition-colors ${
        isActive
          ? 'border-emerald-400 bg-linear-to-r from-lime-200/60 to-emerald-200/60 text-emerald-800 shadow-[0_8px_20px_rgba(34,197,94,0.25)]'
          : 'border-gray-200 bg-white text-gray-600 hover:border-emerald-200 hover:text-emerald-700'
      }`}
    >
      {label}
    </button>
  );
}

export default function DashboardPage() {
  const { user } = useAuth();
  
  const [period, setPeriod] = useState(periodOptions[1].value);
  const [trendMode, setTrendMode] = useState<TrendMode>('productivity');

  // Data state
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [productivityTrends, setProductivityTrends] = useState<ProductivityTrend[]>([]);
  const [completionTrends, setCompletionTrends] = useState<CompletionRateTrend[]>([]);
  const [topPerformers, setTopPerformers] = useState<TopPerformer[]>([]);
  const [riskyProjects, setRiskyProjects] = useState<ProjectMetrics[]>([]);
  const [activeProjects, setActiveProjects] = useState<Project[]>([]);

  const permissions = useMemo(
    () => ({
      canCreateProject: user?.role === 'manager' || user?.role === 'director' || user?.role === 'admin',
      canCreateEmployee: user?.role === 'admin',
    }),
    [user],
  );

  useEffect(() => {
    const fetchData = async () => {
      try {
        // Fetch all data in parallel
        const [
          statsData,
          prodTrendsData,
          compTrendsData,
          performersData,
          projectMetricsData,
          projectsData
        ] = await Promise.all([
          analyticsService.getDashboardStats(),
          analyticsService.getProductivityTrends({ period: 'DAILY', limit: 7 }),
          analyticsService.getCompletionRateTrends({ period: 'WEEKLY', limit: 5 }),
          analyticsService.getTopPerformers({ limit: 3 }),
          analyticsService.listProjectMetrics(),
          projectService.listProjects({ status: 1, page_size: 5 }) // 1 = ACTIVE
        ]);

        setStats(statsData);
        setProductivityTrends(prodTrendsData);
        setCompletionTrends(compTrendsData);
        setTopPerformers(performersData);
        
        // Filter risky projects (health < 70) and sort by health ascending
        const risky = projectMetricsData.metrics
          .filter(p => p.health_score < 70)
          .sort((a, b) => a.health_score - b.health_score)
          .slice(0, 3);
        setRiskyProjects(risky);

        setActiveProjects(projectsData.projects);

      } catch (error) {
        console.error('Failed to fetch dashboard data:', error);
      }
    };

    fetchData();
  }, [period]); // Refetch when period changes (logic to be implemented for period filtering)

  return (
    <div className="flex flex-col gap-8 lg:flex-row">
      <section className="flex-1 space-y-8">
        <div className="rounded-3xl border border-gray-100 bg-white/90 p-6 shadow-[0_20px_60px_rgba(6,95,70,0.07)]">
          <div className="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-emerald-500">Дашборд</p>
              <h1 className="mt-3 text-3xl font-semibold text-gray-900">Добро пожаловать!</h1>
              <p className="mt-2 max-w-2xl text-base text-gray-600">
                Ниже всё, что важно вашей команде сегодня: прогресс проектов, задачи и сигналы риска в одном месте.
              </p>
            </div>
            <div className="flex flex-wrap gap-3">
              {permissions.canCreateProject && (
                <Link
                  to="/projects/new"
                  className="inline-flex items-center gap-2 rounded-2xl bg-linear-to-r from-lime-400 to-emerald-400 px-5 py-2 text-sm font-semibold text-gray-900 shadow-[0_20px_40px_rgba(132,204,22,0.35)] hover:shadow-lg transition-all"
                >
                  <Plus className="h-4 w-4" />
                  Создать проект
                </Link>
              )}
              {permissions.canCreateEmployee && (
                <Link
                  to="/admin/employees/new"
                  className="inline-flex items-center gap-2 rounded-2xl border border-emerald-200 bg-white/80 px-5 py-2 text-sm font-semibold text-emerald-700 shadow-[0_12px_30px_rgba(16,185,129,0.18)] hover:bg-emerald-50 transition-all"
                >
                  <Plus className="h-4 w-4" />
                  Добавить сотрудника
                </Link>
              )}
            </div>
          </div>

          <div className="mt-6 flex flex-wrap gap-3">
            {periodOptions.map((option) => (
              <QuickFilterChip
                key={option.value}
                label={option.label}
                isActive={period === option.value}
                onClick={() => setPeriod(option.value)}
              />
            ))}
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <MetricCard 
            label="Активные сотрудники" 
            value={stats?.active_employees || 0} 
            icon={Users} 
          />
          <MetricCard 
            label="Активные проекты" 
            value={stats?.active_projects || 0} 
            icon={Briefcase} 
          />
          <MetricCard 
            label="Выполненные задачи" 
            value={stats?.completed_tasks || 0} 
            icon={CheckCircle} 
          />
          <MetricCard 
            label="Просроченные задачи" 
            value={stats?.overdue_tasks || 0} 
            trend="down"
            icon={AlertCircle} 
          />
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          {/* Active Projects List (Replacing Kanban) */}
          <div className="rounded-3xl border border-gray-100 bg-white p-6 shadow-[0_12px_40px_rgba(15,118,110,0.08)]">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-semibold uppercase tracking-widest text-emerald-500">Проекты</p>
                <h2 className="mt-2 text-xl font-semibold">Активные проекты</h2>
              </div>
              <Link to="/projects" className="text-sm font-medium text-emerald-600 hover:text-emerald-700">
                Все проекты
              </Link>
            </div>
            <div className="mt-6 space-y-4">
              {activeProjects.length > 0 ? (
                activeProjects.map((project) => (
                  <div key={project.id} className="flex items-center justify-between rounded-2xl border border-gray-50 p-4 hover:bg-gray-50 transition-colors">
                    <div>
                      <h3 className="text-sm font-semibold text-gray-900">{project.name}</h3>
                      <p className="text-xs text-gray-500 mt-1">Менеджер: {project.manager_name || 'Не назначен'}</p>
                    </div>
                    <span className="rounded-full bg-emerald-50 px-3 py-1 text-xs font-medium text-emerald-700">
                      Активен
                    </span>
                  </div>
                ))
              ) : (
                <p className="text-sm text-gray-500 text-center py-4">Нет активных проектов</p>
              )}
            </div>
          </div>

          {/* Charts */}
          <div className="rounded-3xl border border-gray-100 bg-white p-6 shadow-[0_12px_40px_rgba(15,118,110,0.08)]">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-semibold uppercase tracking-widest text-emerald-500">Динамика</p>
                <h2 className="mt-2 text-xl font-semibold">Показатели</h2>
              </div>
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={() => setTrendMode('productivity')}
                  className={`rounded-full px-4 py-1 text-sm font-medium transition-colors ${
                    trendMode === 'productivity'
                      ? 'bg-gray-900 text-white'
                      : 'text-gray-500 hover:text-emerald-600'
                  }`}
                >
                  Эффективность
                </button>
                <button
                  type="button"
                  onClick={() => setTrendMode('completion')}
                  className={`rounded-full px-4 py-1 text-sm font-medium transition-colors ${
                    trendMode === 'completion'
                      ? 'bg-gray-900 text-white'
                      : 'text-gray-500 hover:text-emerald-600'
                  }`}
                >
                  On-time
                </button>
              </div>
            </div>
            <div className="mt-6 h-72">
              {trendMode === 'productivity' ? (
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={productivityTrends}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#d1d5db" />
                    <XAxis dataKey="date" stroke="#9ca3af" tickFormatter={(val) => new Date(val).toLocaleDateString('ru-RU', { weekday: 'short' })} />
                    <YAxis stroke="#9ca3af" domain={[0, 100]} />
                    <Tooltip contentStyle={{ borderRadius: 16, borderColor: '#d1fae5' }} />
                    <Line type="monotone" dataKey="productivity_score" name="Эффективность" stroke="#10b981" strokeWidth={3} dot={{ r: 4 }} />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={completionTrends}>
                    <defs>
                      <linearGradient id="colorOnTime" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#10b981" stopOpacity={0.4} />
                        <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="4 4" stroke="#d1d5db" />
                    <XAxis dataKey="date" stroke="#9ca3af" tickFormatter={(val) => new Date(val).toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' })} />
                    <YAxis stroke="#9ca3af" domain={[0, 100]} />
                    <Tooltip contentStyle={{ borderRadius: 16, borderColor: '#d1fae5' }} />
                    <Area type="monotone" dataKey="on_time_rate" name="В срок" stroke="#10b981" fillOpacity={1} fill="url(#colorOnTime)" />
                  </AreaChart>
                </ResponsiveContainer>
              )}
            </div>
          </div>
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          {/* Leaders */}
          <div className="rounded-3xl border border-gray-100 bg-white p-6 shadow-[0_12px_40px_rgba(15,118,110,0.08)]">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-semibold uppercase tracking-widest text-emerald-500">Фокус</p>
                <h2 className="mt-2 text-xl font-semibold">Лидеры эффективности</h2>
              </div>
            </div>
            <div className="mt-6 space-y-4">
              {topPerformers.length > 0 ? (
                topPerformers.map((leader, index) => (
                  <div key={leader.employee_id} className="flex items-center gap-4 rounded-2xl border border-gray-50 p-4">
                    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-100 text-emerald-700 font-bold">
                      {index + 1}
                    </div>
                    <div className="flex-1">
                      <p className="text-sm font-semibold text-gray-900">{leader.employee_name}</p>
                      <p className="text-xs text-gray-500">{leader.department_name || 'Без отдела'}</p>
                    </div>
                    <div className="text-right">
                      <p className="text-sm font-bold text-emerald-600">{leader.productivity_score.toFixed(0)} баллов</p>
                      <p className="text-xs text-gray-500">{leader.completed_tasks} задач</p>
                    </div>
                  </div>
                ))
              ) : (
                <p className="text-sm text-gray-500 text-center py-4">Нет данных о лидерах</p>
              )}
            </div>
          </div>

          {/* Risky Projects */}
          <div className="rounded-3xl border border-gray-100 bg-white p-6 shadow-[0_12px_40px_rgba(15,118,110,0.08)]">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-semibold uppercase tracking-widest text-rose-500">Внимание</p>
                <h2 className="mt-2 text-xl font-semibold">Проекты в зоне риска</h2>
              </div>
            </div>
            <div className="mt-6 space-y-4">
              {riskyProjects.length > 0 ? (
                riskyProjects.map((project) => (
                  <div key={project.project_id} className="rounded-2xl border border-rose-100 bg-rose-50/30 p-4">
                    <div className="flex items-center justify-between">
                      <p className="text-sm font-semibold text-gray-900">{project.project_name}</p>
                      <span className="rounded-full bg-rose-100 px-2 py-1 text-xs font-medium text-rose-700">
                        Health {project.health_score}%
                      </span>
                    </div>
                    <div className="mt-2 flex items-center justify-between text-sm">
                      <span className="text-gray-600">Менеджер: {project.manager_name}</span>
                      <span className="text-rose-600 font-medium">
                        {project.overdue_tasks} просрочено
                      </span>
                    </div>
                  </div>
                ))
              ) : (
                <div className="flex flex-col items-center justify-center py-8 text-center">
                  <CheckCircle className="h-12 w-12 text-emerald-200 mb-3" />
                  <p className="text-sm font-medium text-gray-900">Всё отлично!</p>
                  <p className="text-xs text-gray-500">Нет проектов с низким показателем здоровья</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
