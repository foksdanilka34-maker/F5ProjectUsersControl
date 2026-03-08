import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  FileText,
  Loader2,
  Target,
  CheckCircle,
  Shield,
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { createProject } from '../services/projectService';

type FormState = {
  name: string;
  description: string;
};

const initialState: FormState = {
  name: '',
  description: '',
};

export default function CreateProjectPage() {
  const { user } = useAuth();
  const [form, setForm] = useState<FormState>(initialState);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const canCreate = user && ['manager', 'developer', 'director', 'admin'].includes(user.role);

  const isValid = form.name.trim().length >= 3;

  const handleChange = (field: keyof FormState) => (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>,
  ) => {
    setForm((prev) => ({ ...prev, [field]: event.target.value }));
    setError(null);
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!isValid || !user) return;
    
    setIsSubmitting(true);
    setError(null);
    
    try {
      const newProject = await createProject({
        name: form.name.trim(),
        description: form.description.trim() || undefined,
        manager_id: user.id, // Автоматически используем текущего пользователя как менеджера
      });

      navigate(`/projects/${newProject.id}`);
    } catch (err) {
      console.error('Failed to create project:', err);
      setError('Не удалось создать проект. Попробуйте ещё раз.');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!canCreate) {
    return (
      <div className="relative min-h-screen bg-gradient-to-br from-white via-emerald-50/50 to-white text-gray-900">
        <div className="relative z-10 mx-auto max-w-2xl px-6 py-10">
          <section className="rounded-3xl border border-gray-100 bg-white/95 p-8 shadow-[0_30px_80px_rgba(6,95,70,0.08)]">
            <div className="text-center py-12">
              <Shield className="h-16 w-16 text-amber-400 mx-auto mb-4" />
              <h1 className="text-2xl font-bold text-gray-900 mb-2">Доступ ограничен</h1>
              <p className="text-gray-500 mb-6">
                Только менеджеры, директора и администраторы могут создавать проекты
              </p>
              <Link
                to="/projects"
                className="inline-flex items-center gap-2 rounded-full bg-emerald-500 px-6 py-3 text-sm font-medium text-white hover:bg-emerald-600"
              >
                К проектам
              </Link>
            </div>
          </section>
        </div>
      </div>
    );
  }

  return (
    <div className="relative min-h-screen bg-gradient-to-br from-white via-emerald-50/50 to-white text-gray-900">
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -left-16 top-12 h-[420px] w-[420px] rounded-full bg-emerald-100/60 blur-3xl" />
        <div className="absolute -right-24 top-36 h-[360px] w-[360px] rounded-full bg-lime-100/60 blur-3xl" />
      </div>

      <div className="relative z-10 mx-auto max-w-2xl px-6 py-10">
        <section className="rounded-3xl border border-gray-100 bg-white/95 p-8 shadow-[0_30px_80px_rgba(6,95,70,0.08)]">
          <header className="border-b border-gray-100 pb-6">
            <p className="text-xs font-semibold uppercase tracking-[0.35em] text-emerald-500">Проекты</p>
            <h1 className="mt-3 text-3xl font-semibold">Создать проект</h1>
            <p className="mt-2 text-sm text-gray-500">
              Укажите название и описание проекта
            </p>
          </header>

          <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
            {error && (
              <div className="rounded-2xl border border-red-100 bg-red-50 px-4 py-3 text-sm text-red-700">
                {error}
              </div>
            )}

            <label className="block text-sm font-medium text-gray-600">
              Название проекта *
              <div className="relative mt-1">
                <Target className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                <input
                  value={form.name}
                  onChange={handleChange('name')}
                  placeholder="Введите название проекта"
                  className="w-full rounded-2xl border border-gray-200 bg-white/80 px-12 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                  required
                  minLength={3}
                />
              </div>
              <p className="mt-1 text-xs text-gray-400">Минимум 3 символа</p>
            </label>

            <label className="block text-sm font-medium text-gray-600">
              Описание
              <div className="relative mt-1">
                <FileText className="absolute left-4 top-4 h-4 w-4 text-gray-400" />
                <textarea
                  value={form.description}
                  onChange={handleChange('description')}
                  placeholder="Опишите цели и задачи проекта"
                  className="w-full rounded-2xl border border-gray-200 bg-white/80 px-12 py-3 text-sm focus:border-emerald-400 focus:outline-none"
                  rows={4}
                />
              </div>
            </label>

            <div className="flex flex-wrap items-center justify-end gap-3 pt-4">
              <button
                type="button"
                onClick={() => setForm(initialState)}
                className="rounded-full border border-gray-200 px-4 py-2 text-sm font-medium text-gray-600 hover:bg-gray-50"
              >
                Сбросить
              </button>
              <button
                type="submit"
                disabled={!isValid || isSubmitting}
                className="inline-flex items-center gap-2 rounded-full bg-gradient-to-r from-lime-400 to-emerald-400 px-6 py-2.5 text-sm font-semibold text-gray-900 shadow-[0_15px_40px_rgba(132,204,22,0.35)] hover:shadow-[0_15px_40px_rgba(132,204,22,0.45)] disabled:cursor-not-allowed disabled:opacity-60 transition-shadow"
              >
                {isSubmitting ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <CheckCircle className="h-4 w-4" />
                )}
                Создать проект
              </button>
            </div>
          </form>
        </section>

        {}
        {form.name && (
          <div className="mt-6 rounded-3xl border border-gray-100 bg-white/90 p-6 shadow-[0_20px_60px_rgba(6,95,70,0.05)]">
            <p className="text-xs font-semibold uppercase tracking-[0.35em] text-gray-400">Предпросмотр</p>
            <div className="mt-4 rounded-2xl border border-emerald-50 bg-gradient-to-br from-white to-emerald-50/50 p-4">
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 rounded-xl bg-gradient-to-r from-emerald-400 to-lime-400 flex items-center justify-center text-white font-semibold">
                  {form.name.charAt(0).toUpperCase()}
                </div>
                <div>
                  <p className="font-semibold text-gray-900">{form.name}</p>
                  <p className="text-sm text-gray-500">{form.description || 'Без описания'}</p>
                  <p className="text-xs text-emerald-600 mt-1">
                    Менеджер: {user?.full_name || 'Вы'}
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}


