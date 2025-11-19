import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Activity,
  AlertTriangle,
  ArrowLeft,
  ArrowRight,
  CalendarDays,
  Filter,
  Layers,
  Loader2,
  MoveRight,
  PlusCircle,
  Target,
  TimerReset,
  UserPlus,
  Users2,
} from 'lucide-react';

type ProjectId = string;
type TaskStatus = 'TODO' | 'IN_PROGRESS' | 'REVIEW' | 'DONE';
type Priority = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';

type ProjectCard = {
  id: ProjectId;
  name: string;
  manager: string;
  status: string;
  due: string;
  health: string;
  progress: number;
};

type Task = {
  id: string;
  title: string;
  assignee: string;
  priority: Priority;
  due: string;
};

type ProjectBoard = Record<TaskStatus, Task[]>;

const projectCards: ProjectCard[] = [];

const initialTaskBoards: Record<ProjectId, ProjectBoard> = {};
const emptyBoard: ProjectBoard = {
  TODO: [],
  IN_PROGRESS: [],
  REVIEW: [],
  DONE: [],
};

const projectMembers: Record<ProjectId, { name: string; role: string }[]> = {};

const boardOrder: TaskStatus[] = ['TODO', 'IN_PROGRESS', 'REVIEW', 'DONE'];
const boardLabels: Record<TaskStatus, string> = {
  TODO: 'Backlog',
  IN_PROGRESS: 'В работе',
  REVIEW: 'Проверка',
  DONE: 'Готово',
};
const statusColor: Record<TaskStatus, string> = {
  TODO: 'border-gray-200',
  IN_PROGRESS: 'border-amber-200',
  REVIEW: 'border-blue-200',
  DONE: 'border-emerald-200',
};
const statusDot: Record<TaskStatus, string> = {
  TODO: 'bg-gray-400',
  IN_PROGRESS: 'bg-amber-400',
  REVIEW: 'bg-sky-400',
  DONE: 'bg-emerald-500',
};
const priorityChip: Record<Priority, string> = {
  LOW: 'bg-gray-100 text-gray-600',
  MEDIUM: 'bg-amber-50 text-amber-700',
  HIGH: 'bg-rose-50 text-rose-700',
  CRITICAL: 'bg-rose-100 text-rose-800',
};
const emptyCopy: Record<TaskStatus, { title: string; body: string }> = {
  TODO: {
    title: 'Нет задач в Backlog',
    body: 'Создайте карточку через CreateTaskRequest, чтобы команда видела входящие.',
  },
  IN_PROGRESS: {
    title: 'В работе пусто',
    body: 'Перетащите карточку или запланируйте старт работ — статус обновится автоматически.',
  },
  REVIEW: {
    title: 'Проверок нет',
    body: 'Команда увидит здесь задачи со статусом REVIEW/READY_FOR_TEST.',
  },
  DONE: {
    title: 'Готовых задач нет',
    body: 'После MoveTaskRequest со статусом DONE карточки попадут сюда.',
  },
};

