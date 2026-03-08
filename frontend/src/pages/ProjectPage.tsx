import { useCallback, useEffect, useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import {
  BarChart3,
  CheckCircle,
  CheckCircle2,
  Edit2,
  Eye,
  GripVertical,
  Loader2,
  Plus,
  Save,
  Settings,
  Trash2,
  TrendingUp,
  User,
  UserPlus,
  Users,
  X,
  Shield,
} from 'lucide-react';
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Tooltip,
} from 'recharts';
import Avatar from '../components/Avatar';

import { useAuth } from '../context/AuthContext';
import { getProjectMetrics, type ProjectMetrics } from '../services/analyticsService';
import {
  getProject,
  getProjectMembers,
  getTasks,
  createTask,
  updateTask,
  moveTask,
  deleteTask,
  addProjectMember,
  removeProjectMember,
  updateProject,
  getReviewStatus,
  approveTask,
  type Project,
  type Task,
  type ProjectMember,
  type ReviewStatus,
} from '../services/projectService';
import { listProfiles } from '../services/employeeService';
import type { ProfileDTO } from '../services/types';

type KanbanColumn = {
  id: string;
  title: string;
  color: string;
  bgColor: string;
};

const KANBAN_COLUMNS: KanbanColumn[] = [
  { id: 'TODO', title: 'К выполнению', color: 'text-gray-500', bgColor: 'bg-gray-100' },
  { id: 'IN_PROGRESS', title: 'В работе', color: 'text-blue-600', bgColor: 'bg-blue-100' },
  { id: 'REVIEW', title: 'На проверке', color: 'text-amber-600', bgColor: 'bg-amber-100' },
  { id: 'DONE', title: 'Готово', color: 'text-emerald-600', bgColor: 'bg-emerald-100' },
];

const PRIORITY_STYLES: Record<string, { label: string; class: string; value: number }> = {
  TASK_PRIORITY_LOW: { label: 'Низкий', class: 'bg-gray-100 text-gray-600', value: 1 },
  TASK_PRIORITY_MEDIUM: { label: 'Средний', class: 'bg-blue-100 text-blue-600', value: 2 },
  TASK_PRIORITY_HIGH: { label: 'Высокий', class: 'bg-amber-100 text-amber-600', value: 3 },
  TASK_PRIORITY_CRITICAL: { label: 'Критический', class: 'bg-red-100 text-red-600', value: 4 },
};

const PROJECT_STATUSES = [
  { id: 'ACTIVE', label: 'Активен', color: 'bg-emerald-100 text-emerald-700' },
  { id: 'ON_HOLD', label: 'Приостановлен', color: 'bg-amber-100 text-amber-700' },
  { id: 'ARCHIVED', label: 'В архиве', color: 'bg-gray-100 text-gray-600' },
];

const STATUS_LABELS: Record<string, string> = {
  PROJECT_STATUS_UNSPECIFIED: 'Не указан',
  ACTIVE: 'Активен',
  ON_HOLD: 'Приостановлен',
  ARCHIVED: 'В архиве',
};

