import { useCallback, useEffect, useState } from 'react';
import { Copy, ExternalLink, GitBranch, GitMerge, Hash, Loader2, Plus, RotateCw } from 'lucide-react';
import { useWSEvent } from '../context/WebSocketContext';
import {
  createTaskBranch,
  getTaskGit,
  pipelineStyle,
  retryPipeline,
  type TaskGitLink,
  type TaskGitOverview,
} from '../services/gitlabService';

type Props = {
  taskId: number;
  canManage: boolean;
  onError: (message: string) => void;
  onSuccess: (message: string) => void;
};

const MR_STATE_LABELS: Record<string, string> = {
  opened: 'Открыт',
  merged: 'Влит',
  closed: 'Закрыт',
  locked: 'Заблокирован',
  created: 'Создана',
};

const LINK_ICONS = {
  BRANCH: GitBranch,
  MERGE_REQUEST: GitMerge,
  COMMIT: Hash,
};

export default function TaskGitPanel({ taskId, canManage, onError, onSuccess }: Props) {
  const [overview, setOverview] = useState<TaskGitOverview | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    try {
      setOverview(await getTaskGit(taskId));
    } catch (err) {
      console.error('Failed to load task git info:', err);
    } finally {
      setLoading(false);
    }
  }, [taskId]);

  useEffect(() => {
    load();
  }, [load]);

  const refreshOnEvent = useCallback(
    (payload: unknown) => {
      const event = payload as { task_id?: number };
      if (event?.task_id === taskId) {
        load();
      }
    },
    [taskId, load]
  );

  useWSEvent('gitlab:pipeline', refreshOnEvent, [refreshOnEvent]);
  useWSEvent('gitlab:link', refreshOnEvent, [refreshOnEvent]);

  const handleCreateBranch = async () => {
    setBusy(true);
    try {
      await createTaskBranch(taskId);
      await load();
      onSuccess('Ветка создана в GitLab');
    } catch (err) {
      console.error('Failed to create branch:', err);
      onError('Не удалось создать ветку — проверьте настройки интеграции');
    } finally {
      setBusy(false);
    }
  };

  const handleRetry = async (pipelineId: number) => {
    setBusy(true);
    try {
      await retryPipeline(taskId, pipelineId);
      await load();
      onSuccess('Пайплайн перезапущен');
    } catch (err) {
      console.error('Failed to retry pipeline:', err);
      onError('Не удалось перезапустить пайплайн');
    } finally {
      setBusy(false);
    }
  };

  const copyBranchCommand = () => {
    if (!overview) return;
    navigator.clipboard
      .writeText(`git checkout -b ${overview.suggested_branch}`)
      .then(() => onSuccess('Команда скопирована'))
      .catch(() => onError('Не удалось скопировать команду'));
  };

  if (loading) {
    return (
      <div className="rounded-xl border border-gray-100 bg-gray-50 p-4 flex justify-center">
        <Loader2 className="h-4 w-4 animate-spin text-gray-400" />
      </div>
    );
  }

  if (!overview?.connected) {
    return null;
  }

  const links = overview.links || [];
  const pipelines = overview.pipelines || [];
  const branches = links.filter((l) => l.kind === 'BRANCH');
  const mergeRequests = links.filter((l) => l.kind === 'MERGE_REQUEST');
  const commits = links.filter((l) => l.kind === 'COMMIT');

  const renderLink = (link: TaskGitLink) => {
    const Icon = LINK_ICONS[link.kind];
    const label = link.kind === 'MERGE_REQUEST' ? `!${link.external_id}` : link.external_id;

    return (
      <div key={`${link.kind}-${link.id}`} className="flex items-center justify-between gap-2 rounded-lg bg-white px-3 py-2">
        <div className="flex min-w-0 items-center gap-2">
          <Icon className="h-4 w-4 shrink-0 text-gray-400" />
          <span className="truncate text-sm text-gray-700">{link.title || label}</span>
          <span className="shrink-0 font-mono text-xs text-gray-400">{label}</span>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          {link.state && (
            <span className="rounded-full bg-gray-100 px-2 py-0.5 text-xs text-gray-600">
              {MR_STATE_LABELS[link.state] || link.state}
            </span>
          )}
          {link.web_url && (
            <a href={link.web_url} target="_blank" rel="noreferrer" className="text-gray-400 hover:text-emerald-600">
              <ExternalLink className="h-3.5 w-3.5" />
            </a>
          )}
        </div>
      </div>
    );
  };

  return (
    <div className="rounded-xl border border-orange-100 bg-orange-50/60 p-4">
      <div className="mb-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <GitBranch className="h-4 w-4 text-orange-500" />
          <h4 className="text-sm font-semibold text-gray-800">GitLab</h4>
          <span className="rounded-full bg-white px-2 py-0.5 font-mono text-xs text-gray-500">{overview.task_key}</span>
        </div>
        {canManage && branches.length === 0 && (
          <button
            onClick={handleCreateBranch}
            disabled={busy}
            className="inline-flex items-center gap-1.5 rounded-lg bg-gray-900 px-3 py-1.5 text-xs font-medium text-white hover:bg-gray-800 disabled:opacity-60 transition-colors"
          >
            {busy ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
            Создать ветку
          </button>
        )}
      </div>

      <div className="mb-3 flex items-center gap-2 rounded-lg bg-white px-3 py-2">
        <code className="min-w-0 flex-1 truncate font-mono text-xs text-gray-600">
          git checkout -b {overview.suggested_branch}
        </code>
        <button onClick={copyBranchCommand} className="text-gray-400 hover:text-emerald-600">
          <Copy className="h-3.5 w-3.5" />
        </button>
      </div>

      {pipelines.length > 0 && (
        <div className="mb-3 space-y-2">
          {pipelines.slice(0, 3).map((pipeline) => {
            const style = pipelineStyle(pipeline.status);
            return (
              <div key={pipeline.id} className="flex items-center justify-between gap-2 rounded-lg bg-white px-3 py-2">
                <div className="flex min-w-0 items-center gap-2">
                  <span className={`h-2 w-2 shrink-0 rounded-full ${style.dot}`} />
                  <span className={`shrink-0 rounded-full px-2 py-0.5 text-xs font-medium ${style.class}`}>
                    {style.label}
                  </span>
                  <span className="truncate text-xs text-gray-500">
                    #{pipeline.pipeline_id} · {pipeline.ref}
                  </span>
                </div>
                <div className="flex shrink-0 items-center gap-2">
                  {canManage && pipeline.status === 'failed' && (
                    <button
                      onClick={() => handleRetry(pipeline.pipeline_id)}
                      disabled={busy}
                      className="inline-flex items-center gap-1 text-xs font-medium text-gray-600 hover:text-emerald-600 disabled:opacity-60"
                    >
                      <RotateCw className="h-3.5 w-3.5" />
                      Перезапустить
                    </button>
                  )}
                  {pipeline.web_url && (
                    <a href={pipeline.web_url} target="_blank" rel="noreferrer" className="text-gray-400 hover:text-emerald-600">
                      <ExternalLink className="h-3.5 w-3.5" />
                    </a>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {links.length > 0 ? (
        <div className="space-y-2">
          {mergeRequests.map(renderLink)}
          {branches.map(renderLink)}
          {commits.slice(0, 3).map(renderLink)}
        </div>
      ) : (
        <p className="text-xs text-gray-500">
          Веток и merge request'ов пока нет. Создайте ветку с ключом задачи — связь появится автоматически.
        </p>
      )}
    </div>
  );
}
