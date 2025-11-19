import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  AlertCircle,
  ArrowLeft,
  Calendar,
  Loader2,
  Lock,
  Mail,
  Shield,
  User,
  UserPlus,
  Users,
} from 'lucide-react';
import { createProfile } from '../services/employeeService';
import type { CreateProfileRequest } from '../services/types';
import { ApiError } from '../lib/apiClient';
import { useEmployeeReferences } from '../hooks/useEmployeeReferences';

const roles = [
  { label: 'Admin', value: 'admin' },
  { label: 'Director', value: 'director' },
  { label: 'Manager', value: 'manager' },
  { label: 'Employee', value: 'employee' },
];

type FormState = {
  firstName: string;
  lastName: string;
  email: string;
  departmentId: string;
  positionId: string;
  hireDate: string;
  login: string;
  password: string;
  role: string;
};

const initialState: FormState = {
  firstName: '',
  lastName: '',
  email: '',
  departmentId: '',
  positionId: '',
  hireDate: '',
  login: '',
  password: '',
  role: roles[3].value,
};

export default function CreateEmployeePage() {
  const [form, setForm] = useState<FormState>(initialState);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitSuccess, setSubmitSuccess] = useState<string | null>(null);
  const { departments, positions } = useEmployeeReferences();

  const isValid = useMemo(() => {
    return (
      form.firstName.trim().length > 1 &&
      form.lastName.trim().length > 1 &&
      form.email.includes('@') &&
      form.positionId.trim().length > 0 &&
      form.login.trim().length >= 3 &&
      form.password.trim().length >= 6 &&
      Boolean(form.hireDate)
    );
  }, [form]);

  const handleChange = (field: keyof FormState) => (
    event: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>,
  ) => {
    setForm((prev) => ({ ...prev, [field]: event.target.value }));
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!isValid || isSubmitting) return;

    setIsSubmitting(true);
    setSubmitError(null);
    setSubmitSuccess(null);

    try {
      const payload = buildCreatePayload(form);
      const response = await createProfile(payload);
      const profile = response.data;
      const message = response.message ?? 'Профиль создан';

      setSubmitSuccess(
        profile ? `${message}: ${profile.first_name} ${profile.last_name} (${profile.position_id})` : message,
      );
      setForm(initialState);
      setShowPassword(false);
    } catch (error) {
      setSubmitError(getErrorMessage(error));
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleReset = () => {
    setForm(initialState);
    setSubmitError(null);
    setSubmitSuccess(null);
  };

  return (
    <div className="relative min-h-screen bg-gradient-to-br from-white via-emerald-50/40 to-white text-gray-900">
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -left-16 top-16 h-96 w-96 rounded-full bg-emerald-100/60 blur-3xl" />
        <div className="absolute -right-24 top-40 h-80 w-80 rounded-full bg-lime-100/60 blur-3xl" />
      </div>

      <div className="relative z-10 mx-auto flex max-w-6xl flex-col gap-8 px-6 py-10 lg:flex-row">
        <section className="flex-1 rounded-[32px] border border-gray-100 bg-white/95 p-8 shadow-[0_30px_80px_rgba(6,95,70,0.08)]">
          <header className="flex flex-wrap items-center justify-between gap-4 border-b border-gray-100 pb-6">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Администрирование</p>
              <h1 className="mt-3 text-3xl font-semibold">Создать профиль сотрудника</h1>
              <p className="mt-2 text-sm text-gray-500">
                Поля повторяют контракт `CreateProfileRequest`, запрос уходит напрямую в employee-service через ApiGateway.
              </p>
            </div>
            <Link
              to="/"
              className="inline-flex items-center gap-2 rounded-full border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-600 hover:text-emerald-600"
            >
              <ArrowLeft className="h-4 w-4" />
              Назад к дашборду
            </Link>
          </header>

          {submitSuccess && (
            <div className="mt-6 rounded-3xl border border-emerald-100 bg-emerald-50/80 px-4 py-3 text-sm text-emerald-700">
              {submitSuccess}
            </div>
          )}

          {submitError && (
            <div className="mt-6 flex items-start gap-3 rounded-3xl border border-rose-100 bg-rose-50/80 px-4 py-3 text-sm text-rose-700">
              <AlertCircle className="mt-0.5 h-4 w-4" />
              <span>{submitError}</span>
            </div>
          )}

          <form className="mt-8 space-y-8" onSubmit={handleSubmit}>
            <div className="space-y-4">
              <h2 className="text-sm font-semibold uppercase tracking-[0.2em] text-gray-400">Данные сотрудника</h2>
              <div className="grid gap-4 md:grid-cols-2">
                <label className="text-sm font-medium text-gray-600">
                  Имя
                  <div className="relative">
                    <User className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                    <input
                      value={form.firstName}
                      onChange={handleChange('firstName')}
                      placeholder="Антон"
                      disabled={isSubmitting}
                      className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/80 px-12 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                  </div>
                </label>
                <label className="text-sm font-medium text-gray-600">
                  Фамилия
                  <div className="relative">
                    <Users className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                    <input
                      value={form.lastName}
                      onChange={handleChange('lastName')}
                      placeholder="Кузнецов"
                      disabled={isSubmitting}
                      className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/80 px-12 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                  </div>
                </label>
                <label className="text-sm font-medium text-gray-600">
                  Email
                  <div className="relative">
                    <Mail className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                    <input
                      type="email"
                      value={form.email}
                      onChange={handleChange('email')}
                      placeholder="name@company.ru"
                      disabled={isSubmitting}
                      className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/80 px-12 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                  </div>
                </label>
              </div>
            </div>

            <div className="space-y-4">
              <h2 className="text-sm font-semibold uppercase tracking-[0.2em] text-gray-400">Роль в компании</h2>
              <div className="grid gap-4 md:grid-cols-2">
                <label className="text-sm font-medium text-gray-600">
                  Отдел (department_id)
                  <div className="relative">
                    <input
                      value={form.departmentId}
                      onChange={handleChange('departmentId')}
                      placeholder="Например, dep-rd"
                      list="department-options"
                      disabled={isSubmitting}
                      className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/80 px-4 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                    <datalist id="department-options">
                      {departments.items.map((department) => (
                        <option key={department.id} value={department.id}>
                          {department.name}
                        </option>
                      ))}
                    </datalist>
                    <p className="mt-2 px-1 text-xs text-gray-400">
                      {departments.loading
                        ? 'Загружаем справочник отделов...'
                        : departments.items.length
                          ? 'Выберите ID из списка или введите вручную.'
                          : 'Справочник пуст — введите ID вручную.'}
                    </p>
                    {departments.error && (
                      <p className="px-1 text-xs text-rose-600">Не удалось загрузить отделы: {departments.error}</p>
                    )}
                  </div>
                </label>
                <label className="text-sm font-medium text-gray-600">
                  Должность (position_id)
                  <div className="relative">
                    <input
                      value={form.positionId}
                      onChange={handleChange('positionId')}
                      placeholder="Например, pos-be"
                      list="position-options"
                      disabled={isSubmitting}
                      className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/80 px-4 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                    <datalist id="position-options">
                      {positions.items.map((position) => (
                        <option key={position.id} value={position.id}>
                          {position.name}
                        </option>
                      ))}
                    </datalist>
                    <p className="mt-2 px-1 text-xs text-gray-400">
                      {positions.loading
                        ? 'Загружаем список позиций...'
                        : positions.items.length
                          ? 'Выберите ID из списка или введите вручную.'
                          : 'Справочник пуст — введите ID вручную.'}
                    </p>
                    {positions.error && (
                      <p className="px-1 text-xs text-rose-600">Не удалось загрузить позиции: {positions.error}</p>
                    )}
                  </div>
                </label>
                <label className="text-sm font-medium text-gray-600">
                  Дата выхода
                  <div className="relative">
                    <Calendar className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                    <input
                      type="date"
                      value={form.hireDate}
                      onChange={handleChange('hireDate')}
                      disabled={isSubmitting}
                      className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/80 px-12 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                  </div>
                </label>
              </div>
            </div>

            <div className="space-y-4">
              <h2 className="text-sm font-semibold uppercase tracking-[0.2em] text-gray-400">Доступ</h2>
              <div className="grid gap-4 md:grid-cols-2">
                <label className="text-sm font-medium text-gray-600">
                  Логин
                  <div className="relative">
                    <Shield className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                    <input
                      value={form.login}
                      onChange={handleChange('login')}
                      placeholder="a.kuznetsov"
                      disabled={isSubmitting}
                      className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/80 px-12 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                  </div>
                </label>
                <label className="text-sm font-medium text-gray-600">
                  Роль доступа
                  <select
                    value={form.role}
                    onChange={handleChange('role')}
                    disabled={isSubmitting}
                    className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/70 px-4 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                  >
                    {roles.map((role) => (
                      <option key={role.value} value={role.value}>
                        {role.label}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="text-sm font-medium text-gray-600">
                  Пароль
                  <div className="relative">
                    <Lock className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                    <input
                      type={showPassword ? 'text' : 'password'}
                      value={form.password}
                      onChange={handleChange('password')}
                      placeholder="Минимум 6 символов"
                      disabled={isSubmitting}
                      className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/80 px-12 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword((prev) => !prev)}
                      disabled={isSubmitting}
                      className="absolute right-4 top-1/2 -translate-y-1/2 text-xs font-semibold text-emerald-600 disabled:text-gray-400"
                    >
                      {showPassword ? 'Скрыть' : 'Показать'}
                    </button>
                  </div>
                </label>
              </div>
            </div>

            <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl bg-gray-50/70 px-4 py-3 text-sm">
              <div className="flex items-center gap-2 text-gray-500">
                <AlertCircle className="h-4 w-4 text-amber-500" />
                POST /api/v1/employees/profiles → CreateProfileRequest уходит в employee-service.
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={handleReset}
                  disabled={isSubmitting}
                  className="rounded-full border border-gray-200 px-4 py-2 font-medium text-gray-600 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  Очистить
                </button>
                <button
                  type="submit"
                  disabled={!isValid || isSubmitting}
                  className="inline-flex items-center gap-2 rounded-full bg-gradient-to-r from-lime-400 to-emerald-400 px-6 py-2 font-semibold text-gray-900 shadow-[0_15px_40px_rgba(132,204,22,0.35)] disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <UserPlus className="h-4 w-4" />}
                  Создать профиль
                </button>
              </div>
            </div>
          </form>
        </section>

        <aside className="w-full rounded-[32px] border border-gray-100 bg-white/90 p-8 shadow-[0_30px_80px_rgba(6,95,70,0.05)] lg:w-96">
          <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Предпросмотр</p>
          <h2 className="mt-3 text-xl font-semibold text-gray-900">Карточка сотрудника</h2>
          <div className="mt-6 rounded-[28px] border border-emerald-50 bg-gradient-to-br from-white to-emerald-50/50 p-6">
            <div className="flex items-center gap-4">
              <div className="h-14 w-14 rounded-2xl bg-gradient-to-br from-emerald-400 to-lime-400" />
              <div>
                <p className="text-lg font-semibold text-gray-900">
                  {form.firstName || 'Имя'} {form.lastName || 'Фамилия'}
                </p>
                <p className="text-sm text-gray-500">{form.positionId || '—'}</p>
              </div>
            </div>
            <dl className="mt-6 space-y-3 text-sm text-gray-600">
              <div className="flex justify-between border-b border-gray-100 pb-2">
                <span>Отдел (ID)</span>
                <span className="font-semibold text-gray-800">{form.departmentId || '—'}</span>
              </div>
              <div className="flex justify-between border-b border-gray-100 pb-2">
                <span>Должность (ID)</span>
                <span className="font-semibold text-gray-800">{form.positionId || '—'}</span>
              </div>
              <div className="flex justify-between">
                <span>Дата выхода</span>
                <span className="font-semibold text-gray-800">{form.hireDate || '—'}</span>
              </div>
            </dl>
          </div>

          <div className="mt-8 space-y-4 rounded-[28px] border border-gray-100 bg-white p-6">
            <p className="text-xs font-semibold uppercase tracking-[0.3em] text-gray-400">Регламент</p>
            <ul className="space-y-3 text-sm text-gray-600">
              <li className="flex items-start gap-3">
                <Shield className="mt-0.5 h-4 w-4 text-emerald-500" />
                Только администраторы видят этот экран.
              </li>
              <li className="flex items-start gap-3">
                <Lock className="mt-0.5 h-4 w-4 text-emerald-500" />
                Роли соответствуют login-service (`role` поле в CreateCredentialsRequest).
              </li>
              <li className="flex items-start gap-3">
                <AlertCircle className="mt-0.5 h-4 w-4 text-amber-500" />
                После интеграции добавим валидацию по API (существование логина, доступные отделы, и т.д.).
              </li>
            </ul>
          </div>
        </aside>
      </div>
    </div>
  );
}

function buildCreatePayload(form: FormState): CreateProfileRequest {
  const trimmedDepartment = form.departmentId.trim();
  const hireDateIso = formatHireDate(form.hireDate);

  return {
    first_name: form.firstName.trim(),
    last_name: form.lastName.trim(),
    position_id: form.positionId.trim(),
    email: form.email.trim(),
    hire_date: hireDateIso,
    login: form.login.trim(),
    password: form.password,
    role: form.role,
    ...(trimmedDepartment ? { department_id: trimmedDepartment } : {}),
  };
}

function formatHireDate(date: string): string {
  if (!date) return new Date().toISOString();
  return new Date(`${date}T00:00:00Z`).toISOString();
}

function getErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    const payload = error.payload;
    if (payload && typeof payload === 'object') {
      if ('error' in payload && payload.error) {
        return String(payload.error);
      }
      if ('message' in payload && payload.message) {
        return String(payload.message);
      }
    }
    return `Ошибка ${error.status}`;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return 'Не удалось создать профиль';
}
