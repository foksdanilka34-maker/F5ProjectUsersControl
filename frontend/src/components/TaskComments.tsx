import { useCallback, useEffect, useState, type FormEvent } from 'react';
import { Loader2, MessageSquare, Send } from 'lucide-react';
import Avatar from './Avatar';
import { useWSEvent } from '../context/WebSocketContext';
import { addTaskComment, getTaskComments, type TaskComment } from '../services/projectService';
import type { ProfileDTO } from '../services/types';

type Props = {
  taskId: number;
  getProfileName: (userId: number) => string;
  getProfile: (userId: number) => ProfileDTO | undefined;
  onError: (message: string) => void;
};

export default function TaskComments({ taskId, getProfileName, getProfile, onError }: Props) {
  const [comments, setComments] = useState<TaskComment[]>([]);
  const [loading, setLoading] = useState(true);
  const [content, setContent] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const load = useCallback(async () => {
    try {
      setComments(await getTaskComments(taskId));
    } catch (err) {
      console.error('Failed to load comments:', err);
    } finally {
      setLoading(false);
    }
  }, [taskId]);

  useEffect(() => {
    load();
  }, [load]);

  const handleNewComment = useCallback(
    (payload: unknown) => {
      const comment = payload as TaskComment;
      if (comment?.task_id === taskId) {
        setComments((prev) => (prev.some((c) => c.id === comment.id) ? prev : [...prev, comment]));
      }
    },
    [taskId]
  );

  useWSEvent('task:comment_added', handleNewComment, [handleNewComment]);

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    const trimmed = content.trim();
    if (!trimmed) return;

    setSubmitting(true);
    try {
      await addTaskComment(taskId, trimmed);
      setContent('');
      await load();
    } catch (err) {
      console.error('Failed to add comment:', err);
      onError('Не удалось добавить комментарий');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div>
      <div className="mb-2 flex items-center gap-2">
        <MessageSquare className="h-4 w-4 text-gray-400" />
        <label className="text-xs font-medium text-gray-400 uppercase tracking-wide">Комментарии</label>
      </div>

      {loading ? (
        <div className="flex justify-center py-3">
          <Loader2 className="h-4 w-4 animate-spin text-gray-400" />
        </div>
      ) : (
        <div className="mb-3 space-y-2">
          {comments.length === 0 && <p className="text-sm text-gray-400">Комментариев пока нет</p>}
          {comments.map((c) => (
            <div key={c.id} className="flex items-start gap-2 rounded-lg bg-gray-50 px-3 py-2">
              <Avatar src={getProfile(c.user_id)?.avatar_url} name={getProfileName(c.user_id)} size="xs" />
              <div className="min-w-0 flex-1">
                <div className="flex items-baseline gap-2">
                  <span className="text-xs font-medium text-gray-900">{getProfileName(c.user_id)}</span>
                  <span className="text-[10px] text-gray-400">
                    {new Date(c.created_at).toLocaleString('ru', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })}
                  </span>
                </div>
                <p className="text-sm text-gray-700 whitespace-pre-wrap break-words">{c.content}</p>
              </div>
            </div>
          ))}
        </div>
      )}

      <form onSubmit={handleSubmit} className="flex items-center gap-2">
        <input
          type="text"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Написать комментарий..."
          className="flex-1 rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
        />
        <button
          type="submit"
          disabled={submitting || !content.trim()}
          className="inline-flex items-center justify-center rounded-xl bg-gray-900 p-2.5 text-white hover:bg-gray-800 disabled:opacity-50 transition-colors"
        >
          {submitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
        </button>
      </form>
    </div>
  );
}
