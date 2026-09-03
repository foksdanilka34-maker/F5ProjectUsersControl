import { useEffect, useState } from 'react';
import { Blocks } from 'lucide-react';
import { listProjectExtensions, type ProjectExtension } from '../services/extensionsService';

type Props = {
  taskId: number;
  projectId: number;
};

export default function TaskExtensionPanels({ taskId, projectId }: Props) {
  const [extensions, setExtensions] = useState<ProjectExtension[]>([]);

  useEffect(() => {
    let cancelled = false;
    listProjectExtensions(projectId)
      .then((all) => {
        if (!cancelled) setExtensions(all.filter((e) => e.enabled && e.task_panel_url));
      })
      .catch((err) => console.error('Failed to load task extension panels:', err));
    return () => {
      cancelled = true;
    };
  }, [projectId]);

  if (extensions.length === 0) {
    return null;
  }

  return (
    <div className="space-y-3">
      {extensions.map((ext) => (
        <div key={ext.key} className="rounded-xl border border-violet-100 bg-violet-50/50 overflow-hidden">
          <div className="flex items-center gap-2 px-4 py-2 border-b border-violet-100">
            <Blocks className="h-3.5 w-3.5 text-violet-500" />
            <span className="text-xs font-semibold text-violet-800">{ext.name}</span>
          </div>
          <iframe
            src={`${ext.base_url}${ext.task_panel_url}?task_id=${taskId}&project_id=${projectId}`}
            title={ext.name}
            className="w-full border-0"
            style={{ height: 220 }}
            sandbox="allow-scripts allow-same-origin allow-forms"
          />
        </div>
      ))}
    </div>
  );
}
