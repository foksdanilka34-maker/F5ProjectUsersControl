import { useCallback, useEffect, useState, type FormEvent } from 'react';
import { Blocks, ChevronDown, Loader2, Plus, Save, Trash2 } from 'lucide-react';
import {
  EXTENSION_EVENTS,
  deleteExtension,
  listProjectExtensions,
  registerExtension,
  toggleProjectExtension,
  uninstallProjectExtension,
  type ProjectExtension,
} from '../services/extensionsService';

type Props = {
  projectId: number;
  canManage: boolean;
  canRegister: boolean;
  onError: (message: string) => void;
  onSuccess: (message: string) => void;
};

const emptyForm = {
  key: '',
  name: '',
  description: '',
  base_url: '',
  shared_secret: '',
  task_panel_url: '',
  project_tab_url: '',
  project_tab_label: '',
  events: [] as string[],
};

export default function ExtensionsSettingsCard({ projectId, canManage, canRegister, onError, onSuccess }: Props) {
  const [extensions, setExtensions] = useState<ProjectExtension[]>([]);
  const [loading, setLoading] = useState(true);
  const [busyKey, setBusyKey] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      setExtensions(await listProjectExtensions(projectId));
    } catch (err) {
      console.error('Failed to load extensions:', err);
      onError('Не удалось загрузить список расширений');
    } finally {
      setLoading(false);
    }
  }, [projectId, onError]);

  useEffect(() => {
    load();
  }, [load]);

  const handleToggle = async (key: string, enabled: boolean) => {
    setBusyKey(key);
    try {
      await toggleProjectExtension(projectId, key, enabled);
      await load();
      onSuccess(enabled ? 'Расширение включено' : 'Расширение отключено');
    } catch (err) {
      console.error('Failed to toggle extension:', err);
      onError('Не удалось изменить состояние расширения');
    } finally {
      setBusyKey(null);
    }
  };

  const handleUninstall = async (key: string) => {
    setBusyKey(key);
    try {
      await uninstallProjectExtension(projectId, key);
      await load();
      onSuccess('Расширение удалено из проекта');
    } catch (err) {
      console.error('Failed to uninstall extension:', err);
      onError('Не удалось удалить расширение из проекта');
    } finally {
      setBusyKey(null);
    }
  };

  const handleDeleteGlobal = async (key: string) => {
    setBusyKey(key);
    try {
      await deleteExtension(key);
      await load();
      onSuccess('Расширение удалено из реестра');
    } catch (err) {
      console.error('Failed to delete extension:', err);
      onError('Не удалось удалить расширение из реестра');
    } finally {
      setBusyKey(null);
    }
  };

  const toggleEvent = (value: string) => {
    setForm((prev) => ({
      ...prev,
      events: prev.events.includes(value)
        ? prev.events.filter((e) => e !== value)
        : [...prev.events, value],
    }));
  };

  const handleRegister = async (event: FormEvent) => {
    event.preventDefault();
    if (!form.key || !form.name || !form.base_url || !form.shared_secret) {
      onError('Заполните ключ, название, base URL и секрет');
      return;
    }

    setSaving(true);
    try {
      await registerExtension({
        key: form.key,
        name: form.name,
        description: form.description || undefined,
        base_url: form.base_url,
        shared_secret: form.shared_secret,
        task_panel_url: form.task_panel_url || undefined,
        project_tab_url: form.project_tab_url || undefined,
        project_tab_label: form.project_tab_label || undefined,
        events: form.events,
      });
      await load();
      setForm(emptyForm);
      setShowForm(false);
      onSuccess('Расширение зарегистрировано в системе');
    } catch (err) {
      console.error('Failed to register extension:', err);
      onError('Не удалось зарегистрировать расширение — проверьте, что ключ уникален');
    } finally {
      setSaving(false);
    }
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
      <div className="flex items-center gap-2 mb-4">
        <Blocks className="h-5 w-5 text-violet-500" />
        <h3 className="font-semibold text-gray-900">Расширения</h3>
      </div>

      {extensions.length === 0 ? (
        <p className="text-sm text-gray-500 mb-4">Ни одно расширение ещё не зарегистрировано в системе.</p>
      ) : (
        <div className="space-y-2 mb-4">
          {extensions.map((ext) => (
            <div key={ext.key} className="flex items-center justify-between gap-3 rounded-xl border border-gray-100 px-4 py-3">
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-gray-900">{ext.name}</span>
                  <span className="rounded-full bg-gray-100 px-2 py-0.5 font-mono text-[10px] text-gray-500">{ext.key}</span>
                </div>
                {ext.description && <p className="mt-0.5 text-xs text-gray-500 line-clamp-1">{ext.description}</p>}
                <p className="mt-0.5 truncate text-xs text-gray-400">{ext.base_url}</p>
              </div>

              <div className="flex shrink-0 items-center gap-2">
                {canManage && (
                  <label className="inline-flex items-center gap-2 text-xs text-gray-600">
                    <input
                      type="checkbox"
                      checked={ext.enabled}
                      disabled={busyKey === ext.key}
                      onChange={(e) => handleToggle(ext.key, e.target.checked)}
                      className="h-4 w-4 rounded border-gray-300 text-emerald-500 focus:ring-emerald-400"
                    />
                    Включено
                  </label>
                )}
                {canManage && ext.installed && (
                  <button
                    onClick={() => handleUninstall(ext.key)}
                    disabled={busyKey === ext.key}
                    className="text-gray-400 hover:text-red-500 disabled:opacity-50"
                    title="Удалить из проекта"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                )}
                {canRegister && (
                  <button
                    onClick={() => handleDeleteGlobal(ext.key)}
                    disabled={busyKey === ext.key}
                    className="text-gray-300 hover:text-red-500 disabled:opacity-50"
                    title="Удалить из реестра совсем"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {canRegister && (
        <div>
          <button
            type="button"
            onClick={() => setShowForm((v) => !v)}
            className="inline-flex items-center gap-1.5 text-sm font-medium text-violet-600 hover:text-violet-700"
          >
            {showForm ? <ChevronDown className="h-4 w-4" /> : <Plus className="h-4 w-4" />}
            Зарегистрировать новое расширение
          </button>

          {showForm && (
            <form onSubmit={handleRegister} className="mt-4 space-y-3 rounded-xl border border-gray-100 bg-gray-50 p-4">
              <div className="grid gap-3 sm:grid-cols-2">
                <div>
                  <label className="text-xs font-medium text-gray-500">Ключ (уникальный)</label>
                  <input
                    value={form.key}
                    onChange={(e) => setForm({ ...form, key: e.target.value })}
                    placeholder="demo-timer"
                    className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:border-violet-400 focus:outline-none"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-gray-500">Название</label>
                  <input
                    value={form.name}
                    onChange={(e) => setForm({ ...form, name: e.target.value })}
                    placeholder="Демо-чеклист"
                    className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:border-violet-400 focus:outline-none"
                  />
                </div>
                <div className="sm:col-span-2">
                  <label className="text-xs font-medium text-gray-500">Base URL внешнего сервиса</label>
                  <input
                    value={form.base_url}
                    onChange={(e) => setForm({ ...form, base_url: e.target.value })}
                    placeholder="http://localhost:8090"
                    className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:border-violet-400 focus:outline-none"
                  />
                </div>
                <div className="sm:col-span-2">
                  <label className="text-xs font-medium text-gray-500">Секрет (общий с расширением)</label>
                  <input
                    type="password"
                    value={form.shared_secret}
                    onChange={(e) => setForm({ ...form, shared_secret: e.target.value })}
                    autoComplete="off"
                    className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:border-violet-400 focus:outline-none"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-gray-500">Путь панели задачи</label>
                  <input
                    value={form.task_panel_url}
                    onChange={(e) => setForm({ ...form, task_panel_url: e.target.value })}
                    placeholder="/panel"
                    className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:border-violet-400 focus:outline-none"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-gray-500">Путь вкладки проекта</label>
                  <input
                    value={form.project_tab_url}
                    onChange={(e) => setForm({ ...form, project_tab_url: e.target.value })}
                    placeholder="/tab"
                    className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:border-violet-400 focus:outline-none"
                  />
                </div>
                <div className="sm:col-span-2">
                  <label className="text-xs font-medium text-gray-500">Название вкладки</label>
                  <input
                    value={form.project_tab_label}
                    onChange={(e) => setForm({ ...form, project_tab_label: e.target.value })}
                    placeholder="Демо-плагин"
                    className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:border-violet-400 focus:outline-none"
                  />
                </div>
              </div>

              <div>
                <label className="text-xs font-medium text-gray-500">Подписка на события</label>
                <div className="mt-1 flex flex-wrap gap-3">
                  {EXTENSION_EVENTS.map(({ value, label }) => (
                    <label key={value} className="flex items-center gap-1.5 text-xs text-gray-700">
                      <input
                        type="checkbox"
                        checked={form.events.includes(value)}
                        onChange={() => toggleEvent(value)}
                        className="h-3.5 w-3.5 rounded border-gray-300 text-violet-500 focus:ring-violet-400"
                      />
                      {label}
                    </label>
                  ))}
                </div>
              </div>

              <button
                type="submit"
                disabled={saving}
                className="inline-flex items-center gap-2 rounded-lg bg-violet-600 px-4 py-2 text-sm font-medium text-white hover:bg-violet-700 disabled:opacity-60 transition-colors"
              >
                {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
                Зарегистрировать
              </button>
            </form>
          )}
        </div>
      )}
    </div>
  );
}