export default function ProjectsHubPage() {
  const [selectedProjectId, setSelectedProjectId] = useState<ProjectId | null>(projectCards[0]?.id ?? null);
  const [boards] = useState<Record<ProjectId, ProjectBoard>>(initialTaskBoards);
  const [isChangingStatus, setIsChangingStatus] = useState<string | null>(null);
  const [dragState, setDragState] = useState<{
    projectId: ProjectId;
    fromColumn: TaskStatus;
    taskId: string;
  } | null>(null);
  const [hoveredColumn, setHoveredColumn] = useState<TaskStatus | null>(null);

  const selectedProject = useMemo(() => projectCards.find((project) => project.id === selectedProjectId) ?? null, [selectedProjectId]);
  const board = useMemo(() => {
    if (!selectedProjectId) {
      return emptyBoard;
    }
    return boards[selectedProjectId] ?? emptyBoard;
  }, [boards, selectedProjectId]);
  const members = selectedProjectId ? projectMembers[selectedProjectId] ?? [] : [];

  const handleMoveTask = (taskId: string, fromColumn: TaskStatus, nextColumn: TaskStatus) => {
    if (fromColumn === nextColumn) return;
    setIsChangingStatus(taskId);
    setTimeout(() => {
      alert('MoveTaskRequest появится после интеграции ProjectService.');
      setIsChangingStatus(null);
    }, 250);
  };

  const handleDrop = (targetColumn: TaskStatus) => {
    if (!dragState || !selectedProjectId || dragState.projectId !== selectedProjectId) return;
    handleMoveTask(dragState.taskId, dragState.fromColumn, targetColumn);
    setDragState(null);
    setHoveredColumn(null);
  };

  return (
    <div className="relative min-h-screen bg-gradient-to-br from-white via-emerald-50/40 to-white text-gray-900">
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -left-28 top-10 h-[420px] w-[420px] rounded-full bg-emerald-100/55 blur-3xl" />
        <div className="absolute -right-12 top-64 h-[360px] w-[360px] rounded-full bg-lime-100/60 blur-3xl" />
      </div>

      <div className="relative z-10 mx-auto max-w-7xl px-6 py-10">
        <header className="flex flex-wrap items-center justify-between gap-4 border-b border-gray-100 pb-6">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.35em] text-emerald-500">Проекты</p>
            <h1 className="mt-3 text-3xl font-semibold text-gray-900">Портфель и задачи</h1>
            <p className="mt-2 text-sm text-gray-500">Список проектов, канбан, участники и действия ProjectService в одном месте.</p>
          </div>
          <div className="flex gap-2">
            <Link
              to="/projects/new"
              className="inline-flex items-center gap-2 rounded-full border border-emerald-200 bg-white px-4 py-2 text-sm font-semibold text-emerald-700"
            >
              <PlusCircle className="h-4 w-4" />
              Новый проект
            </Link>
            <Link
              to="/"
              className="inline-flex items-center gap-2 rounded-full border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-600 hover:text-emerald-600"
            >
              <ArrowLeft className="h-4 w-4" />
              К дашборду
            </Link>
          </div>
        </header>

        <section className="mt-8 grid gap-6 lg:grid-cols-[1.2fr,2.2fr]">
          <div className="space-y-6">
            <div className="rounded-[32px] border border-gray-100 bg-white/95 p-6 shadow-[0_25px_80px_rgba(6,95,70,0.08)]">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Портфель</p>
                  <h2 className="mt-2 text-xl font-semibold">Активные проекты</h2>
                </div>
                <button
                  className="inline-flex items-center gap-1 rounded-full border border-gray-200 px-4 py-2 text-xs font-semibold text-gray-400"
                  type="button"
                  disabled
                >
                  <Filter className="h-3.5 w-3.5" /> Фильтры
                </button>
              </div>
              <div className="mt-5 space-y-4">
                {projectCards.length === 0 ? (
                  <div className="rounded-3xl border-2 border-dashed border-gray-200 bg-gray-50/70 px-4 py-6 text-center text-sm text-gray-500">
                    <p className="font-semibold text-gray-900">Пока нет данных от ProjectService</p>
                    <p className="mt-1 text-xs text-gray-500">
                      Здесь появится список из ListProjectsResponse, как только подключим backend.
                    </p>
                  </div>
                ) : (
                  projectCards.map((project) => (
                    <button
                      key={project.id}
                      type="button"
                      onClick={() => setSelectedProjectId(project.id)}
                      className={`w-full rounded-3xl border px-4 py-4 text-left transition-all ${
                        selectedProjectId === project.id
                          ? 'border-emerald-200 bg-emerald-50/70 shadow-[0_20px_40px_rgba(16,185,129,0.2)]'
                          : 'border-gray-100 bg-gray-50/80 hover:border-emerald-100'
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="text-sm font-semibold text-gray-900">{project.name}</p>
                          <p className="text-xs text-gray-500">Менеджер: {project.manager}</p>
                        </div>
                        <span className="rounded-full bg-white/80 px-3 py-1 text-[11px] font-semibold uppercase text-gray-500">
                          {project.status}
                        </span>
                      </div>
                      <div className="mt-3 flex items-center justify-between text-xs text-gray-500">
                        <span>
                          <CalendarDays className="mr-1 inline h-3.5 w-3.5 text-emerald-500" /> {project.due}
                        </span>
                        <span>
                          <Activity className="mr-1 inline h-3.5 w-3.5 text-amber-500" /> {project.health}
                        </span>
                        <span>
                          <Target className="mr-1 inline h-3.5 w-3.5 text-emerald-500" /> {project.progress}%
                        </span>
                      </div>
                    </button>
                  ))
                )}
              </div>
            </div>

            <div className="rounded-[32px] border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
              <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Команда</p>
              <h2 className="mt-2 text-xl font-semibold">Участники проекта</h2>
              <div className="mt-4 space-y-3 text-sm text-gray-600">
                {members.length === 0 ? (
                  <div className="rounded-2xl border-2 border-dashed border-gray-200 bg-gray-50/70 px-4 py-6 text-center text-xs text-gray-500">
                    ListProjectMembersResponse пока не подключён, поэтому участников нет.
                  </div>
                ) : (
                  members.map((member) => (
                    <div key={member.name} className="flex items-center justify-between rounded-2xl border border-gray-100 bg-gray-50/70 px-4 py-3">
                      <div>
                        <p className="font-semibold text-gray-900">{member.name}</p>
                        <p className="text-xs text-gray-500">{member.role}</p>
                      </div>
                      <button type="button" className="rounded-full border border-gray-200 px-3 py-1 text-xs text-gray-600" disabled>
                        <Users2 className="mr-1 inline h-3.5 w-3.5" /> Роли
                      </button>
                    </div>
                  ))
                )}
              </div>
              <div className="mt-4 flex gap-2">
                <button type="button" className="flex-1 rounded-full border border-gray-200 px-4 py-2 text-xs font-semibold text-gray-400" disabled>
                  <UserPlus className="mr-1 inline h-3.5 w-3.5" /> Добавить участника
                </button>
                <button type="button" className="flex-1 rounded-full border border-gray-200 px-4 py-2 text-xs font-semibold text-gray-400" disabled>
                  <Users2 className="mr-1 inline h-3.5 w-3.5" /> Состав команды
                </button>
              </div>
            </div>
          </div>

          <div className="space-y-6">
            <div className="rounded-[32px] border border-gray-100 bg-white/95 p-6 shadow-[0_25px_80px_rgba(6,95,70,0.08)]">
              <div className="flex flex-wrap items-center justify-between gap-4">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">
                    {selectedProject ? selectedProject.name : 'Нет данных'}
                  </p>
                  <h2 className="mt-2 text-xl font-semibold">Задачи и статусы</h2>
                </div>
                <div className="flex flex-wrap gap-2 text-xs">
                  <button className="rounded-full border border-gray-200 px-4 py-2 font-semibold text-gray-400" type="button" disabled>
                    <Layers className="mr-1 inline h-3.5 w-3.5" /> Канбан
                  </button>
                  <button className="rounded-full border border-gray-200 px-4 py-2 font-semibold text-gray-400" type="button" disabled>
                    <TimerReset className="mr-1 inline h-3.5 w-3.5" /> SLA
                  </button>
                  <button className="rounded-full border border-gray-200 px-4 py-2 font-semibold text-gray-400" type="button" disabled>
                    <MoveRight className="mr-1 inline h-3.5 w-3.5" /> Журнал перемещений
                  </button>
                </div>
              </div>

              {!selectedProject && (
                <div className="mt-4 rounded-2xl border-2 border-dashed border-gray-200 bg-gray-50/80 px-4 py-3 text-sm text-gray-600">
                  Канбан заработает, когда появится ответ ListProjects + ListProjectTasks для выбранного проекта.
                </div>
              )}

              <div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <div className="rounded-2xl border border-gray-100 bg-gray-50/80 p-4 text-sm text-gray-500">
                  <p>Всего задач</p>
                  <p className="mt-2 text-2xl font-semibold text-gray-900">
                    {boardOrder.reduce((acc, column) => acc + board[column].length, 0)}
                  </p>
                </div>
                <div className="rounded-2xl border border-gray-100 bg-gray-50/80 p-4 text-sm text-gray-500">
                  <p>В работе</p>
                  <p className="mt-2 text-2xl font-semibold text-amber-600">{board.IN_PROGRESS.length}</p>
                </div>
                <div className="rounded-2xl border border-gray-100 bg-gray-50/80 p-4 text-sm text-gray-500">
                  <p>Проверка</p>
                  <p className="mt-2 text-2xl font-semibold text-blue-600">{board.REVIEW.length}</p>
                </div>
                <div className="rounded-2xl border border-gray-100 bg-gray-50/80 p-4 text-sm text-gray-500">
                  <p>Готово</p>
                  <p className="mt-2 text-2xl font-semibold text-emerald-600">{board.DONE.length}</p>
                </div>
              </div>

              <div className="mt-4 flex flex-wrap items-center justify-between gap-3 rounded-2xl bg-gray-50/80 px-4 py-3 text-xs text-gray-500">
                <p className="flex items-center gap-2 text-gray-600">
                  Перетащите карточку в другую колонку — статус обновится, а ниже отобразится событие в журнале.
                </p>
                <div className="flex flex-wrap gap-3">
                  {boardOrder.map((column) => (
                    <span key={column} className="inline-flex items-center gap-2 font-semibold text-gray-500">
                      <span className={`h-2 w-2 rounded-full ${statusDot[column]}`} />
                      {boardLabels[column]}
                    </span>
                  ))}
                </div>
              </div>

              <div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                {boardOrder.map((column) => (
                  <div
                    key={column}
                    className={`flex h-full flex-col rounded-3xl border ${statusColor[column]} bg-gray-50/60 p-4 transition-colors ${
                      hoveredColumn === column ? 'bg-emerald-50/50 border-emerald-200' : ''
                    }`}
                    onDragOver={(event) => {
                      if (!selectedProjectId || dragState?.projectId !== selectedProjectId) return;
                      event.preventDefault();
                      setHoveredColumn(column);
                    }}
                    onDragLeave={() => {
                      if (hoveredColumn === column) {
                        setHoveredColumn(null);
                      }
                    }}
                    onDrop={(event) => {
                      event.preventDefault();
                      handleDrop(column);
                    }}
                  >
                    <div className="flex items-center justify-between">
                      <h3 className="text-sm font-semibold text-gray-900">{boardLabels[column]}</h3>
                      <span className="text-xs text-gray-400">{board[column].length}</span>
                    </div>
                    <div className="mt-3 space-y-3">
                      {board[column].length ? (
                        board[column].map((task) => (
                          <div
                            key={task.id}
                            draggable
                            onDragStart={(event) => {
                              event.dataTransfer.setData('text/plain', task.id);
                              if (!selectedProjectId) return;
                              setDragState({
                                taskId: task.id,
                                fromColumn: column,
                                projectId: selectedProjectId,
                              });
                            }}
                            onDragEnd={() => {
                              setDragState(null);
                              setHoveredColumn(null);
                            }}
                            className="cursor-grab rounded-2xl border border-white/80 bg-white/90 p-3 shadow-sm active:cursor-grabbing"
                          >
                            <div className="flex items-center justify-between">
                              <p className="text-sm font-semibold text-gray-900">{task.title}</p>
                              <span className={`rounded-full px-2 py-1 text-[11px] font-semibold ${priorityChip[task.priority]}`}>
                                {task.priority}
                              </span>
                            </div>
                            <p className="mt-1 text-xs text-gray-500">{task.assignee}</p>
                            <p className="mt-1 text-xs text-gray-400">
                              <CalendarDays className="mr-1 inline h-3 w-3 text-emerald-500" /> {task.due}
                            </p>
                            {isChangingStatus === task.id && (
                              <div className="mt-2 flex items-center gap-2 text-xs text-emerald-600">
                                <Loader2 className="h-3 w-3 animate-spin" />
                                MoveTaskRequest...
                              </div>
                            )}
                          </div>
                        ))
                      ) : (
                        <div className="rounded-2xl border-2 border-dashed border-gray-200 bg-white/70 p-4 text-center text-sm text-gray-500">
                          <p className="font-semibold text-gray-900">{emptyCopy[column].title}</p>
                          <p className="mt-1 text-xs text-gray-500">{emptyCopy[column].body}</p>
                          <button
                            type="button"
                            className="mt-3 inline-flex items-center gap-2 rounded-full border border-gray-200 px-4 py-1 text-xs font-semibold text-gray-400"
                            disabled
                          >
                            <PlusCircle className="h-3.5 w-3.5" /> Новая задача
                          </button>
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="rounded-[32px] border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
                <div className="flex flex-wrap items-center justify-between gap-4">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Активность</p>
                  <h2 className="mt-2 text-xl font-semibold">Журнал действий</h2>
                </div>
                <button className="rounded-full border border-gray-200 px-4 py-2 text-xs font-semibold text-gray-400" type="button" disabled>
                  <ArrowRight className="mr-1 inline h-3.5 w-3.5" /> Export
                </button>
              </div>
              <div className="mt-4 rounded-2xl border-2 border-dashed border-gray-200 bg-gray-50/70 px-4 py-6 text-sm text-gray-600">
                Здесь будет журнал из ListProjectActivityResponse. Сейчас ProjectService ещё не отдаёт события — оставляем плейсхолдер.
              </div>
            </div>
          </div>
        </section>

        <div className="mt-8 grid gap-6 lg:grid-cols-2">
          <div className="rounded-[32px] border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
            <p className="text-xs font-semibold uppercase tracking-[0.35em] text-emerald-500">API ProjectService</p>
            <ul className="mt-4 space-y-3 text-sm text-gray-600">
              <li className="flex items-start gap-3">
                <PlusCircle className="mt-0.5 h-4 w-4 text-emerald-500" /> Create/Update/Delete Project
              </li>
              <li className="flex items-start gap-3">
                <Layers className="mt-0.5 h-4 w-4 text-emerald-500" /> Create/Update/Delete/Move Task
              </li>
              <li className="flex items-start gap-3">
                <Users2 className="mt-0.5 h-4 w-4 text-emerald-500" /> Add/Remove/List Members
              </li>
              <li className="flex items-start gap-3">
                <Activity className="mt-0.5 h-4 w-4 text-emerald-500" /> ListProjects + фильтры (manager_id, status)
              </li>
            </ul>
            <p className="mt-4 rounded-2xl bg-emerald-50/80 px-4 py-3 text-xs text-emerald-700">
              Перетаскивание уже работает локально, при интеграции сюда добавим реальные запросы MoveTask/AssignTask и обновление по WebSocket/SSE.
            </p>
          </div>

          <div className="rounded-[32px] border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
            <p className="text-xs font-semibold uppercase tracking-[0.35em] text-gray-400">Процессы</p>
            <ul className="mt-4 space-y-3 text-sm text-gray-600">
              <li className="flex items-start gap-3">
                <Target className="mt-0.5 h-4 w-4 text-emerald-500" />
                ListProjectsRequest → фильтр по статусу, пагинация
              </li>
              <li className="flex items-start gap-3">
                <Users2 className="mt-0.5 h-4 w-4 text-emerald-500" />
                ListProjectMembersResponse для списка участников
              </li>
              <li className="flex items-start gap-3">
                <AlertTriangle className="mt-0.5 h-4 w-4 text-amber-500" />
                AssignTaskRequest для смены ответственных
              </li>
              <li className="flex items-start gap-3">
                <TimerReset className="mt-0.5 h-4 w-4 text-emerald-500" />
                MoveTaskRequest + order_index для канбана
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}
