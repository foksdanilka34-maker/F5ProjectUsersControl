import { useMemo, useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
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
import { employeeService, type Department, type Position, type UserRole } from '../api';

const roles: Array<{ label: string; value: UserRole }> = [
  { label: 'Специалист', value: 'specialist' },
  { label: 'Менеджер', value: 'manager' },
  { label: 'Директор', value: 'director' },
  { label: 'Администратор', value: 'admin' },
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
  role: UserRole;
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
  role: 'specialist',
};

export default function CreateEmployeePage() {
  const navigate = useNavigate();
  const [form, setForm] = useState<FormState>(initialState);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  // Справочники
  const [departments, setDepartments] = useState<Department[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [isLoadingDictionaries, setIsLoadingDictionaries] = useState(true);

  // Загрузка справочников при монтировании
  useEffect(() => {
    let isMounted = true;

    const loadDictionaries = async () => {
      try {
        const [depts, pos] = await Promise.all([
          employeeService.listDepartments().catch(err => {
            console.error('Failed to load departments:', err);
            return [];
          }),
          employeeService.listPositions().catch(err => {
            console.error('Failed to load positions:', err);
            return [];
          }),
        ]);
        
        if (!isMounted) return;

        const safeDepts = Array.isArray(depts) ? depts : [];
        const safePos = Array.isArray(pos) ? pos : [];

        setDepartments(safeDepts);
        setPositions(safePos);
        
        // Устанавливаем первые значения по умолчанию
        if (safeDepts.length > 0 && !form.departmentId) {
          setForm(prev => ({ ...prev, departmentId: safeDepts[0].id }));
        }
        if (safePos.length > 0 && !form.positionId) {
          setForm(prev => ({ ...prev, positionId: safePos[0].id }));
        }
      } catch (err) {
        console.error('Failed to load dictionaries:', err);
        if (isMounted) {
          setError('Не удалось загрузить справочники. Попробуйте обновить страницу.');
          setDepartments([]);
          setPositions([]);
        }
      } finally {
        if (isMounted) {
          setIsLoadingDictionaries(false);
        }
      }
    };

    loadDictionaries();

    return () => {
      isMounted = false;
    };
  }, []);

  const isValid = useMemo(() => {
    return (
      form.firstName.trim().length > 1 &&
      form.lastName.trim().length > 1 &&
      form.email.includes('@') &&
      form.login.trim().length >= 3 &&
      form.password.trim().length >= 6 &&
      Boolean(form.hireDate) &&
      Boolean(form.positionId)
    );
  }, [form]);

  const handleChange = (field: keyof FormState) => (
    event: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>,
  ) => {
    setForm((prev) => ({ ...prev, [field]: event.target.value }));
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!isValid) return;

    setIsSubmitting(true);
    setError(null);

    try {
      await employeeService.createProfile({
        first_name: form.firstName,
        last_name: form.lastName,
        email: form.email,
        hire_date: form.hireDate,
        login: form.login,
        password: form.password,
        role: form.role,
        position_id: form.positionId,
        department_id: form.departmentId || undefined,
      });

      alert('Профиль сотрудника успешно создан!');
      navigate('/admin/employees');
    } catch (err: any) {
      console.error('Failed to create profile:', err);
      setError(err.response?.data?.error || 'Не удалось создать профиль');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="relative min-h-screen bg-linear-to-br from-white via-emerald-50/40 to-white text-gray-900">
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -left-16 top-16 h-96 w-96 rounded-full bg-emerald-100/60 blur-3xl" />
        <div className="absolute -right-24 top-40 h-80 w-80 rounded-full bg-lime-100/60 blur-3xl" />
      </div>

      <div className="relative z-10 mx-auto flex max-w-6xl flex-col gap-8 px-6 py-10 lg:flex-row">
        <section className="flex-1 rounded-4xl border border-gray-100 bg-white/95 p-8 shadow-[0_30px_80px_rgba(6,95,70,0.08)]">
          <header className="flex flex-wrap items-center justify-between gap-4 border-b border-gray-100 pb-6">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Администрирование</p>
              <h1 className="mt-3 text-3xl font-semibold">Создать профиль сотрудника</h1>
              <p className="mt-2 text-sm text-gray-500">
                Поля повторяют контракт `CreateProfileRequest` — позже сюда подключим API employee-service.
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
                  Отдел
                  <select
                    value={form.departmentId}
                    onChange={handleChange('departmentId')}
                    className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/70 px-4 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                  >
                    <option value="">Не выбран</option>
                    {departments.length > 0 ? (
                      departments.map((dep) => (
                        <option key={dep.id} value={dep.id}>{dep.name}</option>
                      ))
                    ) : (
                      <option value="" disabled>Нет доступных отделов</option>
                    )}
                  </select>
                </label>
                <label className="text-sm font-medium text-gray-600">
                  Должность *
                  <select
                    value={form.positionId}
                    onChange={handleChange('positionId')}
                    className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/70 px-4 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                  >
                    <option value="">Выберите должность</option>
                    {positions.length > 0 ? (
                      positions.map((position) => (
                        <option key={position.id} value={position.id}>{position.name}</option>
                      ))
                    ) : (
                      <option value="" disabled>Нет доступных должностей</option>
                    )}
                  </select>
                </label>
                <label className="text-sm font-medium text-gray-600">
                  Дата выхода
                  <div className="relative">
                    <Calendar className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                    <input
                      type="date"
                      value={form.hireDate}
                      onChange={handleChange('hireDate')}
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
                      className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/80 px-12 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                  </div>
                </label>
                <label className="text-sm font-medium text-gray-600">
                  Роль доступа
                  <select
                    value={form.role}
                    onChange={handleChange('role')}
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
                      className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/80 px-12 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword((prev) => !prev)}
                      className="absolute right-4 top-1/2 -translate-y-1/2 text-xs font-semibold text-emerald-600"
                    >
                      {showPassword ? 'Скрыть' : 'Показать'}
                    </button>
                  </div>
                </label>
              </div>
            </div>

            {error && (
              <div className="rounded-2xl bg-rose-50 border border-rose-200 px-4 py-3 text-sm text-rose-700 flex items-center gap-2">
                <AlertCircle className="h-4 w-4" />
                {error}
              </div>
            )}
            
            {isLoadingDictionaries && (
              <div className="rounded-2xl bg-blue-50 border border-blue-200 px-4 py-3 text-sm text-blue-700 flex items-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" />
                Загрузка справочников...
              </div>
            )}

            <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl bg-gray-50/70 px-4 py-3 text-sm">
              <div className="flex items-center gap-2 text-gray-500">
                <AlertCircle className="h-4 w-4 text-emerald-500" />
                Данные будут отправлены на сервер при создании.
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setForm(initialState)}
                  className="rounded-full border border-gray-200 px-4 py-2 font-medium text-gray-600"
                >
                  Очистить
                </button>
                <button
                  type="submit"
                  disabled={!isValid || isSubmitting}
                  className="inline-flex items-center gap-2 rounded-full bg-linear-to-r from-lime-400 to-emerald-400 px-6 py-2 font-semibold text-gray-900 shadow-[0_15px_40px_rgba(132,204,22,0.35)] disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <UserPlus className="h-4 w-4" />}
                  Создать профиль
                </button>
              </div>
            </div>
          </form>
        </section>

        <aside className="w-full rounded-4xl border border-gray-100 bg-white/90 p-8 shadow-[0_30px_80px_rgba(6,95,70,0.05)] lg:w-96">
          <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Предпросмотр</p>
          <h2 className="mt-3 text-xl font-semibold text-gray-900">Карточка сотрудника</h2>
          <div className="mt-6 rounded-[28px] border border-emerald-50 bg-linear-to-br from-white to-emerald-50/50 p-6">
            <div className="flex items-center gap-4">
              <div className="h-14 w-14 rounded-2xl bg-linear-to-br from-emerald-400 to-lime-400" />
              <div>
                <p className="text-lg font-semibold text-gray-900">
                  {form.firstName || 'Имя'} {form.lastName || 'Фамилия'}
                </p>
                <p className="text-sm text-gray-500">
                  {positions.find(p => p.id === form.positionId)?.name || 'Должность'}
                </p>
              </div>
            </div>
            <dl className="mt-6 space-y-3 text-sm text-gray-600">
              <div className="flex justify-between border-b border-gray-100 pb-2">
                <span>Отдел</span>
                <span className="font-semibold text-gray-800">
                  {departments.find(d => d.id === form.departmentId)?.name || 'Не выбран'}
                </span>
              </div>
              <div className="flex justify-between border-b border-gray-100 pb-2">
                <span>Роль</span>
                <span className="font-semibold text-gray-800">{form.role}</span>
              </div>
              <div className="flex justify-between border-b border-gray-100 pb-2">
                <span>Статус</span>
                <span className="font-semibold text-emerald-600">Активен</span>
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