export default function ProjectPage() {
  const { id } = useParams<{ id: string }>();
  const projectId = Number(id);
  const { user } = useAuth();

  const canManage = user && ['manager', 'developer', 'director', 'admin'].includes(user.role);
  const isManagerOrHigher = user && ['manager', 'developer', 'director', 'admin'].includes(user.role);

  const [project, setProject] = useState<Project | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [members, setMembers] = useState<ProjectMember[]>([]);
  const [allProfiles, setAllProfiles] = useState<ProfileDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [metrics, setMetrics] = useState<ProjectMetrics | null>(null);

  const [activeTab, setActiveTab] = useState<'kanban' | 'team' | 'stats' | 'settings'>('kanban');
  const [showTaskForm, setShowTaskForm] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [showMemberModal, setShowMemberModal] = useState(false);
  const [draggingTask, setDraggingTask] = useState<Task | null>(null);
  const [dragOverColumn, setDragOverColumn] = useState<string | null>(null);

  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [reviewStatus, setReviewStatus] = useState<ReviewStatus | null>(null);
  const [reviewLoading, setReviewLoading] = useState(false);

  const [taskTitle, setTaskTitle] = useState('');
  const [taskDescription, setTaskDescription] = useState('');
  const [taskPriority, setTaskPriority] = useState('TASK_PRIORITY_MEDIUM');
  const [taskAssignee, setTaskAssignee] = useState('');
  const [taskDueDate, setTaskDueDate] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const [projectName, setProjectName] = useState('');
  const [projectDescription, setProjectDescription] = useState('');
  const [projectManager, setProjectManager] = useState('');
  const [projectStatus, setProjectStatus] = useState('');
  const [projectDueDate, setProjectDueDate] = useState('');

  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      const [projectData, tasksData, membersData, profilesData, metricsData] = await Promise.all([
        getProject(projectId),
        getTasks(projectId),
        getProjectMembers(projectId),
        listProfiles({ pageSize: 100 }),
        getProjectMetrics(projectId).catch(() => null),
      ]);
      setProject(projectData);
      setTasks(tasksData || []);
      setMembers(membersData || []);
      setAllProfiles(profilesData.profiles || []);
      setMetrics(metricsData);

      setProjectName(projectData.name);
      setProjectDescription(projectData.description || '');
      setProjectManager(projectData.manager_id?.toString() || '');
      setProjectStatus(projectData.status || 'ACTIVE');
      setProjectDueDate(projectData.due_date ? projectData.due_date.split('T')[0] : '');
    } catch (err) {
      console.error('Failed to load project:', err);
      setError('Не удалось загрузить проект');
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const tasksByStatus = useMemo(() => {
    const grouped: Record<string, Task[]> = {};
    KANBAN_COLUMNS.forEach((col) => (grouped[col.id] = []));
    tasks.forEach((task) => {
      if (grouped[task.status]) {
        grouped[task.status].push(task);
      } else {
        grouped['TODO'].push(task);
      }
    });
    return grouped;
  }, [tasks]);

  const memberProfiles = useMemo(() => {
    const memberUserIds = new Set(members.map((m) => m.user_id));
    return allProfiles.filter((p) => memberUserIds.has(p.id));
  }, [members, allProfiles]);

  const assignableProfiles = useMemo(() => {
    const memberUserIds = new Set(members.map((m) => m.user_id));
    return allProfiles.filter((p) => memberUserIds.has(p.id) && p.role === 'employee');
  }, [members, allProfiles]);

  const availableProfiles = useMemo(() => {
    const memberUserIds = new Set(members.map((m) => m.user_id));
    return allProfiles.filter((p) => !memberUserIds.has(p.id));
  }, [members, allProfiles]);

  const getProfileName = (userId: number) => {
    const profile = allProfiles.find((p) => p.id === userId);
    return profile ? `${profile.first_name} ${profile.last_name}` : '—';
  };

  const getProfile = (userId: number) => {
    return allProfiles.find((p) => p.id === userId);
  };

  const handleCreateTask = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!taskTitle.trim()) return;
    setSubmitting(true);
    try {
      if (editingTask) {

        await updateTask(editingTask.id, {
          title: taskTitle.trim(),
          description: taskDescription.trim() || undefined,
          priority: PRIORITY_STYLES[taskPriority]?.value,
          assignee_id: taskAssignee ? Number(taskAssignee) : undefined,
          due_date: taskDueDate || undefined,
        });
        setSuccess('Задача обновлена');
      } else {

        await createTask({
          project_id: projectId,
          title: taskTitle.trim(),
          description: taskDescription.trim() || undefined,
          priority: PRIORITY_STYLES[taskPriority]?.value,
          assignee_id: taskAssignee ? Number(taskAssignee) : undefined,
          due_date: taskDueDate || undefined,
        });
        setSuccess('Задача создана');
      }
      resetTaskForm();
      loadData();
    } catch (err) {
      console.error('Task error:', err);
      setError(editingTask ? 'Не удалось обновить задачу' : 'Не удалось создать задачу');
    } finally {
      setSubmitting(false);
    }
  };

  const resetTaskForm = () => {
    setTaskTitle('');
    setTaskDescription('');
    setTaskPriority('TASK_PRIORITY_MEDIUM');
    setTaskAssignee('');
    setTaskDueDate('');
    setShowTaskForm(false);
    setEditingTask(null);
  };

  const openEditTask = (task: Task) => {
    setEditingTask(task);
    setTaskTitle(task.title);
    setTaskDescription(task.description || '');
    setTaskPriority(task.priority || 'TASK_PRIORITY_MEDIUM');
    setTaskAssignee(task.assignee_id?.toString() || '');
    setTaskDueDate(task.due_date ? task.due_date.split('T')[0] : '');
    setShowTaskForm(true);
  };

  const handleDeleteTask = async (taskId: number) => {
    if (!confirm('Удалить задачу?')) return;
    try {
      await deleteTask(taskId);
      setSuccess('Задача удалена');
      loadData();
    } catch {
      setError('Не удалось удалить задачу');
    }
  };

  const openTaskDetail = async (task: Task) => {
    setSelectedTask(task);
    setReviewStatus(null);
    if (task.status === 'REVIEW') {
      setReviewLoading(true);
      try {
        const status = await getReviewStatus(task.id);
        setReviewStatus(status);
      } catch {
        // review status not available
      } finally {
        setReviewLoading(false);
      }
    }
  };

  const handleApproveTask = async () => {
    if (!selectedTask) return;
    setReviewLoading(true);
    try {
      await approveTask(selectedTask.id);
      setSuccess('Задача одобрена');
      const status = await getReviewStatus(selectedTask.id);
      setReviewStatus(status);
      loadData();
    } catch {
      setError('Не удалось одобрить задачу');
    } finally {
      setReviewLoading(false);
    }
  };


  const canDragTask = (task: Task): boolean => {
    if (!user) return false;
    if (task.status === 'REVIEW') return false;

    if (user.role === 'developer') return true;

    if (task.assignee_id === user.id) return true;
    return false;
  };

  const handleDragStart = (e: React.DragEvent, task: Task) => {
    if (!canDragTask(task)) {
      e.preventDefault();
      return;
    }
    setDraggingTask(task);
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', task.id.toString());
  };

  const handleDragOver = (e: React.DragEvent, columnId: string) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    setDragOverColumn(columnId);
  };

  const handleDragLeave = () => {
    setDragOverColumn(null);
  };

  const handleDrop = async (e: React.DragEvent, columnId: string) => {
    e.preventDefault();
    setDragOverColumn(null);

    if (!draggingTask || draggingTask.status === columnId) {
      setDraggingTask(null);
      return;
    }

    try {
      await moveTask(draggingTask.id, columnId);

      setTasks((prev) => prev.map((t) => (t.id === draggingTask.id ? { ...t, status: columnId } : t)));
    } catch {
      setError('Не удалось переместить задачу');
    } finally {
      setDraggingTask(null);
    }
  };

  const handleAddMember = async (userId: number) => {
    try {
      await addProjectMember(projectId, userId, 'member');
      setSuccess('Участник добавлен');
      setShowMemberModal(false);
      loadData();
    } catch {
      setError('Не удалось добавить участника');
    }
  };

  const handleRemoveMember = async (userId: number) => {
    if (!confirm('Удалить участника?')) return;
    try {
      await removeProjectMember(projectId, userId);
      setSuccess('Участник удалён');
      loadData();
    } catch {
      setError('Не удалось удалить участника');
    }
  };

  const handleSaveSettings = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!isManagerOrHigher) {
      setError('Только менеджеры могут изменять настройки проекта');
      return;
    }
    setSubmitting(true);
    try {
      await updateProject(projectId, {
        name: projectName.trim(),
        description: projectDescription.trim() || undefined,
        manager_id: projectManager ? Number(projectManager) : undefined,
        status: projectStatus,
        due_date: projectDueDate || undefined,
      });
      setSuccess('Настройки сохранены');
      loadData();
    } catch {
      setError('Не удалось сохранить настройки');
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-white via-emerald-50/30 to-white">
        <Loader2 className="h-8 w-8 animate-spin text-emerald-500" />
      </div>
    );
  }

  if (!project) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-white via-emerald-50/30 to-white">
        <p className="text-gray-500 mb-4">Проект не найден</p>
        <Link to="/projects" className="text-emerald-600 hover:underline">
          Вернуться на главную
        </Link>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-white via-emerald-50/30 to-white text-gray-900">
      <div className="mx-auto max-w-7xl px-4 py-6">
        {}
        <header className="flex flex-wrap items-start justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">{project.name}</h1>
            <div className="flex items-center gap-3 mt-1">
              <span className="inline-flex items-center gap-1 text-sm text-gray-500">
                <User className="h-3.5 w-3.5" />
                {getProfileName(project.manager_id)}
              </span>
              <span
                className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${
                  PROJECT_STATUSES.find((s) => s.id === project.status)?.color || 'bg-gray-100 text-gray-600'
                }`}
              >
                {STATUS_LABELS[project.status] || project.status}
              </span>
            </div>
          </div>
        </header>

        {}
        {success && (
          <div className="mb-4 rounded-xl border border-emerald-100 bg-emerald-50 px-4 py-2 text-sm text-emerald-700 flex justify-between">
            {success}
            <button onClick={() => setSuccess(null)}>
              <X className="h-4 w-4" />
            </button>
          </div>
        )}
        {error && (
          <div className="mb-4 rounded-xl border border-red-100 bg-red-50 px-4 py-2 text-sm text-red-700 flex justify-between">
            {error}
            <button onClick={() => setError(null)}>
              <X className="h-4 w-4" />
            </button>
          </div>
        )}

        {}
        <div className="flex flex-wrap gap-2 mb-6">
          {[
            { key: 'kanban', label: 'Задачи', icon: CheckCircle2 },
            { key: 'team', label: 'Команда', icon: Users },
            { key: 'stats', label: 'Статистика', icon: BarChart3 },
            { key: 'settings', label: 'Настройки', icon: Settings },
          ].map(({ key, label, icon: Icon }) => (
            <button
              key={key}
              type="button"
              onClick={() => setActiveTab(key as typeof activeTab)}
              className={`inline-flex items-center gap-2 rounded-full px-4 py-2 text-sm font-medium transition-all ${
                activeTab === key
                  ? 'bg-gray-900 text-white'
                  : 'border border-gray-200 bg-white text-gray-600 hover:text-emerald-600'
              }`}
            >
              <Icon className="h-4 w-4" />
              {label}
            </button>
          ))}
        </div>

        {}
        {activeTab === 'kanban' && (
          <div>
            <div className="flex items-center justify-between mb-4">
              <h2 className="font-semibold text-gray-900">Доска задач</h2>
              {isManagerOrHigher && (
                <button
                  type="button"
                  onClick={() => setShowTaskForm(true)}
                  className="inline-flex items-center gap-2 rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600"
                >
                  <Plus className="h-4 w-4" />
                  Новая задача
                </button>
              )}
            </div>

            <div className="flex gap-4 overflow-x-auto pb-4">
              {KANBAN_COLUMNS.map((column) => (
                <div
                  key={column.id}
                  onDragOver={(e) => handleDragOver(e, column.id)}
                  onDragLeave={handleDragLeave}
                  onDrop={(e) => handleDrop(e, column.id)}
                  className={`flex-shrink-0 w-72 rounded-xl border p-3 transition-colors flex flex-col ${
                    dragOverColumn === column.id
                      ? 'border-emerald-400 bg-emerald-50/50'
                      : 'border-gray-200 bg-gray-50/50'
                  }`}
                >
                  <div className="flex items-center justify-between mb-3 flex-shrink-0">
                    <div className="flex items-center gap-2">
                      <span className={`h-2.5 w-2.5 rounded-full ${column.bgColor}`} />
                      <h3 className={`text-sm font-semibold ${column.color}`}>{column.title}</h3>
                    </div>
                    <span className="rounded-full bg-gray-200 px-2 py-0.5 text-xs font-medium text-gray-600">
                      {tasksByStatus[column.id]?.length || 0}
                    </span>
                  </div>

                  <div className="flex-1 overflow-y-auto space-y-2 max-h-[calc(100vh-320px)] min-h-[200px] pr-1">
                    {tasksByStatus[column.id]?.map((task) => {
                      const isDraggable = canDragTask(task);
                      const isOverdue = task.due_date && new Date(task.due_date) < new Date() && task.status !== 'DONE';
                      return (
                      <div
                        key={task.id}
                        draggable={isDraggable}
                        onDragStart={(e) => handleDragStart(e, task)}
                        onClick={() => openTaskDetail(task)}
                        className={`rounded-lg border bg-white p-2.5 shadow-sm transition-all group ${
                          isDraggable ? 'cursor-grab active:cursor-grabbing hover:shadow-md hover:border-emerald-200' : 'cursor-pointer'
                        } ${draggingTask?.id === task.id ? 'opacity-50 scale-95' : ''} ${
                          isOverdue ? 'border-l-2 border-l-red-400' : 'border-gray-100'
                        }`}
                      >
                        <div className="flex items-start justify-between gap-1 mb-1.5">
                          <h4 className="flex-1 text-sm font-medium text-gray-900 line-clamp-2 leading-tight">{task.title}</h4>
                          <div className="flex gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                openEditTask(task);
                              }}
                              className="p-1 text-gray-400 hover:text-emerald-500 rounded"
                            >
                              <Edit2 className="h-3 w-3" />
                            </button>
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                handleDeleteTask(task.id);
                              }}
                              className="p-1 text-gray-400 hover:text-red-500 rounded"
                            >
                              <Trash2 className="h-3 w-3" />
                            </button>
                          </div>
                        </div>
                        
                        <div className="flex items-center justify-between gap-2">
                          <div className="flex items-center gap-1.5">
                            <span
                              className={`inline-flex rounded px-1.5 py-0.5 text-[10px] font-medium ${
                                PRIORITY_STYLES[task.priority]?.class || 'bg-gray-100 text-gray-600'
                              }`}
                            >
                              {PRIORITY_STYLES[task.priority]?.label || '—'}
                            </span>
                            {task.status === 'REVIEW' && (
                              <span className="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-[10px] font-medium bg-amber-100 text-amber-700">
                                <Eye className="h-2.5 w-2.5" />
                                Ревью
                              </span>
                            )}
                          </div>
                          
                          <div className="flex items-center gap-1.5">
                            {task.due_date && (
                              <span className={`text-[10px] ${isOverdue ? 'text-red-500 font-medium' : 'text-gray-400'}`}>
                                {new Date(task.due_date).toLocaleDateString('ru', { day: 'numeric', month: 'short' })}
                              </span>
                            )}
                            {task.assignee_id && (() => {
                              const assigneeProfile = getProfile(task.assignee_id);
                              return (
                                <Avatar
                                  src={assigneeProfile?.avatar_url}
                                  name={getProfileName(task.assignee_id)}
                                  size="xs"
                                />
                              );
                            })()}
                          </div>
                        </div>
                      </div>
                    );
                    })}
                    
                    {}
                    {(!tasksByStatus[column.id] || tasksByStatus[column.id].length === 0) && (
                      <div className="flex items-center justify-center h-20 text-xs text-gray-400 border-2 border-dashed border-gray-200 rounded-lg">
                        Перетащите задачу сюда
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {}
        {activeTab === 'team' && (
          <div className="grid gap-6 lg:grid-cols-3">
            <div className="lg:col-span-2">
              <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="font-semibold text-gray-900">Участники проекта</h3>
                  <button
                    type="button"
                    onClick={() => setShowMemberModal(true)}
                    className="inline-flex items-center gap-2 rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600"
                  >
                    <UserPlus className="h-4 w-4" />
                    Добавить
                  </button>
                </div>

                {memberProfiles.length ? (
                  <div className="space-y-2">
                    {memberProfiles.map((profile) => {
                      const isManager = project.manager_id === profile.id;
                      return (
                        <div
                          key={profile.id}
                          className="flex items-center justify-between rounded-xl bg-gray-50 px-4 py-3"
                        >
                          <div className="flex items-center gap-3">
                            <Avatar
                              src={profile.avatar_url}
                              name={`${profile.first_name} ${profile.last_name}`}
                              size="md"
                            />
                            <div>
                              <p className="font-medium text-gray-900">
                                {profile.first_name} {profile.last_name}
                              </p>
                              <p className="text-xs text-gray-400">{profile.email}</p>
                            </div>
                          </div>
                          <div className="flex items-center gap-2">
                            {isManager && (
                              <span className="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-700">
                                Менеджер
                              </span>
                            )}
                            {!isManager && user && ['manager', 'developer'].includes(user.role) && (
                              <button
                                onClick={() => handleRemoveMember(profile.id)}
                                className="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg"
                              >
                                <Trash2 className="h-4 w-4" />
                              </button>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                ) : (
                  <p className="text-center text-sm text-gray-500 py-8">В проекте пока нет участников</p>
                )}
              </div>
            </div>

            {}
            <div>
              <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm">
                <h4 className="text-xs font-semibold uppercase tracking-wide text-gray-400 mb-3">Статистика</h4>
                <div className="space-y-3">
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Участников</span>
                    <span className="font-semibold text-gray-900">{memberProfiles.length}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Всего задач</span>
                    <span className="font-semibold text-gray-900">{tasks.length}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">Завершено</span>
                    <span className="font-semibold text-emerald-600">
                      {tasksByStatus['DONE']?.length || 0}
                    </span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-500">В работе</span>
                    <span className="font-semibold text-blue-600">
                      {tasksByStatus['IN_PROGRESS']?.length || 0}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {}
        {activeTab === 'stats' && (
          <div className="grid gap-6 lg:grid-cols-2">
            {}
            <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
              <div className="flex items-center gap-2 mb-4">
                <TrendingUp className="h-5 w-5 text-emerald-500" />
                <h3 className="font-semibold text-gray-900">Прогресс проекта</h3>
              </div>
              
              {metrics ? (
                <div className="space-y-4">
                  <div>
                    <div className="flex justify-between text-sm mb-1">
                      <span className="text-gray-500">Выполнение задач</span>
                      <span className="font-medium text-emerald-600">
                        {metrics.progress_percent}%
                      </span>
                    </div>
                    <div className="h-3 rounded-full bg-gray-100">
                      <div
                        className="h-3 rounded-full bg-gradient-to-r from-emerald-400 to-lime-400 transition-all duration-500"
                        style={{ width: `${Math.min(100, metrics.progress_percent || 0)}%` }}
                      />
                    </div>
                  </div>
                  
                  <div>
                    <div className="flex justify-between text-sm mb-1">
                      <span className="text-gray-500">В срок</span>
                      <span className="font-medium text-blue-600">
                        {Math.round(metrics.on_time_rate || 0)}%
                      </span>
                    </div>
                    <div className="h-3 rounded-full bg-gray-100">
                      <div
                        className="h-3 rounded-full bg-gradient-to-r from-blue-400 to-cyan-400 transition-all duration-500"
                        style={{ width: `${Math.min(100, Math.max(0, metrics.on_time_rate || 0))}%` }}
                      />
                    </div>
                  </div>
                </div>
              ) : (
                <p className="text-gray-500 text-sm">Загрузка статистики...</p>
              )}
            </div>

            {}
            <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
              <div className="flex items-center gap-2 mb-4">
                <BarChart3 className="h-5 w-5 text-blue-500" />
                <h3 className="font-semibold text-gray-900">Распределение задач</h3>
              </div>
              
              {metrics ? (
                (() => {
                  const donutData = [
                    { name: 'В срок', value: metrics.completed_on_time || 0, color: '#10b981' },
                    { name: 'С опозданием', value: metrics.completed_late || 0, color: '#f59e0b' },
                    { name: 'В работе', value: metrics.in_progress_tasks || 0, color: '#3b82f6' },
                    { name: 'Просрочено', value: metrics.overdue_tasks || 0, color: '#ef4444' },
                  ].filter((d) => d.value > 0);
                  const remaining = metrics.total_tasks - (metrics.completed_tasks + metrics.in_progress_tasks + metrics.overdue_tasks);
                  if (remaining > 0) {
                    donutData.push({ name: 'Открыто', value: remaining, color: '#9ca3af' });
                  }
                  return donutData.length > 0 ? (
                    <div>
                      <div className="relative" style={{ height: 180 }}>
                        <ResponsiveContainer width="100%" height="100%">
                          <PieChart>
                            <Pie
                              data={donutData}
                              cx="50%"
                              cy="50%"
                              innerRadius={50}
                              outerRadius={75}
                              paddingAngle={3}
                              dataKey="value"
                              stroke="none"
                            >
                              {donutData.map((entry, i) => (
                                <Cell key={i} fill={entry.color} />
                              ))}
                            </Pie>
                            <Tooltip
                              formatter={(value: number, name: string) => [`${value} задач`, name]}
                              contentStyle={{ borderRadius: '12px', border: '1px solid #e5e7eb', fontSize: '12px' }}
                            />
                          </PieChart>
                        </ResponsiveContainer>
                        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                          <div className="text-center">
                            <p className="text-2xl font-bold text-gray-900">{metrics.total_tasks}</p>
                            <p className="text-xs text-gray-400">задач</p>
                          </div>
                        </div>
                      </div>
                      <div className="grid grid-cols-2 gap-x-4 gap-y-2 mt-3">
                        {donutData.map((d) => (
                          <div key={d.name} className="flex items-center gap-2 text-sm text-gray-600">
                            <span className="h-2.5 w-2.5 rounded-full flex-shrink-0" style={{ backgroundColor: d.color }} />
                            <span className="truncate">{d.name}</span>
                            <span className="ml-auto font-semibold">{d.value}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  ) : (
                    <p className="text-gray-500 text-sm text-center py-8">Нет задач</p>
                  );
                })()
              ) : (
                <p className="text-gray-500 text-sm">Загрузка статистики...</p>
              )}
            </div>

            {}
            <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
              <div className="flex items-center gap-2 mb-4">
                <Users className="h-5 w-5 text-violet-500" />
                <h3 className="font-semibold text-gray-900">Команда</h3>
              </div>
              
              {metrics ? (
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-500">Размер команды</span>
                    <span className="font-bold text-gray-900">{metrics.team_size} чел.</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-500">Задач на человека</span>
                    <span className="font-bold text-gray-900">
                      {metrics.team_size > 0 ? Math.round(metrics.total_tasks / metrics.team_size * 10) / 10 : 0}
                    </span>
                  </div>
                </div>
              ) : (
                <p className="text-gray-500 text-sm">Загрузка статистики...</p>
              )}
            </div>

            {}
            <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
              <div className="flex items-center gap-2 mb-4">
                <CheckCircle2 className="h-5 w-5 text-amber-500" />
                <h3 className="font-semibold text-gray-900">Здоровье проекта</h3>
              </div>
              
              {metrics ? (
                (() => {
                  const isHealthy = metrics.health_status === 'HEALTH_STATUS_HEALTHY';
                  const isAtRisk = metrics.health_status === 'HEALTH_STATUS_AT_RISK';
                  const statusColor = isHealthy ? '#10b981' : isAtRisk ? '#f59e0b' : '#ef4444';
                  const statusBg = isHealthy ? 'bg-emerald-50' : isAtRisk ? 'bg-amber-50' : 'bg-red-50';
                  const statusText = isHealthy ? 'text-emerald-700' : isAtRisk ? 'text-amber-700' : 'text-red-700';
                  const statusLabel = isHealthy ? 'Отлично' : isAtRisk ? 'Требует внимания' : 'Критично';
                  const statusDesc = isHealthy
                    ? 'Проект идёт по плану'
                    : isAtRisk
                    ? 'Есть риски отставания'
                    : 'Много просроченных задач';
                  return (
                    <div className="space-y-4">
                      {/* Traffic light */}
                      <div className="flex items-center gap-4">
                        <div className={`rounded-2xl ${statusBg} p-3 flex items-center gap-2`}>
                          <div className="flex flex-col gap-1.5">
                            <span className={`h-4 w-4 rounded-full border-2 ${isHealthy ? 'bg-emerald-500 border-emerald-400' : 'bg-gray-200 border-gray-200'}`} />
                            <span className={`h-4 w-4 rounded-full border-2 ${isAtRisk ? 'bg-amber-500 border-amber-400' : 'bg-gray-200 border-gray-200'}`} />
                            <span className={`h-4 w-4 rounded-full border-2 ${!isHealthy && !isAtRisk ? 'bg-red-500 border-red-400' : 'bg-gray-200 border-gray-200'}`} />
                          </div>
                        </div>
                        <div>
                          <p className={`font-semibold ${statusText}`}>{statusLabel}</p>
                          <p className="text-sm text-gray-500">{statusDesc}</p>
                        </div>
                      </div>

                      {/* On-time rate mini bar */}
                      <div>
                        <div className="flex justify-between text-xs mb-1">
                          <span className="text-gray-500">Выполнение в срок</span>
                          <span className="font-medium" style={{ color: statusColor }}>
                            {Math.round(metrics.on_time_rate || 0)}%
                          </span>
                        </div>
                        <div className="h-2.5 rounded-full bg-gray-100">
                          <div
                            className="h-2.5 rounded-full transition-all duration-500"
                            style={{
                              width: `${Math.min(100, Math.max(0, metrics.on_time_rate || 0))}%`,
                              backgroundColor: statusColor,
                            }}
                          />
                        </div>
                      </div>

                      {/* Quick metric counters */}
                      <div className="grid grid-cols-3 gap-2 text-center">
                        <div className="rounded-xl bg-emerald-50 py-2">
                          <p className="text-lg font-bold text-emerald-700">{metrics.completed_tasks}</p>
                          <p className="text-[10px] text-emerald-600">Готово</p>
                        </div>
                        <div className="rounded-xl bg-blue-50 py-2">
                          <p className="text-lg font-bold text-blue-700">{metrics.in_progress_tasks}</p>
                          <p className="text-[10px] text-blue-600">В работе</p>
                        </div>
                        <div className="rounded-xl bg-red-50 py-2">
                          <p className="text-lg font-bold text-red-700">{metrics.overdue_tasks}</p>
                          <p className="text-[10px] text-red-600">Просрочено</p>
                        </div>
                      </div>
                    </div>
                  );
                })()
              ) : (
                <p className="text-gray-500 text-sm">Загрузка статистики...</p>
              )}
            </div>
          </div>
        )}

        {}
        {activeTab === 'settings' && (
          <div className="max-w-2xl">
            {!isManagerOrHigher && (
              <div className="mb-4 rounded-xl border border-amber-100 bg-amber-50 px-4 py-3 flex items-center gap-2">
                <Shield className="h-5 w-5 text-amber-600" />
                <span className="text-sm text-amber-700">
                  Только менеджеры могут изменять настройки проекта
                </span>
              </div>
            )}

            <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
              <h3 className="font-semibold text-gray-900 mb-4">Настройки проекта</h3>

              <form onSubmit={handleSaveSettings} className="space-y-4">
                <div>
                  <label className="text-xs font-medium text-gray-500">Название проекта</label>
                  <input
                    type="text"
                    value={projectName}
                    onChange={(e) => setProjectName(e.target.value)}
                    required
                    disabled={!isManagerOrHigher}
                    className="mt-1 w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none disabled:bg-gray-50 disabled:text-gray-500"
                  />
                </div>

                <div>
                  <label className="text-xs font-medium text-gray-500">Описание</label>
                  <textarea
                    value={projectDescription}
                    onChange={(e) => setProjectDescription(e.target.value)}
                    rows={3}
                    disabled={!isManagerOrHigher}
                    className="mt-1 w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none resize-none disabled:bg-gray-50 disabled:text-gray-500"
                  />
                </div>

                <div className="grid gap-4 sm:grid-cols-2">
                  <div>
                    <label className="text-xs font-medium text-gray-500">Статус проекта</label>
                    <select
                      value={projectStatus}
                      onChange={(e) => setProjectStatus(e.target.value)}
                      disabled={!isManagerOrHigher}
                      className="mt-1 w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none disabled:bg-gray-50 disabled:text-gray-500"
                    >
                      {PROJECT_STATUSES.map((s) => (
                        <option key={s.id} value={s.id}>
                          {s.label}
                        </option>
                      ))}
                    </select>
                  </div>

                  <div>
                    <label className="text-xs font-medium text-gray-500">Менеджер проекта</label>
                    <select
                      value={projectManager}
                      onChange={(e) => setProjectManager(e.target.value)}
                      disabled={!isManagerOrHigher}
                      className="mt-1 w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none disabled:bg-gray-50 disabled:text-gray-500"
                    >
                      <option value="">Не назначен</option>
                      {allProfiles
                        .filter((p) => p.role === 'manager' || p.role === 'director' || p.role === 'admin')
                        .map((p) => (
                          <option key={p.id} value={p.id}>
                            {p.first_name} {p.last_name}
                          </option>
                        ))}
                    </select>
                  </div>
                </div>

                <div>
                  <label className="text-xs font-medium text-gray-500">Дедлайн проекта</label>
                  <input
                    type="date"
                    value={projectDueDate}
                    onChange={(e) => setProjectDueDate(e.target.value)}
                    disabled={!isManagerOrHigher}
                    className="mt-1 w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none disabled:bg-gray-50 disabled:text-gray-500"
                  />
                </div>

                {isManagerOrHigher && (
                  <div className="pt-4">
                    <button
                      type="submit"
                      disabled={submitting}
                      className="inline-flex items-center gap-2 rounded-xl bg-emerald-500 px-6 py-2.5 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-60"
                    >
                      {submitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
                      Сохранить
                    </button>
                  </div>
                )}
              </form>
            </div>
          </div>
        )}
      </div>

      {}
      {showTaskForm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl shadow-xl w-full max-w-md">
            <div className="flex items-center justify-between p-4 border-b border-gray-100">
              <h3 className="font-semibold text-gray-900">
                {editingTask ? 'Редактировать задачу' : 'Новая задача'}
              </h3>
              <button onClick={resetTaskForm} className="p-1 text-gray-400 hover:text-gray-600">
                <X className="h-5 w-5" />
              </button>
            </div>

            <form onSubmit={handleCreateTask} className="p-4 space-y-4">
              <div>
                <label className="text-xs font-medium text-gray-500">Название *</label>
                <input
                  type="text"
                  value={taskTitle}
                  onChange={(e) => setTaskTitle(e.target.value)}
                  required
                  placeholder="Что нужно сделать?"
                  className="mt-1 w-full rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                />
              </div>

              <div>
                <label className="text-xs font-medium text-gray-500">Описание</label>
                <textarea
                  value={taskDescription}
                  onChange={(e) => setTaskDescription(e.target.value)}
                  rows={3}
                  placeholder="Детали задачи..."
                  className="mt-1 w-full rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none resize-none"
                />
              </div>

              <div className="grid gap-4 sm:grid-cols-2">
                <div>
                  <label className="text-xs font-medium text-gray-500">Приоритет</label>
                  <select
                    value={taskPriority}
                    onChange={(e) => setTaskPriority(e.target.value)}
                    className="mt-1 w-full rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                  >
                    <option value="TASK_PRIORITY_LOW">Низкий</option>
                    <option value="TASK_PRIORITY_MEDIUM">Средний</option>
                    <option value="TASK_PRIORITY_HIGH">Высокий</option>
                    <option value="TASK_PRIORITY_CRITICAL">Критический</option>
                  </select>
                </div>

                <div>
                  <label className="text-xs font-medium text-gray-500">Исполнитель</label>
                  <select
                    value={taskAssignee}
                    onChange={(e) => setTaskAssignee(e.target.value)}
                    className="mt-1 w-full rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                  >
                    <option value="">Не назначен</option>
                    {assignableProfiles.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.first_name} {p.last_name}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div>
                <label className="text-xs font-medium text-gray-500">Дедлайн</label>
                <input
                  type="date"
                  value={taskDueDate}
                  onChange={(e) => setTaskDueDate(e.target.value)}
                  className="mt-1 w-full rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={resetTaskForm}
                  className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800"
                >
                  Отмена
                </button>
                <button
                  type="submit"
                  disabled={submitting || !taskTitle.trim()}
                  className="inline-flex items-center gap-2 rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-60"
                >
                  {submitting ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : editingTask ? (
                    <Save className="h-4 w-4" />
                  ) : (
                    <Plus className="h-4 w-4" />
                  )}
                  {editingTask ? 'Сохранить' : 'Создать'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {}
      {showMemberModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl shadow-xl w-full max-w-md">
            <div className="flex items-center justify-between p-4 border-b border-gray-100">
              <h3 className="font-semibold text-gray-900">Добавить участника</h3>
              <button onClick={() => setShowMemberModal(false)} className="p-1 text-gray-400 hover:text-gray-600">
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="p-4 max-h-[400px] overflow-y-auto">
              {availableProfiles.length ? (
                <div className="space-y-2">
                  {availableProfiles.map((profile) => (
                    <button
                      key={profile.id}
                      type="button"
                      onClick={() => handleAddMember(profile.id)}
                      className="w-full flex items-center gap-3 rounded-xl bg-gray-50 px-4 py-3 hover:bg-emerald-50 transition-colors text-left"
                    >
                      <div className="h-10 w-10 rounded-full bg-gradient-to-r from-emerald-400 to-lime-400 flex items-center justify-center text-white text-sm font-medium">
                        {profile.first_name[0]}
                        {profile.last_name[0]}
                      </div>
                      <div>
                        <p className="font-medium text-gray-900">
                          {profile.first_name} {profile.last_name}
                        </p>
                        <p className="text-xs text-gray-400">{profile.email}</p>
                      </div>
                    </button>
                  ))}
                </div>
              ) : (
                <p className="text-center text-sm text-gray-500 py-8">Все сотрудники уже добавлены в проект</p>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Task Detail Modal */}
      {selectedTask && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onClick={() => setSelectedTask(null)}>
          <div className="bg-white rounded-2xl shadow-xl w-full max-w-lg" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between p-4 border-b border-gray-100">
              <h3 className="font-semibold text-gray-900 line-clamp-1">{selectedTask.title}</h3>
              <button onClick={() => setSelectedTask(null)} className="p-1 text-gray-400 hover:text-gray-600">
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="p-4 space-y-4">
              {/* Description */}
              {selectedTask.description && (
                <div>
                  <label className="text-xs font-medium text-gray-400 uppercase tracking-wide">Описание</label>
                  <p className="mt-1 text-sm text-gray-700 whitespace-pre-wrap">{selectedTask.description}</p>
                </div>
              )}

              {/* Meta Info */}
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-gray-400 uppercase tracking-wide">Статус</label>
                  <p className="mt-1 text-sm font-medium text-gray-900">
                    {KANBAN_COLUMNS.find((c) => c.id === selectedTask.status)?.title || selectedTask.status}
                  </p>
                </div>
                <div>
                  <label className="text-xs font-medium text-gray-400 uppercase tracking-wide">Приоритет</label>
                  <p className="mt-1">
                    <span className={`inline-flex rounded px-2 py-0.5 text-xs font-medium ${PRIORITY_STYLES[selectedTask.priority]?.class || 'bg-gray-100 text-gray-600'}`}>
                      {PRIORITY_STYLES[selectedTask.priority]?.label || '—'}
                    </span>
                  </p>
                </div>
                <div>
                  <label className="text-xs font-medium text-gray-400 uppercase tracking-wide">Исполнитель</label>
                  <div className="mt-1 flex items-center gap-2">
                    {selectedTask.assignee_id ? (
                      <>
                        <Avatar
                          src={getProfile(selectedTask.assignee_id)?.avatar_url}
                          name={getProfileName(selectedTask.assignee_id)}
                          size="xs"
                        />
                        <span className="text-sm text-gray-900">{getProfileName(selectedTask.assignee_id)}</span>
                      </>
                    ) : (
                      <span className="text-sm text-gray-400">Не назначен</span>
                    )}
                  </div>
                </div>
                {selectedTask.due_date && (
                  <div>
                    <label className="text-xs font-medium text-gray-400 uppercase tracking-wide">Дедлайн</label>
                    <p className={`mt-1 text-sm font-medium ${
                      new Date(selectedTask.due_date) < new Date() && selectedTask.status !== 'DONE'
                        ? 'text-red-500' : 'text-gray-900'
                    }`}>
                      {new Date(selectedTask.due_date).toLocaleDateString('ru', { day: 'numeric', month: 'long', year: 'numeric' })}
                    </p>
                  </div>
                )}
              </div>

              {/* Review Section */}
              {selectedTask.status === 'REVIEW' && (
                <div className="rounded-xl border border-amber-200 bg-amber-50 p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <Eye className="h-4 w-4 text-amber-600" />
                    <h4 className="text-sm font-semibold text-amber-800">Код-ревью</h4>
                  </div>

                  {reviewLoading ? (
                    <div className="flex items-center justify-center py-4">
                      <Loader2 className="h-5 w-5 animate-spin text-amber-500" />
                    </div>
                  ) : reviewStatus ? (
                    <div className="space-y-2">
                      {reviewStatus.reviewers.map((reviewer) => {
                        const profile = getProfile(reviewer.user_id);
                        return (
                          <div key={reviewer.user_id} className="flex items-center justify-between rounded-lg bg-white px-3 py-2">
                            <div className="flex items-center gap-2">
                              <Avatar
                                src={profile?.avatar_url}
                                name={getProfileName(reviewer.user_id)}
                                size="xs"
                              />
                              <span className="text-sm text-gray-900">{getProfileName(reviewer.user_id)}</span>
                            </div>
                            {reviewer.approved ? (
                              <span className="inline-flex items-center gap-1 rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-700">
                                <CheckCircle className="h-3 w-3" />
                                Одобрено
                              </span>
                            ) : (
                              <span className="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-500">
                                Ожидание
                              </span>
                            )}
                          </div>
                        );
                      })}

                      {/* Approve button — visible only if current user is a reviewer who hasn't approved yet */}
                      {user && reviewStatus.reviewers.some((r) => r.user_id === user.id && !r.approved) && (
                        <button
                          onClick={handleApproveTask}
                          disabled={reviewLoading}
                          className="mt-2 w-full inline-flex items-center justify-center gap-2 rounded-xl bg-emerald-500 px-4 py-2.5 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-60 transition-colors"
                        >
                          {reviewLoading ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <CheckCircle className="h-4 w-4" />
                          )}
                          Одобрить
                        </button>
                      )}
                    </div>
                  ) : (
                    <p className="text-sm text-amber-700">Информация о ревью недоступна</p>
                  )}
                </div>
              )}
            </div>

            {/* Modal Footer */}
            <div className="flex items-center justify-end gap-2 p-4 border-t border-gray-100">
              {canManage && (
                <button
                  onClick={() => {
                    openEditTask(selectedTask);
                    setSelectedTask(null);
                  }}
                  className="inline-flex items-center gap-2 rounded-xl border border-gray-200 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  <Edit2 className="h-4 w-4" />
                  Редактировать
                </button>
              )}
              <button
                onClick={() => setSelectedTask(null)}
                className="inline-flex items-center gap-2 rounded-xl bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200 transition-colors"
              >
                Закрыть
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}


