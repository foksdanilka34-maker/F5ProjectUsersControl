import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  AlertTriangle,
  ArrowLeft,
  BadgeCheck,
  Building2,
  CheckCircle2,
  Edit,
  Layers3,
  ListChecks,
  Search,
  Settings2,
  ShieldCheck,
  UserCog,
  UserPlus,
  Users,
} from 'lucide-react';

import { useEmployeesData } from '../hooks/useEmployeesData';

const tabs = ['Профили', 'Оргструктура', 'Навыки'];
const statusStyles: Record<string, string> = {
  Активен: 'bg-emerald-50 text-emerald-700',
  'В отпуске': 'bg-amber-50 text-amber-700',
  Заблокирован: 'bg-rose-50 text-rose-700',
};

export default function EmployeesHubPage() {
  const [activeTab, setActiveTab] = useState('Профили');
  const [searchTerm, setSearchTerm] = useState('');

  const { profiles: profilesState, departments: departmentsState, positions: positionsState, skills: skillsState, stats: profileStats } = useEmployeesData();

  const filteredProfiles = useMemo(() => {
    const dataset = profilesState.items;
    if (!searchTerm.trim()) {
      return dataset;
    }
    const normalized = searchTerm.toLowerCase();
    return dataset.filter((profile) => {
      return (
        `${profile.first_name} ${profile.last_name}`.toLowerCase().includes(normalized) ||
        profile.position_id.toLowerCase().includes(normalized) ||
        profile.department?.name?.toLowerCase().includes(normalized)
      );
    });
  }, [profilesState.items, searchTerm]);

  return (
    <div className="relative min-h-screen bg-gradient-to-br from-white via-emerald-50/40 to-white text-gray-900">
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -left-24 top-10 h-[420px] w-[420px] rounded-full bg-emerald-100/60 blur-3xl" />
        <div className="absolute -right-16 top-52 h-[360px] w-[360px] rounded-full bg-lime-100/60 blur-3xl" />
      </div>

      <div className="relative z-10 mx-auto max-w-7xl px-6 py-10">
        <header className="flex flex-wrap items-center justify-between gap-4 border-b border-gray-100 pb-6">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.35em] text-emerald-500">Сотрудники</p>
            <h1 className="mt-3 text-3xl font-semibold text-gray-900">Центр управления командами</h1>
            <p className="mt-2 text-sm text-gray-500">
              Все операции employee-service: профили, отделы, должности, навыки и статусы — в одном интерфейсе.
            </p>
          </div>
          <Link
            to="/"
            className="inline-flex items-center gap-2 rounded-full border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-600 hover:text-emerald-600"
          >
            <ArrowLeft className="h-4 w-4" />
            К дашборду
          </Link>
        </header>

        <div className="mt-8 flex flex-wrap gap-3">
          {tabs.map((tab) => (
            <button
              key={tab}
              type="button"
              onClick={() => setActiveTab(tab)}
              className={`rounded-full px-5 py-2 text-sm font-semibold transition-all ${
                activeTab === tab ? 'bg-gray-900 text-white' : 'border border-gray-200 bg-white text-gray-600 hover:text-emerald-600'
              }`}
            >
              {tab}
            </button>
          ))}
        </div>

        <div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {profileStats.map((stat) => (
            <div
              key={stat.label}
              className="rounded-3xl border border-gray-100 bg-white/90 px-4 py-5 text-sm shadow-[0_15px_35px_rgba(6,95,70,0.08)]"
            >
              <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">{stat.label}</p>
              <p className="mt-2 text-2xl font-semibold text-gray-900">{stat.value}</p>
            </div>
          ))}
        </div>

        <section className="mt-8 grid gap-6 lg:grid-cols-[2fr,1fr]">
          <div className="space-y-6">
            <div className="rounded-[32px] border border-gray-100 bg-white/95 p-6 shadow-[0_25px_80px_rgba(6,95,70,0.08)]">
              <div className="flex flex-wrap items-center justify-between gap-4">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Быстрые действия</p>
                  <h2 className="mt-2 text-xl font-semibold">Что вы хотите сделать?</h2>
                </div>
                <label className="relative flex items-center">
                  <Search className="absolute left-3 h-4 w-4 text-gray-400" />
                  <input
                    type="text"
                    value={searchTerm}
                    onChange={(event) => setSearchTerm(event.target.value)}
                    placeholder="Команда, логин, навык..."
                    className="w-64 rounded-full border border-gray-200 bg-white px-9 py-2 text-sm text-gray-600 focus:border-emerald-400 focus:outline-none"
                  />
                  {searchTerm && (
                    <button
                      type="button"
                      onClick={() => setSearchTerm('')}
                      className="absolute right-3 text-xs font-semibold text-emerald-600"
                    >
                      Очистить
                    </button>
                  )}
                </label>
              </div>
              <div className="mt-5 grid gap-4 md:grid-cols-3">
                <Link
                  to="/admin/employees/new"
                  className="rounded-3xl border border-emerald-100 bg-emerald-50/70 p-4 text-sm font-semibold text-emerald-800 shadow-[0_15px_40px_rgba(16,185,129,0.25)]"
                >
                  <UserPlus className="mb-2 h-5 w-5" />
                  Добавить профиль
                  <p className="mt-1 text-xs font-normal text-emerald-700">CreateProfileRequest</p>
                </Link>
                <button
                  type="button"
                  disabled
                  className="rounded-3xl border border-gray-100 bg-gray-50/80 p-4 text-sm font-semibold text-gray-400"
                  title="Редактирование появится после подключения employee-service"
                >
                  <UserCog className="mb-2 h-5 w-5 text-emerald-500" />
                  Редактировать данные
                  <p className="mt-1 text-xs font-normal text-gray-400">UpdateProfileRequest (скоро)</p>
                </button>
                <button
                  type="button"
                  disabled
                  className="rounded-3xl border border-gray-100 bg-gray-50/80 p-4 text-sm font-semibold text-gray-400"
                  title="Изменение статусов включим после интеграции"
                >
                  <ShieldCheck className="mb-2 h-5 w-5 text-emerald-500" />
                  Управление статусами
                  <p className="mt-1 text-xs font-normal text-gray-400">ChangeUserStatusProfile (скоро)</p>
                </button>
              </div>
              <p className="mt-3 text-xs text-gray-400">Заблокированные действия станут активными, когда подключим соответствующие endpoint&apos;ы.</p>
            </div>

            {activeTab === 'Профили' && (
              <div className="rounded-[32px] border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
                <div className="flex flex-wrap items-center justify-between gap-4">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Сотрудники</p>
                    <h2 className="mt-2 text-xl font-semibold">Список профилей</h2>
                  </div>
                  <div className="flex gap-2 text-xs text-gray-500">
                    <span>department_id</span>
                    <span className="text-gray-300">·</span>
                    <span>position_id</span>
                  </div>
                </div>
                <div className="mt-5 overflow-hidden rounded-[28px] border border-gray-100">
                  {profilesState.loading ? (
                    <div className="flex flex-col items-center justify-center bg-white/80 px-6 py-12 text-center text-sm text-gray-500">
                      <Settings2 className="mb-3 h-6 w-6 animate-spin text-emerald-500" />
                      Загружаем профили...
                    </div>
                  ) : filteredProfiles.length ? (
                    <table className="w-full text-left text-sm text-gray-600">
                      <thead className="bg-gray-50 text-xs uppercase tracking-wide text-gray-400">
                        <tr>
                          <th className="px-6 py-4">Сотрудник</th>
                          <th className="px-6 py-4">Роль</th>
                          <th className="px-6 py-4">Отдел</th>
                          <th className="px-6 py-4">Статус</th>
                          <th className="px-6 py-4">Действия</th>
                        </tr>
                      </thead>
                      <tbody>
                        {filteredProfiles.map((profile) => (
                          <tr key={profile.id} className="border-t border-gray-100 bg-white/80">
                            <td className="px-6 py-4 font-semibold text-gray-900">
                              {profile.first_name} {profile.last_name}
                              <p className="text-xs text-gray-400">{profile.email}</p>
                            </td>
                            <td className="px-6 py-4">{profile.position_id}</td>
                            <td className="px-6 py-4">{profile.department?.name ?? '—'}</td>
                            <td className="px-6 py-4">
                              <span
                                className={`rounded-full px-3 py-1 text-xs font-semibold ${
                                  profile.is_active ? statusStyles['Активен'] : statusStyles['Заблокирован']
                                }`}
                              >
                                {profile.is_active ? 'Активен' : 'Заблокирован'}
                              </span>
                            </td>
                            <td className="px-6 py-4">
                              <div className="flex gap-2">
                                <button
                                  type="button"
                                  disabled
                                  className="rounded-full border border-gray-200 px-3 py-1 text-xs text-gray-400 disabled:cursor-not-allowed"
                                  title="Редактирование профилей появится после интеграции"
                                >
                                  <Edit className="mr-1 inline h-3.5 w-3.5" /> Редактировать
                                </button>
                                <button
                                  type="button"
                                  disabled
                                  className="rounded-full border border-gray-200 px-3 py-1 text-xs text-gray-400 disabled:cursor-not-allowed"
                                  title="Управление статусами появится после интеграции"
                                >
                                  <CheckCircle2 className="mr-1 inline h-3.5 w-3.5" /> Статус
                                </button>
                              </div>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  ) : (
                    <div className="flex flex-col items-center justify-center bg-white/80 px-6 py-12 text-center text-sm text-gray-500">
                      <BadgeCheck className="mb-3 h-6 w-6 text-emerald-500" />
                      {profilesState.error
                        ? `Не удалось загрузить профили: ${profilesState.error}`
                        : searchTerm
                          ? `Нет совпадений по запросу «${searchTerm}». Перенастройте фильтр.`
                          : 'Данные от employee-service пока недоступны.'}
                    </div>
                  )}
                </div>
                <p className="mt-2 text-xs text-gray-400">
                  Запрос: GET /api/v1/employees/profiles. Пагинацию добавим позже.
                </p>
              </div>
            )}

            {activeTab === 'Оргструктура' && (
              <div className="grid gap-6 lg:grid-cols-2">
                <div className="rounded-[32px] border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Отделы</p>
                      <h2 className="mt-2 text-xl font-semibold">Управление департаментами</h2>
                    </div>
                    <button
                      className="rounded-full border border-gray-200 px-4 py-2 text-xs font-semibold text-gray-400 disabled:cursor-not-allowed"
                      type="button"
                      disabled
                      title="CRUD по департаментам появится после интеграции"
                    >
                      <Building2 className="mr-1 inline h-3.5 w-3.5" /> Добавить
                    </button>
                  </div>
                  {departmentsState.loading ? (
                    <p className="mt-5 text-sm text-gray-500">Загружаем данные...</p>
                  ) : departmentsState.items.length ? (
                    <ul className="mt-5 space-y-3 text-sm text-gray-600">
                      {departmentsState.items.map((dep) => (
                        <li key={dep.id} className="rounded-3xl border border-gray-100 bg-gray-50/80 px-4 py-3">
                          <div className="flex items-center justify-between">
                            <p className="font-semibold text-gray-900">{dep.name}</p>
                            <span className="text-xs text-gray-400">ID: {dep.id}</span>
                          </div>
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="mt-5 rounded-3xl border border-dashed border-gray-200 px-4 py-6 text-sm text-gray-500">
                      {departmentsState.error ? `Ошибка: ${departmentsState.error}` : 'Список департаментов пуст.'}
                    </p>
                  )}
                </div>
                <div className="rounded-[32px] border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Должности</p>
                      <h2 className="mt-2 text-xl font-semibold">Каталог позиций</h2>
                    </div>
                    <button
                      className="rounded-full border border-gray-200 px-4 py-2 text-xs font-semibold text-gray-400 disabled:cursor-not-allowed"
                      type="button"
                      disabled
                      title="CRUD по должностям появится после интеграции"
                    >
                      <Layers3 className="mr-1 inline h-3.5 w-3.5" /> Добавить
                    </button>
                  </div>
                  {positionsState.loading ? (
                    <p className="mt-5 text-sm text-gray-500">Загружаем позиции...</p>
                  ) : positionsState.items.length ? (
                    <ul className="mt-5 space-y-3 text-sm text-gray-600">
                      {positionsState.items.map((position) => (
                        <li key={position.id} className="rounded-3xl border border-gray-100 bg-gray-50/80 px-4 py-3">
                          <div className="flex items-center justify-between">
                            <p className="font-semibold text-gray-900">{position.name}</p>
                            <span className="text-xs text-gray-400">ID: {position.id}</span>
                          </div>
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="mt-5 rounded-3xl border border-dashed border-gray-200 px-4 py-6 text-sm text-gray-500">
                      {positionsState.error ? `Ошибка: ${positionsState.error}` : 'Каталог должностей пуст.'}
                    </p>
                  )}
                </div>
              </div>
            )}

            {activeTab === 'Навыки' && (
              <div className="rounded-[32px] border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Навыки</p>
                    <h2 className="mt-2 text-xl font-semibold">Библиотека и назначение</h2>
                  </div>
                  <div className="flex gap-2">
                    <button
                      className="rounded-full border border-gray-200 px-4 py-2 text-xs font-semibold text-gray-400 disabled:cursor-not-allowed"
                      type="button"
                      disabled
                      title="CRUD по навыкам появится после интеграции"
                    >
                      <ListChecks className="mr-1 inline h-3.5 w-3.5" /> Добавить навык
                    </button>
                    <button
                      className="rounded-full border border-gray-200 px-4 py-2 text-xs font-semibold text-gray-400 disabled:cursor-not-allowed"
                      type="button"
                      disabled
                      title="Назначение навыков подключим позже"
                    >
                      <BadgeCheck className="mr-1 inline h-3.5 w-3.5" /> Назначить
                    </button>
                  </div>
                </div>
                {skillsState.loading ? (
                  <p className="mt-5 text-sm text-gray-500">Загружаем навыки...</p>
                ) : skillsState.items.length ? (
                  <div className="mt-5 grid gap-4 md:grid-cols-2">
                    {skillsState.items.map((skill) => (
                      <div key={skill.id} className="rounded-3xl border border-gray-100 bg-gray-50/80 p-4">
                        <p className="text-sm font-semibold text-gray-900">{skill.name}</p>
                        <p className="text-xs text-gray-500">ID: {skill.id}</p>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="mt-5 rounded-3xl border border-dashed border-gray-200 px-4 py-6 text-sm text-gray-500">
                    {skillsState.error ? `Ошибка: ${skillsState.error}` : 'Навыков пока нет.'}
                  </p>
                )}
              </div>
            )}
          </div>

          <aside className="space-y-6">
            <div className="rounded-[32px] border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
              <p className="text-xs font-semibold uppercase tracking-[0.35em] text-emerald-500">API</p>
              <h3 className="mt-3 text-lg font-semibold text-gray-900">EmployeeService маршруты</h3>
              <ul className="mt-4 space-y-3 text-sm text-gray-600">
                <li className="flex items-start gap-3">
                  <Users className="mt-0.5 h-4 w-4 text-emerald-500" /> Create/List/Update Profile
                </li>
                <li className="flex items-start gap-3">
                  <Settings2 className="mt-0.5 h-4 w-4 text-emerald-500" /> Departments & Positions CRUD
                </li>
                <li className="flex items-start gap-3">
                  <ListChecks className="mt-0.5 h-4 w-4 text-emerald-500" /> Skills registry + assignments
                </li>
                <li className="flex items-start gap-3">
                  <AlertTriangle className="mt-0.5 h-4 w-4 text-amber-500" /> Аккаунт статус / блокировки
                </li>
              </ul>
              <p className="mt-4 rounded-2xl bg-emerald-50/80 px-4 py-3 text-xs text-emerald-700">
                Позже подключим gRPC/REST и состояния. Сейчас это референс дизайна.
              </p>
            </div>

            <div className="rounded-[32px] border border-gray-100 bg-white/90 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
              <p className="text-xs font-semibold uppercase tracking-[0.35em] text-gray-400">Процессы</p>
              <ul className="mt-4 space-y-3 text-sm text-gray-600">
                <li className="flex items-start gap-3">
                  <ShieldCheck className="mt-0.5 h-4 w-4 text-emerald-500" />
                  Роли и доступы из login-service (role field)
                </li>
                <li className="flex items-start gap-3">
                  <UserCog className="mt-0.5 h-4 w-4 text-emerald-500" />
                  Аккаунт активен/заморожен через ChangeUserStatusProfile
                </li>
                <li className="flex items-start gap-3">
                  <Settings2 className="mt-0.5 h-4 w-4 text-emerald-500" />
                  Departments/Positions синхронизируются с оргструктурой
                </li>
              </ul>
            </div>
          </aside>
        </section>
      </div>
    </div>
  );
}
