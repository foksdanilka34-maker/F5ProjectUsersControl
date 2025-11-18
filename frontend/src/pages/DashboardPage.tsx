import { useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import type { LucideIcon } from 'lucide-react';
import {
  Activity,
  AlertCircle,
  ArrowDownRight,
  ArrowRight,
  ArrowUpRight,
  Bell,
  Briefcase,
  CheckCircle,
  ChevronDown,
  Filter,
  Plus,
  Search,
  Target,
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

type UserRole = 'admin' | 'director' | 'manager' | 'member';

const navTabs = ['Обзор', 'Проекты', 'Сотрудники', 'Аналитика'];
const periodOptions = [
  { label: 'Неделя', value: 'weekly' },
  { label: 'Месяц', value: 'monthly' },
  { label: 'Квартал', value: 'quarterly' },
];
const departmentOptions = ['Все отделы', 'Разработка', 'Маркетинг', 'HR'];
const projectStatusOptions = ['Все проекты', 'Активные', 'На паузе', 'Риск'];

const metricCards = [
  {
    label: 'Активные сотрудники',
    value: 128,
    delta: '+6,2% vs прошлый месяц',
    trend: 'up' as const,
    icon: Users,
  },
  {
    label: 'Активные проекты',
    value: 24,
    delta: '+2 новых на этой неделе',
    trend: 'up' as const,
    icon: Briefcase,
  },
  {
    label: 'Выполненные задачи',
    value: 486,
    delta: '−4,1% к прошлой неделе',
    trend: 'down' as const,
    icon: CheckCircle,
  },
  {
    label: 'Просроченные задачи',
    value: 17,
    delta: '−3 за последние 24ч',
    trend: 'up' as const,
    icon: AlertCircle,
  },
];

const productivityTrend = [
  { label: 'Пн', efficiency: 72, completed: 64 },
  { label: 'Вт', efficiency: 75, completed: 71 },
  { label: 'Ср', efficiency: 81, completed: 79 },
  { label: 'Чт', efficiency: 77, completed: 68 },
  { label: 'Пт', efficiency: 84, completed: 82 },
  { label: 'Сб', efficiency: 69, completed: 54 },
  { label: 'Вс', efficiency: 66, completed: 48 },
];

const completionTrend = [
  { week: '1-7', onTime: 68, overall: 74 },
  { week: '8-14', onTime: 72, overall: 79 },
  { week: '15-21', onTime: 76, overall: 81 },
  { week: '22-28', onTime: 74, overall: 77 },
  { week: '29-4', onTime: 79, overall: 85 },
];

const kanbanColumns = [
  {
    title: 'Backlog',
    color: 'border-gray-200',
    tasks: [
      { title: 'Формирование KPI отдела маркетинга', owner: 'Мария П.', priority: 'Высокий' },
      { title: 'Проработка ролей в новом проекте Vostok', owner: 'Илья К.', priority: 'Средний' },
    ],
  },
  {
    title: 'В работе',
    color: 'border-emerald-200',
    tasks: [
      { title: 'Релиз мобильного приложения 2.1', owner: 'Команда Mobile', priority: 'Критичный' },
      { title: 'Онбординг 4 новых специалистов', owner: 'HR Team', priority: 'Высокий' },
      { title: 'Подготовка демо по проекту Atlas', owner: 'Product Lab', priority: 'Средний' },
    ],
  },
  {
    title: 'На проверке',
    color: 'border-amber-200',
    tasks: [
      { title: 'Аналитика потребления лицензий', owner: 'FinOps', priority: 'Низкий' },
      { title: 'Фреймворк метрик эффективности команд', owner: 'PMO', priority: 'Высокий' },
    ],
  },
];

const timeline = [
  { title: 'Стендап команды Fusion', time: '09:30', status: 'В работе' },
  { title: 'Дедлайн задач по проекту Northwind', time: '11:00', status: 'Высокий приоритет' },
  { title: 'Согласование бюджета Q1', time: '13:00', status: 'Ожидает' },
  { title: 'One-on-one с руководителем', time: '16:30', status: 'Запланировано' },
];

const leaders = [
  { name: 'Александр М.', role: 'Lead Backend', score: '92', tasks: 48 },
  { name: 'Екатерина Л.', role: 'Product Manager', score: '89', tasks: 41 },
  { name: 'Владислав С.', role: 'QA Lead', score: '87', tasks: 36 },
];

const riskyProjects = [
  { name: 'Project Nimbus', health: 54, eta: '−8 дней', status: 'Высокий риск' },
  { name: 'Atlas 2.0', health: 62, eta: '−3 дня', status: 'Требует внимания' },
  { name: 'Aurora', health: 71, eta: '+2 дня', status: 'На контроле' },
];

const activityFeed = [
  { title: 'Создан новый проект «Siberia»', time: '5 мин назад', type: 'project' },
  { title: 'Антон Н. завершил задачу «Настройка ETL»', time: '24 мин назад', type: 'task' },
  { title: 'К команде Mobile присоединилась Дарья С.', time: '1 ч назад', type: 'employee' },
  { title: 'Показатель on-time rate < 70% в Project Nimbus', time: '2 ч назад', type: 'alert' },
];

const systemStatuses = [
  { label: 'Api Gateway', value: 'Зеленый', icon: CheckCircle, color: 'text-emerald-500' },
  { label: 'Analytics service', value: 'Нагрузка ↑12%', icon: Activity, color: 'text-amber-500' },
  { label: 'Projects DB', value: 'Репликация 25мс', icon: Target, color: 'text-emerald-500' },
];

type TrendMode = 'productivity' | 'completion';

type MetricCardProps = {
  label: string;
  value: number;
  delta: string;
  trend: 'up' | 'down';
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
      <div className="mt-4 flex items-center gap-2 text-sm">
        {trend === 'up' ? (
          <ArrowUpRight className="h-4 w-4 text-emerald-500" />
        ) : (
          <ArrowDownRight className="h-4 w-4 text-rose-500" />
        )}
        <span className={trend === 'up' ? 'text-emerald-600' : 'text-rose-600'}>{delta}</span>
      </div>
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
          ? 'border-emerald-400 bg-gradient-to-r from-lime-200/60 to-emerald-200/60 text-emerald-800 shadow-[0_8px_20px_rgba(34,197,94,0.25)]'
          : 'border-gray-200 bg-white text-gray-600 hover:border-emerald-200 hover:text-emerald-700'
      }`}
    >
      {label}
    </button>
  );
}

export default function DashboardPage() {
  const [activeTab, setActiveTab] = useState('Обзор');
  const [period, setPeriod] = useState(periodOptions[1].value);
  const [department, setDepartment] = useState(departmentOptions[0]);
  const [projectFilter, setProjectFilter] = useState(projectStatusOptions[1]);
  const [trendMode, setTrendMode] = useState<TrendMode>('productivity');
  const [userRole] = useState<UserRole>('admin');
  const navigate = useNavigate();

  const permissions = useMemo(
    () => ({
      canCreateProject: userRole === 'manager' || userRole === 'director',
      canViewMonitoring: userRole === 'admin',
      canCreateEmployee: userRole === 'admin',
    }),
    [userRole],
  );

  const heroCopy = useMemo(
    () => ({
      greeting: 'Добро пожаловать, Антон!',
      subTitle: 'Ниже всё, что важно вашей команде сегодня: прогресс проектов, задачи и сигналы риска в одном месте.',
    }),
    [],
  );

  return (
    <div className="relative min-h-screen bg-white text-gray-900">
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -left-24 top-12 h-80 w-80 rounded-full bg-emerald-200/40 blur-3xl" />
        <div className="absolute -right-16 top-36 h-64 w-64 rounded-full bg-lime-200/40 blur-3xl" />
        <div className="absolute left-1/3 top-1/2 h-96 w-96 rounded-full bg-emerald-100/30 blur-3xl" />
      </div>

      <div className="relative z-10">
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
              <div className="hidden items-center rounded-full border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-500 lg:flex">
                <Search className="mr-2 h-4 w-4 text-gray-400" />
                Поиск
              </div>
              <button
                type="button"
                className="rounded-full border border-gray-200 bg-white p-2 text-gray-500 hover:text-emerald-600"
              >
                <Bell className="h-5 w-5" />
              </button>
              <button
                type="button"
                className="flex items-center gap-2 rounded-full border border-gray-200 bg-white px-3 py-1.5 text-sm"
              >
                <span className="h-8 w-8 rounded-full bg-gradient-to-r from-emerald-500 to-lime-500" />
                <div className="text-left">
                  <p className="text-sm font-medium text-gray-700">Антон</p>
                  <p className="text-[11px] uppercase tracking-wide text-gray-400">{userRole}</p>
                </div>
                <ChevronDown className="h-4 w-4 text-gray-400" />
              </button>
            </div>
          </div>
        </header>

        <main className="mx-auto flex max-w-7xl flex-col gap-8 px-6 py-8 lg:flex-row">
          <section className="flex-1 space-y-8">
            <div className="rounded-3xl border border-gray-100 bg-white/90 p-6 shadow-[0_20px_60px_rgba(6,95,70,0.07)]">
              <div className="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
                <div>
                  <p className="text-sm font-semibold uppercase tracking-[0.3em] text-emerald-500">Дашборд</p>
                  <h1 className="mt-3 text-3xl font-semibold text-gray-900">{heroCopy.greeting}</h1>
                  <p className="mt-2 max-w-2xl text-base text-gray-600">{heroCopy.subTitle}</p>
                </div>
                <div className="flex flex-wrap gap-3">
                  <button
                    type="button"
                    className="inline-flex items-center gap-2 rounded-2xl border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:border-emerald-200"
                  >
                    <Filter className="h-4 w-4 text-gray-400" />
                    Сохранить представление
                  </button>
                  {permissions.canCreateProject ? (
                    <Link
                      to="/projects/new"
                      className="inline-flex items-center gap-2 rounded-2xl bg-gradient-to-r from-lime-400 to-emerald-400 px-5 py-2 text-sm font-semibold text-gray-900 shadow-[0_20px_40px_rgba(132,204,22,0.35)]"
                    >
                      <Plus className="h-4 w-4" />
                      Создать проект
                    </Link>
                  ) : (
                    <button
                      type="button"
                      className="inline-flex items-center gap-2 rounded-2xl bg-gray-200 px-5 py-2 text-sm font-semibold text-gray-500 opacity-70"
                      title="Доступно только менеджеру или директору"
                      disabled
                    >
                      <Plus className="h-4 w-4" />
                      Создать проект
                    </button>
                  )}
                  {permissions.canCreateEmployee && (
                    <Link
                      to="/admin/employees/new"
                      className="inline-flex items-center gap-2 rounded-2xl border border-emerald-200 bg-white/80 px-5 py-2 text-sm font-semibold text-emerald-700 shadow-[0_12px_30px_rgba(16,185,129,0.18)]"
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
                {departmentOptions.map((option) => (
                  <QuickFilterChip
                    key={option}
                    label={option}
                    isActive={department === option}
                    onClick={() => setDepartment(option)}
                  />
                ))}
                {projectStatusOptions.map((option) => (
                  <QuickFilterChip
                    key={option}
                    label={option}
                    isActive={projectFilter === option}
                    onClick={() => setProjectFilter(option)}
                  />
                ))}
              </div>

              <div className="mt-6 grid gap-4 md:grid-cols-2">
                <Link
                  to="/projects"
                  className="rounded-3xl border border-emerald-100 bg-emerald-50/70 p-5 text-gray-900 shadow-[0_18px_40px_rgba(16,185,129,0.2)]"
                >
                  <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-600">Проекты</p>
                  <h3 className="mt-2 text-lg font-semibold">Канбан и состав команд</h3>
                  <p className="mt-1 text-sm text-emerald-800">Откройте Projects Hub, чтобы управлять задачами и ролями.</p>
                  <span className="mt-4 inline-flex items-center gap-2 text-sm font-semibold text-emerald-700">
                    Перейти
                    <ArrowRight className="h-4 w-4" />
                  </span>
                </Link>
                <Link
                  to="/admin/employees"
                  className="rounded-3xl border border-gray-100 bg-white/80 p-5 text-gray-900 shadow-[0_18px_40px_rgba(15,118,110,0.1)]"
                >
                  <p className="text-xs font-semibold uppercase tracking-[0.3em] text-gray-400">Сотрудники</p>
                  <h3 className="mt-2 text-lg font-semibold">Центр профилей</h3>
                  <p className="mt-1 text-sm text-gray-600">Настройте отделы, роли и создайте нового сотрудника за пару кликов.</p>
                  <span className="mt-4 inline-flex items-center gap-2 text-sm font-semibold text-gray-700">
                    Открыть hub
                    <ArrowRight className="h-4 w-4" />
                  </span>
                </Link>
              </div>
            </div>

            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
              {metricCards.map((card) => (
                <MetricCard key={card.label} {...card} />
              ))}
            </div>

            <div className="grid gap-6 lg:grid-cols-2">
              <div className="rounded-3xl border border-gray-100 bg-white p-6 shadow-[0_12px_40px_rgba(15,118,110,0.08)]">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-semibold uppercase tracking-widest text-emerald-500">Проекты</p>
                    <h2 className="mt-2 text-xl font-semibold">Статус портфеля</h2>
                  </div>
                  <button type="button" className="text-sm font-medium text-emerald-600 hover:text-emerald-700">
                    Все проекты
                  </button>
                </div>
                <div className="mt-6 grid gap-4 md:grid-cols-3">
                  {kanbanColumns.map((column) => (
                    <div key={column.title} className={`rounded-2xl border ${column.color} bg-gray-50/70 p-4`}>
                      <div className="flex items-center justify-between">
                        <h3 className="text-sm font-semibold text-gray-800">{column.title}</h3>
                        <span className="text-xs text-gray-400">{column.tasks.length}</span>
                      </div>
                      <div className="mt-4 space-y-4">
                        {column.tasks.map((task) => (
                          <div key={task.title} className="rounded-2xl bg-white p-3 shadow-sm">
                            <p className="text-sm font-semibold text-gray-900">{task.title}</p>
                            <p className="mt-1 text-xs text-gray-500">{task.owner}</p>
                            <span className="mt-2 inline-flex rounded-full bg-emerald-50 px-2 py-1 text-xs font-medium text-emerald-700">
                              {task.priority}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              <div className="rounded-3xl border border-gray-100 bg-white p-6 shadow-[0_12px_40px_rgba(15,118,110,0.08)]">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-semibold uppercase tracking-widest text-emerald-500">Сегодня</p>
                    <h2 className="mt-2 text-xl font-semibold">Твоя лента задач</h2>
                  </div>
                  <button type="button" className="text-sm text-gray-500 hover:text-emerald-600">
                    Смотреть календарь
                  </button>
                </div>
                <div className="mt-6 space-y-5">
                  {timeline.map((item) => (
                    <div key={item.title} className="flex items-start gap-4">
                      <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-gray-50 text-gray-700">
                        {item.time}
                      </div>
                      <div>
                        <p className="text-sm font-semibold text-gray-900">{item.title}</p>
                        <p className="text-xs text-gray-500">{item.status}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div className="grid gap-6 lg:grid-cols-2">
              <div className="rounded-3xl border border-gray-100 bg-white p-6 shadow-[0_12px_40px_rgba(15,118,110,0.08)]">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-semibold uppercase tracking-widest text-emerald-500">Динамика</p>
                    <h2 className="mt-2 text-xl font-semibold">Производительность и сроки</h2>
                  </div>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={() => setTrendMode('productivity')}
                      className={`rounded-full px-4 py-1 text-sm font-medium ${
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
                      className={`rounded-full px-4 py-1 text-sm font-medium ${
                        trendMode === 'completion'
                          ? 'bg-gray-900 text-white'
                          : 'text-gray-500 hover:text-emerald-600'
                      }`}
                    >
                      On-time rate
                    </button>
                  </div>
                </div>
                <div className="mt-6 h-72">
                  {trendMode === 'productivity' ? (
                    <ResponsiveContainer width="100%" height="100%">
                      <LineChart data={productivityTrend}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#d1d5db" />
                        <XAxis dataKey="label" stroke="#9ca3af" />
                        <YAxis stroke="#9ca3af" domain={[50, 100]} />
                        <Tooltip contentStyle={{ borderRadius: 16, borderColor: '#d1fae5' }} />
                        <Line type="monotone" dataKey="efficiency" stroke="#10b981" strokeWidth={3} dot={{ r: 4 }} />
                        <Line type="monotone" dataKey="completed" stroke="#a3e635" strokeWidth={3} dot={{ r: 4 }} />
                      </LineChart>
                    </ResponsiveContainer>
                  ) : (
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart data={completionTrend}>
                        <defs>
                          <linearGradient id="colorOnTime" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#10b981" stopOpacity={0.4} />
                            <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                          </linearGradient>
                          <linearGradient id="colorOverall" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#a3e635" stopOpacity={0.4} />
                            <stop offset="95%" stopColor="#a3e635" stopOpacity={0} />
                          </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="4 4" stroke="#d1d5db" />
                        <XAxis dataKey="week" stroke="#9ca3af" />
                        <YAxis stroke="#9ca3af" domain={[50, 100]} />
                        <Tooltip contentStyle={{ borderRadius: 16, borderColor: '#d1fae5' }} />
                        <Area type="monotone" dataKey="onTime" stroke="#10b981" fillOpacity={1} fill="url(#colorOnTime)" />
                        <Area type="monotone" dataKey="overall" stroke="#a3e635" fillOpacity={1} fill="url(#colorOverall)" />
                      </AreaChart>
                    </ResponsiveContainer>
                  )}
                </div>
              </div>

              <div className="rounded-3xl border border-gray-100 bg-white p-6 shadow-[0_12px_40px_rgba(15,118,110,0.08)]">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-semibold uppercase tracking-widest text-emerald-500">Фокус</p>
                    <h2 className="mt-2 text-xl font-semibold">Лидеры и зоны риска</h2>
                  </div>
                  <button type="button" className="text-sm text-gray-500 hover:text-emerald-600">
                    Экспорт
                  </button>
                </div>
                <div className="mt-6 grid gap-4 md:grid-cols-2">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-widest text-gray-400">Лидеры</p>
                    <div className="mt-3 space-y-3">
                      {leaders.map((leader) => (
                        <div key={leader.name} className="rounded-2xl border border-gray-100 p-4">
                          <p className="text-sm font-semibold text-gray-900">{leader.name}</p>
                          <p className="text-xs text-gray-500">{leader.role}</p>
                          <div className="mt-2 flex items-center justify-between text-sm">
                            <span className="font-semibold text-emerald-600">{leader.score} баллов</span>
                            <span className="text-gray-500">{leader.tasks} задач</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-widest text-gray-400">Проекты риска</p>
                    <div className="mt-3 space-y-3">
                      {riskyProjects.map((project) => (
                        <div key={project.name} className="rounded-2xl border border-gray-100 p-4">
                          <p className="text-sm font-semibold text-gray-900">{project.name}</p>
                          <p className="text-xs text-gray-500">{project.status}</p>
                          <div className="mt-2 flex items-center justify-between text-sm">
                            <span className="font-semibold text-rose-600">Health {project.health}%</span>
                            <span className="text-gray-500">ETA {project.eta}</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <aside className="w-full space-y-6 lg:w-80">
            {permissions.canViewMonitoring && (
              <div className="rounded-3xl border border-gray-100 bg-white p-6 shadow-[0_12px_40px_rgba(15,118,110,0.08)]">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-widest text-emerald-500">Системы</p>
                    <h3 className="mt-2 text-lg font-semibold">Мониторинг сервисов</h3>
                  </div>
                  <span className="rounded-full bg-emerald-50 px-3 py-1 text-[11px] font-semibold uppercase text-emerald-700">
                    Admin
                  </span>
                </div>
                <div className="mt-4 space-y-4">
                  {systemStatuses.map((status) => (
                    <div key={status.label} className="flex items-center justify-between rounded-2xl border border-gray-50 p-3">
                      <div className="flex items-center gap-3">
                        <status.icon className={`h-5 w-5 ${status.color}`} />
                        <div>
                          <p className="text-sm font-semibold text-gray-900">{status.label}</p>
                          <p className="text-xs text-gray-500">{status.value}</p>
                        </div>
                      </div>
                      <ChevronDown className="h-4 w-4 text-gray-400" />
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div className="rounded-3xl border border-gray-100 bg-white p-6 shadow-[0_12px_40px_rgba(15,118,110,0.08)]">
              <p className="text-xs font-semibold uppercase tracking-widest text-emerald-500">Активность</p>
              <h3 className="mt-2 text-lg font-semibold">Что произошло</h3>
              <div className="mt-4 space-y-4">
                {activityFeed.map((activity) => (
                  <div key={activity.title} className="flex gap-3 rounded-2xl bg-gray-50/70 p-3">
                    <div className="mt-1 h-2 w-2 rounded-full bg-emerald-400" />
                    <div>
                      <p className="text-sm font-semibold text-gray-900">{activity.title}</p>
                      <p className="text-xs text-gray-500">{activity.time}</p>
                    </div>
                  </div>
                ))}
              </div>
              <button type="button" className="mt-4 text-sm font-medium text-emerald-600">
                Смотреть всю ленту
              </button>
            </div>
          </aside>
        </main>
      </div>
    </div>
  );
}
