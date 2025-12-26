import { useCallback, useEffect, useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import {
  ArrowLeft,
  BarChart3,
  CheckCircle2,
  Edit2,
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
  type Project,
  type Task,
  type ProjectMember,
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

  // Проверка прав - менеджер, разработчик, директор или админ
  const canManage = user && ['manager', 'developer', 'director', 'admin'].includes(user.role);
  const isManagerOrHigher = user && ['manager', 'developer', 'director', 'admin'].includes(user.role);

  const [project, setProject] = useState<Project | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [members, setMembers] = useState<ProjectMember[]>([]);
  const [allProfiles, setAllProfiles] = useState<ProfileDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [metrics, setMetrics] = useState<ProjectMetrics | null>(null);

  // UI states
  const [activeTab, setActiveTab] = useState<'kanban' | 'team' | 'stats' | 'settings'>('kanban');
  const [showTaskForm, setShowTaskForm] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [showMemberModal, setShowMemberModal] = useState(false);
  const [draggingTask, setDraggingTask] = useState<Task | null>(null);
  const [dragOverColumn, setDragOverColumn] = useState<string | null>(null);

  // Task form
  const [taskTitle, setTaskTitle] = useState('');
  const [taskDescription, setTaskDescription] = useState('');
  const [taskPriority, setTaskPriority] = useState('TASK_PRIORITY_MEDIUM');
  const [taskAssignee, setTaskAssignee] = useState('');
  const [taskDueDate, setTaskDueDate] = useState('');
  const [submitting, setSubmitting] = useState(false);

  // Settings
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

      // Init settings form
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

  // Исполнителями могут быть только employee (не менеджеры и не developer)
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

  // Task handlers
  const handleCreateTask = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!taskTitle.trim()) return;
    setSubmitting(true);
    try {
      if (editingTask) {
        // Update existing task
        await updateTask(editingTask.id, {
          title: taskTitle.trim(),
          description: taskDescription.trim() || undefined,
          priority: PRIORITY_STYLES[taskPriority]?.value,
          assignee_id: taskAssignee ? Number(taskAssignee) : undefined,
          due_date: taskDueDate || undefined,
        });
        setSuccess('Задача обновлена');
      } else {
        // Create new task
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

  // Drag & Drop handlers
  // Перетаскивать может только: assignee задачи или developer
  const canDragTask = (task: Task): boolean => {
    if (!user) return false;
    // Developer может перетаскивать любые задачи
    if (user.role === 'developer') return true;
    // Assignee может перетаскивать свою задачу
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
      // Optimistic update
      setTasks((prev) => prev.map((t) => (t.id === draggingTask.id ? { ...t, status: columnId } : t)));
    } catch {
      setError('Не удалось переместить задачу');
    } finally {
      setDraggingTask(null);
    }
  };

  // Member handlers
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

  // Settings handler
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
        <Link to="/" className="text-emerald-600 hover:underline">
          Вернуться на главную
        </Link>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-white via-emerald-50/30 to-white text-gray-900">
      <div className="mx-auto max-w-7xl px-4 py-6">
        {/* Header */}
        <header className="flex flex-wrap items-start justify-between gap-4 mb-6">
          <div>
            <Link
              to="/"
              className="inline-flex items-center gap-1 text-xs text-gray-400 hover:text-emerald-600 mb-2"
            >
              <ArrowLeft className="h-3 w-3" />
              Все проекты
            </Link>
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

        {/* Messages */}
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

        {/* Tabs */}
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

        {/* Kanban Board */}
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
                        className={`rounded-lg border bg-white p-2.5 shadow-sm transition-all group ${
                          isDraggable ? 'cursor-grab active:cursor-grabbing hover:shadow-md hover:border-emerald-200' : 'cursor-default'
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
                          <span
                            className={`inline-flex rounded px-1.5 py-0.5 text-[10px] font-medium ${
                              PRIORITY_STYLES[task.priority]?.class || 'bg-gray-100 text-gray-600'
                            }`}
                          >
                            {PRIORITY_STYLES[task.priority]?.label || '—'}
                          </span>
                          
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
                    
                    {/* Empty state */}
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

        {/* Team Tab */}
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

            {/* Team stats */}
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

        {/* Stats Tab */}
        {activeTab === 'stats' && (
          <div className="grid gap-6 lg:grid-cols-2">
            {/* Progress Overview */}
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

            {/* Task Breakdown */}
            <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
              <div className="flex items-center gap-2 mb-4">
                <BarChart3 className="h-5 w-5 text-blue-500" />
                <h3 className="font-semibold text-gray-900">Распределение задач</h3>
              </div>
              
              {metrics ? (
                <div className="space-y-3">
                  <div className="flex items-center justify-between p-3 rounded-xl bg-gray-50">
                    <span className="text-sm text-gray-600">Всего задач</span>
                    <span className="font-bold text-gray-900">{metrics.total_tasks}</span>
                  </div>
                  <div className="flex items-center justify-between p-3 rounded-xl bg-emerald-50">
                    <span className="text-sm text-emerald-700">Выполнено</span>
                    <span className="font-bold text-emerald-700">{metrics.completed_tasks}</span>
                  </div>
                  <div className="flex items-center justify-between p-3 rounded-xl bg-emerald-50/50 pl-6">
                    <span className="text-xs text-emerald-600">• В срок</span>
                    <span className="font-medium text-emerald-600">{metrics.completed_on_time || 0}</span>
                  </div>
                  <div className="flex items-center justify-between p-3 rounded-xl bg-amber-50/50 pl-6">
                    <span className="text-xs text-amber-600">• С опозданием</span>
                    <span className="font-medium text-amber-600">{metrics.completed_late || 0}</span>
                  </div>
                  <div className="flex items-center justify-between p-3 rounded-xl bg-blue-50">
                    <span className="text-sm text-blue-700">В работе</span>
                    <span className="font-bold text-blue-700">{metrics.in_progress_tasks}</span>
                  </div>
                  <div className="flex items-center justify-between p-3 rounded-xl bg-red-50">
                    <span className="text-sm text-red-700">Просрочено (не завершено)</span>
                    <span className="font-bold text-red-700">{metrics.overdue_tasks}</span>
                  </div>
                </div>
              ) : (
                <p className="text-gray-500 text-sm">Загрузка статистики...</p>
              )}
            </div>

            {/* Team & Health */}
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

            {/* Health Status */}
            <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
              <div className="flex items-center gap-2 mb-4">
                <CheckCircle2 className="h-5 w-5 text-amber-500" />
                <h3 className="font-semibold text-gray-900">Здоровье проекта</h3>
              </div>
              
              {metrics ? (
                <div className="flex items-center gap-4">
                  <div className={`p-4 rounded-2xl ${
                    metrics.health_status === 'HEALTH_STATUS_HEALTHY' ? 'bg-emerald-100' :
                    metrics.health_status === 'HEALTH_STATUS_AT_RISK' ? 'bg-amber-100' :
                    'bg-red-100'
                  }`}>
                    <div className={`text-3xl font-bold ${
                      metrics.health_status === 'HEALTH_STATUS_HEALTHY' ? 'text-emerald-700' :
                      metrics.health_status === 'HEALTH_STATUS_AT_RISK' ? 'text-amber-700' :
                      'text-red-700'
                    }`}>
                      {metrics.health_status === 'HEALTH_STATUS_HEALTHY' ? '✓' :
                       metrics.health_status === 'HEALTH_STATUS_AT_RISK' ? '!' : '✗'}
                    </div>
                  </div>
                  <div>
                    <p className={`font-semibold ${
                      metrics.health_status === 'HEALTH_STATUS_HEALTHY' ? 'text-emerald-700' :
                      metrics.health_status === 'HEALTH_STATUS_AT_RISK' ? 'text-amber-700' :
                      'text-red-700'
                    }`}>
                      {metrics.health_status === 'HEALTH_STATUS_HEALTHY' ? 'Отлично' :
                       metrics.health_status === 'HEALTH_STATUS_AT_RISK' ? 'Требует внимания' : 'Критично'}
                    </p>
                    <p className="text-sm text-gray-500">
                      {metrics.health_status === 'HEALTH_STATUS_HEALTHY' 
                        ? 'Проект идёт по плану' 
                        : metrics.health_status === 'HEALTH_STATUS_AT_RISK'
                        ? 'Есть риски отставания'
                        : 'Много просроченных задач'}
                    </p>
                  </div>
                </div>
              ) : (
                <p className="text-gray-500 text-sm">Загрузка статистики...</p>
              )}
            </div>
          </div>
        )}

        {/* Settings Tab */}
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

      {/* Task Creation/Edit Modal */}
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

      {/* Add Member Modal */}
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
    </div>
  );
}
