import { useMemo, useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  AlertCircle,
  AlertTriangle,
  ArrowLeft,
  Calendar,
  FileText,
  Loader2,
  Target,
  UserCheck,
} from 'lucide-react';
import { projectService, employeeService } from '../api';
import { useToast } from '../contexts/ToastContext';
import { useAuth } from '../contexts/AuthContext';
import type { Profile } from '../api/types';

type FormState = {
  name: string;
  description: string;
  dueDate: string;
};

const initialState: FormState = {
  name: '',
  description: '',
  dueDate: '',
};

export default function CreateProjectPage() {
  const navigate = useNavigate();
  const { showToast } = useToast();
  const { user } = useAuth();
  const [form, setForm] = useState<FormState>(initialState);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentUserProfile, setCurrentUserProfile] = useState<Profile | null>(null);

  useEffect(() => {
    if (user?.userId) {
      employeeService.getProfile(user.userId)
        .then(setCurrentUserProfile)
        .catch(console.error);
    }
  }, [user?.userId]);

  const isValid = useMemo(() => {
    return form.name.trim().length >= 3;
  }, [form.name]);

  const handleChange = (field: keyof FormState) => (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    setForm((prev) => ({ ...prev, [field]: event.target.value }));
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!isValid) return;
    
    setIsSubmitting(true);
    setError(null);

    try {
      // Convert date to RFC3339 format (ISO 8601) if present
      const dueDateISO = form.dueDate ? new Date(form.dueDate).toISOString() : undefined;

      await projectService.createProject({
        name: form.name,
        description: form.description || undefined,
        due_date: dueDateISO,
      });

      showToast('Проект успешно создан!', 'success');
      navigate('/projects');
    } catch (err: any) {
      console.error('Failed to create project:', err);
      setError(err.response?.data?.error || 'Не удалось создать проект');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="relative min-h-screen bg-linear-to-br from-white via-emerald-50/50 to-white text-gray-900">
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -left-16 top-12 h-[420px] w-[420px] rounded-full bg-emerald-100/60 blur-3xl" />
        <div className="absolute -right-24 top-36 h-[360px] w-[360px] rounded-full bg-lime-100/60 blur-3xl" />
      </div>

      <div className="relative z-10 mx-auto flex max-w-6xl flex-col gap-8 px-6 py-10 lg:flex-row">
        <section className="flex-1 rounded-4xl border border-gray-100 bg-white/95 p-8 shadow-[0_30px_80px_rgba(6,95,70,0.08)]">
          <header className="flex flex-wrap items-center justify-between gap-4 border-b border-gray-100 pb-6">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.35em] text-emerald-500">Менеджмент</p>
              <h1 className="mt-3 text-3xl font-semibold">Создать проект</h1>
              <p className="mt-2 text-sm text-gray-500">
                Заполняем поля gRPC запроса `CreateProjectRequest`: название, manager_id и дедлайн (по желанию).
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

          <form className="mt-8 space-y-8" onSubmit={handleSubmit}>
            <div className="space-y-4">
              <h2 className="text-sm font-semibold uppercase tracking-[0.25em] text-gray-400">Общее</h2>
              <div className="space-y-4">
                <label className="text-sm font-medium text-gray-600">
                  Название проекта
                  <div className="relative">
                    <Target className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                    <input
                      value={form.name}
                      onChange={handleChange('name')}
                      placeholder="Например, Project Nimbus"
                      className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/80 px-12 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                  </div>
                </label>
                <label className="text-sm font-medium text-gray-600">
                  Краткое описание
                  <div className="relative">
                    <FileText className="absolute left-4 top-4 h-4 w-4 text-gray-400" />
                    <textarea
                      value={form.description}
                      onChange={handleChange('description')}
                      placeholder="Цель, ожидаемый результат, метрики успеха"
                      className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/80 px-12 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                      rows={4}
                    />
                  </div>
                </label>
              </div>
            </div>

            <div className="space-y-4">
              <h2 className="text-sm font-semibold uppercase tracking-[0.25em] text-gray-400">Сроки</h2>
              <div className="grid gap-4">
                <label className="text-sm font-medium text-gray-600">
                  Дата завершения (опционально)
                  <div className="relative">
                    <Calendar className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                    <input
                      type="date"
                      value={form.dueDate}
                      onChange={handleChange('dueDate')}
                      className="mt-1 w-full rounded-2xl border border-gray-200 bg-white/80 px-12 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                    />
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

            <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl bg-gray-50/80 px-4 py-4 text-sm text-gray-500">
              <div className="flex items-center gap-2">
                <AlertTriangle className="h-4 w-4 text-emerald-500" />
                Менеджером проекта станет: <span className="font-semibold text-gray-700">
                  {currentUserProfile ? `${currentUserProfile.first_name} ${currentUserProfile.last_name}` : 'Загрузка...'}
                </span>
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setForm(initialState)}
                  className="rounded-full border border-gray-200 px-4 py-2 font-medium text-gray-600"
                >
                  Сбросить
                </button>
                <button
                  type="submit"
                  disabled={!isValid || isSubmitting}
                  className="inline-flex items-center gap-2 rounded-full bg-linear-to-r from-lime-400 to-emerald-400 px-6 py-2 font-semibold text-gray-900 shadow-[0_15px_40px_rgba(132,204,22,0.35)] disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <UserCheck className="h-4 w-4" />}
                  Создать проект
                </button>
              </div>
            </div>
          </form>
        </section>

        <aside className="w-full rounded-4xl border border-gray-100 bg-white/90 p-8 shadow-[0_30px_80px_rgba(6,95,70,0.05)] lg:w-96">
          <p className="text-xs font-semibold uppercase tracking-[0.35em] text-emerald-500">Предпросмотр</p>
          <h2 className="mt-3 text-xl font-semibold text-gray-900">Обложка проекта</h2>
          <div className="mt-6 rounded-[28px] border border-emerald-50 bg-linear-to-br from-white to-emerald-50/50 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-lg font-semibold text-gray-900">{form.name || 'Название проекта'}</p>
                <p className="text-sm text-gray-500">
                  Менеджер: {currentUserProfile ? `${currentUserProfile.first_name} ${currentUserProfile.last_name}` : '...'}
                </p>
              </div>
            </div>
            <p className="mt-4 text-sm text-gray-600">{form.description || 'Описание появится здесь.'}</p>
            <dl className="mt-6 space-y-3 text-sm text-gray-600">
              <div className="flex items-center justify-between">
                <span className="flex items-center gap-2 text-gray-500">
                  <Calendar className="h-4 w-4 text-emerald-500" />
                  Due date
                </span>
                <span className="font-semibold text-gray-800">{form.dueDate || '—'}</span>
              </div>
            </dl>
          </div>

          <div className="mt-8 space-y-4 rounded-[28px] border border-gray-100 bg-white p-6 text-sm text-gray-600">
            <p className="text-xs font-semibold uppercase tracking-[0.35em] text-gray-400">Права</p>
            <p className="flex items-start gap-3">
              <UserCheck className="mt-0.5 h-4 w-4 text-emerald-500" />
              Создавать проекты могут только менеджеры и директора. Остальным кнопка будет задизейблена.
            </p>
            <p className="flex items-start gap-3">
              <AlertTriangle className="mt-0.5 h-4 w-4 text-amber-500" />
              Validation/permissions подключим при интеграции с login-service и project-service.
            </p>
          </div>
        </aside>
      </div>
    </div>
  );
}
