import { useCallback, useEffect, useState, type FormEvent } from 'react';
import { CheckCircle, Copy, GitBranch, Loader2, Link2Off, Plug, Save } from 'lucide-react';
import {
  deleteIntegration,
  getIntegration,
  saveIntegration,
  testConnection,
  type GitLabIntegration,
} from '../services/gitlabService';

type Props = {
  projectId: number;
  canManage: boolean;
  onError: (message: string) => void;
  onSuccess: (message: string) => void;
};

const emptyForm = {
  base_url: 'https://gitlab.com',
  gitlab_project_id: '',
  access_token: '',
  default_branch: 'main',
  task_key_prefix: 'F5',
  auto_move_in_progress: true,
  auto_move_review: true,
  auto_close_on_merge: true,
  is_active: true,
};

export default function GitLabSettingsCard({ projectId, canManage, onError, onSuccess }: Props) {
  const [integration, setIntegration] = useState<GitLabIntegration | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);

  const applyIntegration = (value: GitLabIntegration | null) => {
    setIntegration(value);
    if (!value) {
      setForm(emptyForm);
      return;
    }
    setForm({
      base_url: value.base_url,
      gitlab_project_id: String(value.gitlab_project_id),
      access_token: '',
      default_branch: value.default_branch,
      task_key_prefix: value.task_key_prefix,
      auto_move_in_progress: value.auto_move_in_progress,
      auto_move_review: value.auto_move_review,
      auto_close_on_merge: value.auto_close_on_merge,
      is_active: value.is_active,
    });
  };

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const state = await getIntegration(projectId);
      applyIntegration(state.integration);
    } catch (err) {
      console.error('Failed to load gitlab integration:', err);
      onError('Не удалось загрузить настройки GitLab');
    } finally {
      setLoading(false);
    }
  }, [projectId, onError]);

  useEffect(() => {
    load();
  }, [load]);

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    const gitlabProjectId = Number(form.gitlab_project_id);
    if (!gitlabProjectId) {
      onError('Укажите числовой ID проекта в GitLab');
      return;
    }

    setSaving(true);
    try {
      const state = await saveIntegration(projectId, {
        base_url: form.base_url,
        gitlab_project_id: gitlabProjectId,
        access_token: form.access_token || undefined,
        default_branch: form.default_branch,
        task_key_prefix: form.task_key_prefix,
        auto_move_in_progress: form.auto_move_in_progress,
        auto_move_review: form.auto_move_review,
        auto_close_on_merge: form.auto_close_on_merge,
        is_active: form.is_active,
      });
      applyIntegration(state.integration);
      onSuccess('Интеграция с GitLab сохранена');
    } catch (err) {
      console.error('Failed to save gitlab integration:', err);
      onError('Не удалось сохранить интеграцию');
    } finally {
      setSaving(false);
    }
  };

  const handleTest = async () => {
    setTesting(true);
    try {
      const project = await testConnection(projectId);
      onSuccess(`Подключено: ${project.path_with_namespace}`);
    } catch (err) {
      console.error('Failed to test gitlab connection:', err);
      onError('GitLab не отвечает — проверьте адрес, ID проекта и токен');
    } finally {
      setTesting(false);
    }
  };

  const handleDisconnect = async () => {
    setSaving(true);
    try {
      await deleteIntegration(projectId);
      applyIntegration(null);
      onSuccess('Интеграция отключена');
    } catch (err) {
      console.error('Failed to delete gitlab integration:', err);
      onError('Не удалось отключить интеграцию');
    } finally {
      setSaving(false);
    }
  };

  const copy = (value: string, label: string) => {
    navigator.clipboard
      .writeText(value)
      .then(() => onSuccess(`${label} скопирован в буфер обмена`))
      .catch(() => onError('Не удалось скопировать значение'));
  };

  if (loading) {
    return (
      <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm flex justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-gray-400" />
      </div>
    );
  }

  return (
    <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <GitBranch className="h-5 w-5 text-orange-500" />
          <h3 className="font-semibold text-gray-900">Интеграция с GitLab</h3>
        </div>
        {integration && (
          <span
            className={`inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs font-medium ${
              integration.is_active ? 'bg-emerald-100 text-emerald-700' : 'bg-gray-100 text-gray-600'
            }`}
          >
            <CheckCircle className="h-3 w-3" />
            {integration.is_active ? 'Подключено' : 'Приостановлено'}
          </span>
        )}
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <label className="text-xs font-medium text-gray-500">Адрес GitLab</label>
            <input
              type="url"
              value={form.base_url}
              onChange={(e) => setForm({ ...form, base_url: e.target.value })}
              placeholder="https://gitlab.com"
              disabled={!canManage}
              className="mt-1 w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none disabled:bg-gray-50 disabled:text-gray-500"
            />
          </div>
          <div>
            <label className="text-xs font-medium text-gray-500">ID проекта в GitLab</label>
            <input
              type="number"
              value={form.gitlab_project_id}
              onChange={(e) => setForm({ ...form, gitlab_project_id: e.target.value })}
              placeholder="12345678"
              required
              disabled={!canManage}
              className="mt-1 w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none disabled:bg-gray-50 disabled:text-gray-500"
            />
          </div>
          <div>
            <label className="text-xs font-medium text-gray-500">Основная ветка</label>
            <input
              type="text"
              value={form.default_branch}
              onChange={(e) => setForm({ ...form, default_branch: e.target.value })}
              disabled={!canManage}
              className="mt-1 w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none disabled:bg-gray-50 disabled:text-gray-500"
            />
          </div>
          <div>
            <label className="text-xs font-medium text-gray-500">Префикс задач</label>
            <input
              type="text"
              value={form.task_key_prefix}
              onChange={(e) => setForm({ ...form, task_key_prefix: e.target.value.toUpperCase() })}
              maxLength={16}
              disabled={!canManage}
              className="mt-1 w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none disabled:bg-gray-50 disabled:text-gray-500"
            />
          </div>
        </div>

        <div>
          <label className="text-xs font-medium text-gray-500">
            Access token {integration?.token_set && <span className="text-emerald-600">(сохранён)</span>}
          </label>
          <input
            type="password"
            value={form.access_token}
            onChange={(e) => setForm({ ...form, access_token: e.target.value })}
            placeholder={integration?.token_set ? 'Оставьте пустым, чтобы не менять' : 'glpat-...'}
            autoComplete="off"
            disabled={!canManage}
            className="mt-1 w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none disabled:bg-gray-50 disabled:text-gray-500"
          />
          <p className="mt-1 text-xs text-gray-400">
            Нужны права api. Токен хранится в зашифрованном виде и не возвращается обратно.
          </p>
        </div>

        <div className="space-y-2">
          {[
            { key: 'auto_move_in_progress' as const, label: 'Пуш в ветку задачи переводит её в «В работе»' },
            { key: 'auto_move_review' as const, label: 'Открытие merge request отправляет задачу на код-ревью' },
            { key: 'auto_close_on_merge' as const, label: 'Влитый merge request закрывает задачу (после подтверждений)' },
            { key: 'is_active' as const, label: 'Интеграция активна' },
          ].map(({ key, label }) => (
            <label key={key} className="flex items-center gap-2 text-sm text-gray-700">
              <input
                type="checkbox"
                checked={form[key]}
                onChange={(e) => setForm({ ...form, [key]: e.target.checked })}
                disabled={!canManage}
                className="h-4 w-4 rounded border-gray-300 text-emerald-500 focus:ring-emerald-400"
              />
              {label}
            </label>
          ))}
        </div>

        {integration && (
          <div className="rounded-xl border border-gray-100 bg-gray-50 p-4 space-y-3">
            <p className="text-xs font-medium text-gray-500">
              Добавьте вебхук в GitLab: Settings → Webhooks (события Push, Merge request, Pipeline, Comments)
            </p>
            {[
              { label: 'URL вебхука', value: integration.webhook_url },
              { label: 'Secret token', value: integration.webhook_secret },
            ].map(({ label, value }) => (
              <div key={label}>
                <label className="text-xs font-medium text-gray-500">{label}</label>
                <div className="mt-1 flex gap-2">
                  <input
                    readOnly
                    value={value}
                    className="w-full rounded-xl border border-gray-200 bg-white px-4 py-2 text-xs font-mono text-gray-700"
                  />
                  <button
                    type="button"
                    onClick={() => copy(value, label)}
                    className="inline-flex items-center rounded-xl border border-gray-200 bg-white px-3 text-gray-500 hover:text-emerald-600"
                  >
                    <Copy className="h-4 w-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        {canManage && (
          <div className="flex flex-wrap gap-2">
            <button
              type="submit"
              disabled={saving}
              className="inline-flex items-center gap-2 rounded-xl bg-gray-900 px-4 py-2.5 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-60 transition-colors"
            >
              {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
              Сохранить
            </button>

            {integration && (
              <>
                <button
                  type="button"
                  onClick={handleTest}
                  disabled={testing}
                  className="inline-flex items-center gap-2 rounded-xl border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-60 transition-colors"
                >
                  {testing ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plug className="h-4 w-4" />}
                  Проверить подключение
                </button>
                <button
                  type="button"
                  onClick={handleDisconnect}
                  disabled={saving}
                  className="inline-flex items-center gap-2 rounded-xl border border-red-100 px-4 py-2.5 text-sm font-medium text-red-600 hover:bg-red-50 disabled:opacity-60 transition-colors"
                >
                  <Link2Off className="h-4 w-4" />
                  Отключить
                </button>
              </>
            )}
          </div>
        )}
      </form>
    </div>
  );
}
