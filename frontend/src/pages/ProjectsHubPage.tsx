import { useState, useEffect, useMemo } from 'react';
import { Link } from 'react-router-dom';
import {
  ArrowLeft,
  Clock,
  FolderKanban,
  Plus,
  Loader2,
  Trash2,
  X,
  Search,
  Pencil
} from 'lucide-react';
import { projectService } from '../api/services/project.service';
import { employeeService } from '../api/services/employee.service';
import type { Project, Task, Profile, Department, Skill, ProjectMember } from '../api/types';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';

// Map UI column IDs to API TaskStatus enum values
// TaskStatus: 0=UNSPECIFIED, 1=TODO, 2=IN_PROGRESS, 3=REVIEW, 4=DONE
const STATUS_MAP = {
  TODO: 1,
  IN_PROGRESS: 2,
  DONE: 4
} as const;

// Reverse map for grouping
const REVERSE_STATUS_MAP: Record<number, string> = {
  1: 'TODO',
  2: 'IN_PROGRESS',
  3: 'IN_PROGRESS', // Map REVIEW to IN_PROGRESS for this simple board
  4: 'DONE'
};

const columns = [
  { id: 'TODO', title: 'К выполнению', color: 'bg-gray-100 text-gray-600' },
  { id: 'IN_PROGRESS', title: 'В работе', color: 'bg-blue-50 text-blue-600' },
  { id: 'DONE', title: 'Готово', color: 'bg-emerald-50 text-emerald-600' },
];

interface TaskModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (data: any) => Promise<void>;
  task?: Task | null;
  initialStatus?: number;
  members: ProjectMember[];
}

const safeDate = (dateStr?: string | any) => {
  if (!dateStr) return '';
  
  // Handle Protobuf Timestamp object { seconds: number, nanos: number }
  if (typeof dateStr === 'object' && dateStr !== null && 'seconds' in dateStr) {
    const seconds = Number(dateStr.seconds);
    const d = new Date(seconds * 1000);
    return !isNaN(d.getTime()) ? d.toISOString().split('T')[0] : '';
  }

  // Handle ISO string
  if (typeof dateStr === 'string') {
    const d = new Date(dateStr);
    if (!isNaN(d.getTime())) {
      return d.toISOString().split('T')[0];
    }
  }
  
  return '';
};

function TaskModal({ isOpen, onClose, onSave, task, initialStatus, members }: TaskModalProps) {
  const [title, setTitle] = useState(task?.title || '');
  const [description, setDescription] = useState(task?.description || '');
  const [priority, setPriority] = useState<number>(task?.priority || 1);
  const [assigneeId, setAssigneeId] = useState(task?.assignee_id || '');
  const [dueDate, setDueDate] = useState(safeDate(task?.due_date) || new Date().toISOString().split('T')[0]);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setTitle(task?.title || '');
      setDescription(task?.description || '');
      setPriority(task?.priority || 1);
      setAssigneeId(task?.assignee_id || '');
      setDueDate(safeDate(task?.due_date) || new Date().toISOString().split('T')[0]);
    }
  }, [isOpen, task]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    try {
      await onSave({
        title,
        description,
        priority: Number(priority),
        assignee_id: assigneeId || undefined,
        due_date: dueDate ? new Date(dueDate).toISOString() : new Date().toISOString(),
        status: task ? undefined : initialStatus // Only set status on create
      });
    } finally {
      setIsLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/20 backdrop-blur-sm p-4 transition-all duration-300 ease-out">
      <div className="w-full max-w-md rounded-3xl bg-white p-6 shadow-2xl animate-in fade-in zoom-in-95 duration-300 ease-out">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-lg font-semibold text-gray-900">{task ? 'Редактировать задачу' : 'Новая задача'}</h3>
          <button onClick={onClose} className="rounded-full p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600">
            <X className="h-5 w-5" />
          </button>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Название</label>
            <input
              required
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full rounded-xl border border-gray-200 bg-gray-50 py-2 px-3 text-sm focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Описание</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
              className="w-full rounded-xl border border-gray-200 bg-gray-50 py-2 px-3 text-sm focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Приоритет</label>
              <select
                value={priority}
                onChange={(e) => setPriority(Number(e.target.value))}
                className="w-full rounded-xl border border-gray-200 bg-gray-50 py-2 px-3 text-sm focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
              >
                <option value={1}>Low</option>
                <option value={2}>Medium</option>
                <option value={3}>High</option>
                <option value={4}>Critical</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Срок</label>
              <input
                type="date"
                value={dueDate}
                onChange={(e) => setDueDate(e.target.value)}
                className="w-full rounded-xl border border-gray-200 bg-gray-50 py-2 px-3 text-sm focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
              />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Исполнитель</label>
            <select
              value={assigneeId}
              onChange={(e) => setAssigneeId(e.target.value)}
              className="w-full rounded-xl border border-gray-200 bg-gray-50 py-2 px-3 text-sm focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
            >
              <option value="">Не назначен</option>
              {members.map(m => (
                <option key={m.user_id} value={m.user_id}>{m.full_name || 'Unknown'}</option>
              ))}
            </select>
          </div>
          <div className="pt-4 flex justify-end gap-3">
            <button
              type="button"
              onClick={onClose}
              className="rounded-xl px-4 py-2 text-sm font-medium text-gray-600 hover:bg-gray-100"
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={isLoading}
              className="rounded-xl bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-700 disabled:opacity-50"
            >
              {isLoading ? 'Сохранение...' : 'Сохранить'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default function ProjectsHubPage() {
  const { user } = useAuth();
  const { showToast } = useToast();
  const [isLoading, setIsLoading] = useState(true);
  
  // Data state
  const [projects, setProjects] = useState<Project[]>([]);
  const [selectedProject, setSelectedProject] = useState<Project | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [projectMembers, setProjectMembers] = useState<ProjectMember[]>([]);

  // Add Member Modal State
  const [isAddMemberModalOpen, setIsAddMemberModalOpen] = useState(false);
  const [availableEmployees, setAvailableEmployees] = useState<Profile[]>([]);
  const [isLoadingEmployees, setIsLoadingEmployees] = useState(false);
  const [memberSearchQuery, setMemberSearchQuery] = useState('');

  // Task Modal State
  const [isTaskModalOpen, setIsTaskModalOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [taskModalInitialStatus, setTaskModalInitialStatus] = useState<number>(STATUS_MAP.TODO);

  // Filters
  const [departments, setDepartments] = useState<Department[]>([]);
  const [skills, setSkills] = useState<Skill[]>([]);
  const [selectedDepartmentId, setSelectedDepartmentId] = useState<string>('');
  const [selectedSkillIds, setSelectedSkillIds] = useState<string[]>([]);

  const [draggedTaskId, setDraggedTaskId] = useState<string | null>(null);

  // Check if user can create projects (Manager/Director/Admin)
  const isManager = user?.role === 'manager' || user?.role === 'director' || user?.role === 'admin';
  const isSpecialist = user?.role === 'specialist';
  const canCreateProject = isManager;

  const fetchProjects = async () => {
    setIsLoading(true);
    try {
      const data = await projectService.listProjects({ page_size: 100 }).catch(err => {
        console.error('Error fetching projects:', err);
        return { projects: [], meta: {} };
      });
      
      const projectsList = Array.isArray(data?.projects) ? data.projects : [];
      setProjects(projectsList);
      
      if (!selectedProject && projectsList.length > 0) {
        setSelectedProject(projectsList[0]);
      }
    } catch (error) {
      console.error('Failed to fetch projects:', error);
      setProjects([]);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchProjects();
  }, []);

  const fetchProjectData = async () => {
    if (!selectedProject) return;
    try {
      const [tasksData, membersData] = await Promise.all([
        projectService.listTasksByProject(selectedProject.id).catch(() => []),
        projectService.listProjectMembers(selectedProject.id).catch(() => [])
      ]);
      
      setTasks(Array.isArray(tasksData) ? tasksData : []);
      setProjectMembers(Array.isArray(membersData) ? membersData : []);
    } catch (error) {
      console.error('Failed to fetch project details:', error);
      setTasks([]);
      setProjectMembers([]);
    }
  };

  useEffect(() => {
    fetchProjectData();
  }, [selectedProject]);

  const handleRemoveMember = async (userId: string) => {
    if (!selectedProject || !window.confirm('Удалить участника из проекта?')) return;
    try {
      await projectService.removeMemberFromProject(selectedProject.id, userId);
      showToast('Участник удален', 'success');
      fetchProjectData();
    } catch (error) {
      console.error('Failed to remove member:', error);
      showToast('Не удалось удалить участника', 'error');
    }
  };

  const handleDragStart = (e: React.DragEvent, taskId: string) => {
    e.dataTransfer.setData('text/plain', taskId);
    e.dataTransfer.effectAllowed = 'move';
    setDraggedTaskId(taskId);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
  };

  const handleDrop = async (e: React.DragEvent, statusKey: keyof typeof STATUS_MAP) => {
    e.preventDefault();
    const taskId = e.dataTransfer.getData('text/plain');
    if (!taskId) return;

    const targetStatus = STATUS_MAP[statusKey];
    
    // Find the task to check if status is actually changing
    const task = tasks.find(t => t.id === taskId);
    if (task && task.status !== targetStatus) {
      await handleMoveTask(taskId, statusKey);
    }
    
    setDraggedTaskId(null);
  };

  const handleMoveTask = async (taskId: string, newStatusKey: keyof typeof STATUS_MAP) => {
    if (!selectedProject) return;
    
    const newStatus = STATUS_MAP[newStatusKey];
    
    try {
      setTasks((prev) =>
        prev.map((t) => (t.id === taskId ? { ...t, status: newStatus } : t))
      );
      
      await projectService.moveTask(selectedProject.id, taskId, {
        new_status: newStatus,
        new_order_index: 0
      });
      await fetchProjectData();
    } catch (error) {
      console.error('Failed to move task:', error);
      showToast('Не удалось переместить задачу', 'error');
      await fetchProjectData();
    }
  };

  const handleDeleteProject = async (project: Project) => {
    if (!window.confirm(`Вы уверены, что хотите удалить проект "${project.name}"?`)) {
      return;
    }

    try {
      await projectService.deleteProject(project.id);
      showToast('Проект успешно удален', 'success');
      
      await fetchProjects();
      if (selectedProject?.id === project.id) {
        setSelectedProject(null);
      }
    } catch (error) {
      console.error('Failed to delete project:', error);
      showToast('Не удалось удалить проект', 'error');
    }
  };

  const handleOpenAddMemberModal = async () => {
    setIsAddMemberModalOpen(true);
    setMemberSearchQuery('');
    setSelectedDepartmentId('');
    setSelectedSkillIds([]);

    if (availableEmployees.length === 0) {
      setIsLoadingEmployees(true);
      try {
        const [profilesData, deptsData, skillsData] = await Promise.all([
          employeeService.listProfiles({ page_size: 100 }),
          employeeService.listDepartments(),
          employeeService.listSkills()
        ]);
        
        const activeEmployees = (profilesData.profiles || []).filter(p => p.is_active !== false);
        setAvailableEmployees(activeEmployees);
        setDepartments(Array.isArray(deptsData) ? deptsData : []);
        setSkills(Array.isArray(skillsData) ? skillsData : []);
      } catch (error) {
        console.error('Failed to fetch employees/metadata:', error);
        showToast('Не удалось загрузить данные', 'error');
      } finally {
        setIsLoadingEmployees(false);
      }
    }
  };

  const handleAddMember = async (userId: string) => {
    if (!selectedProject) return;

    try {
      await projectService.addMemberToProject(selectedProject.id, { user_id: userId });
      showToast('Сотрудник добавлен в проект', 'success');
      setIsAddMemberModalOpen(false);
      fetchProjectData();
    } catch (error) {
      console.error('Failed to add member:', error);
      showToast('Не удалось добавить участника', 'error');
    }
  };

  const toggleSkillFilter = (skillId: string) => {
    setSelectedSkillIds(prev => 
      prev.includes(skillId) 
        ? prev.filter(id => id !== skillId)
        : [...prev, skillId]
    );
  };

  // Task Handlers
  const handleOpenCreateTask = (status: number) => {
    setEditingTask(null);
    setTaskModalInitialStatus(status);
    setIsTaskModalOpen(true);
  };

  const handleOpenEditTask = (task: Task) => {
    setEditingTask(task);
    setTaskModalInitialStatus(task.status);
    setIsTaskModalOpen(true);
  };

  const handleSaveTask = async (taskData: any) => {
    if (!selectedProject) return;
    
    try {
      if (editingTask) {
        await projectService.updateTask(selectedProject.id, editingTask.id, taskData);
        showToast('Задача обновлена', 'success');
      } else {
        await projectService.createTask(selectedProject.id, {
          ...taskData,
          status: taskModalInitialStatus
        });
        showToast('Задача создана', 'success');
      }
      setIsTaskModalOpen(false);
      fetchProjectData();
    } catch (error) {
      console.error('Failed to save task:', error);
      showToast('Не удалось сохранить задачу', 'error');
    }
  };

  const handleDeleteTask = async (taskId: string) => {
    if (!selectedProject || !window.confirm('Удалить задачу?')) return;
    try {
      await projectService.deleteTask(selectedProject.id, taskId);
      showToast('Задача удалена', 'success');
      fetchProjectData();
    } catch (error) {
      console.error('Failed to delete task:', error);
      showToast('Не удалось удалить задачу', 'error');
    }
  };

  const filteredEmployees = useMemo(() => {
    return availableEmployees.filter(emp => {
      if (memberSearchQuery) {
        const lowerQuery = memberSearchQuery.toLowerCase();
        const matchesName = (emp.first_name || '').toLowerCase().includes(lowerQuery) || 
                            (emp.last_name || '').toLowerCase().includes(lowerQuery) ||
                            (emp.email || '').toLowerCase().includes(lowerQuery);
        if (!matchesName) return false;
      }

      if (selectedDepartmentId && emp.department?.id !== selectedDepartmentId) {
        return false;
      }

      if (selectedSkillIds.length > 0) {
        const empSkillIds = emp.skills?.map(s => s.id) || [];
        const hasAllSkills = selectedSkillIds.every(id => empSkillIds.includes(id));
        if (!hasAllSkills) return false;
      }

      return true;
    });
  }, [availableEmployees, memberSearchQuery, selectedDepartmentId, selectedSkillIds]);

  const tasksByStatus = useMemo(() => {
    const grouped: Record<string, Task[]> = {
      TODO: [],
      IN_PROGRESS: [],
      DONE: []
    };
    
    if (!Array.isArray(tasks)) return grouped;

    tasks.forEach((task) => {
      if (!task) return;
      
      // Filter for specialists: only show tasks assigned to them
      if (isSpecialist && task.assignee_id !== user?.userId) {
        return;
      }

      const statusKey = REVERSE_STATUS_MAP[task.status] || 'TODO';
      if (grouped[statusKey]) {
        grouped[statusKey].push(task);
      } else {
        grouped['TODO'].push(task);
      }
    });
    return grouped;
  }, [tasks, isSpecialist, user]);

  const getPriorityBadge = (priority: number) => {
    switch (priority) {
      case 3:
      case 4:
        return 'bg-rose-50 text-rose-600';
      case 2:
        return 'bg-amber-50 text-amber-600';
      default:
        return 'bg-blue-50 text-blue-600';
    }
  };

  const getPriorityLabel = (priority: number) => {
    switch (priority) {
      case 4: return 'CRITICAL';
      case 3: return 'HIGH';
      case 2: return 'MEDIUM';
      case 1: return 'LOW';
      default: return 'NORMAL';
    }
  };

  if (isLoading && projects.length === 0) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Loader2 className="w-12 h-12 text-emerald-500 animate-spin" />
      </div>
    );
  }

  return (
    <div className="relative min-h-screen bg-linear-to-br from-white via-emerald-50/40 to-white text-gray-900">
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -left-24 top-10 h-[420px] w-[420px] rounded-full bg-emerald-100/60 blur-3xl" />
        <div className="absolute -right-16 top-52 h-[360px] w-[360px] rounded-full bg-lime-100/60 blur-3xl" />
      </div>

      <div className="relative z-10 mx-auto max-w-7xl px-6 py-10">
        <header className="flex flex-wrap items-center justify-between gap-4 border-b border-gray-100 pb-6">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.35em] text-emerald-500">Проекты</p>
            <h1 className="mt-3 text-3xl font-semibold text-gray-900">Управление задачами</h1>
            <p className="mt-2 text-sm text-gray-500">
              Канбан-доска и управление участниками проектов.
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

        <div className="mt-8 grid gap-6 lg:grid-cols-[300px,1fr]">
          {/* Sidebar: Project List */}
          <aside className="space-y-6">
            <div className="rounded-4xl border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-gray-900">Мои проекты</h2>
                {canCreateProject && (
                  <Link to="/projects/new" className="rounded-full p-2 text-gray-400 hover:bg-gray-50 hover:text-emerald-600">
                    <Plus className="h-5 w-5" />
                  </Link>
                )}
              </div>
              <div className="mt-4 space-y-2">
                {projects.map((project) => (
                  <button
                    key={project.id}
                    onClick={() => setSelectedProject(project)}
                    className={`flex w-full items-center justify-between rounded-2xl px-4 py-3 text-left transition-all ${
                      selectedProject?.id === project.id
                        ? 'bg-emerald-50 text-emerald-900 ring-1 ring-emerald-200'
                        : 'bg-white text-gray-600 hover:bg-gray-50'
                    }`}
                  >
                    <div className="flex items-center gap-3">
                      <div className={`flex h-8 w-8 items-center justify-center rounded-full ${
                        selectedProject?.id === project.id ? 'bg-emerald-100 text-emerald-600' : 'bg-gray-100 text-gray-500'
                      }`}>
                        <FolderKanban className="h-4 w-4" />
                      </div>
                      <span className="font-medium">{project.name}</span>
                    </div>
                  </button>
                ))}
                {projects.length === 0 && (
                  <div className="text-center py-4 text-gray-500 text-sm">
                    Нет активных проектов
                  </div>
                )}
              </div>
            </div>
          </aside>

          {/* Main Content */}
          <div className="space-y-6">
            {selectedProject ? (
              <>
                <div className="rounded-4xl border border-gray-100 bg-white/95 p-6 shadow-[0_25px_80px_rgba(6,95,70,0.08)]">
                  <div className="flex flex-wrap items-center justify-between gap-4">
                    <div>
                      <div className="flex items-center gap-3">
                        <h2 className="text-2xl font-semibold text-gray-900">{selectedProject.name}</h2>
                        <span className="rounded-full bg-emerald-50 px-3 py-1 text-xs font-medium text-emerald-700">
                          Active
                        </span>
                      </div>
                      <p className="mt-1 text-sm text-gray-500">{selectedProject.description}</p>
                    </div>
                    <div className="flex items-center gap-3">
                      {canCreateProject && (
                        <button 
                          onClick={() => handleDeleteProject(selectedProject)}
                          className="rounded-full border border-gray-200 bg-white p-2 text-gray-400 hover:text-rose-600 hover:border-rose-200 hover:bg-rose-50 transition-colors"
                          title="Удалить проект"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      )}
                    </div>
                  </div>

                  {/* Team Section */}
                  <div className="mt-6 border-t border-gray-100 pt-4">
                    <div className="flex items-center justify-between mb-3">
                        <h3 className="text-sm font-semibold text-gray-900">Команда проекта</h3>
                         {canCreateProject && (
                            <button 
                              onClick={handleOpenAddMemberModal}
                              className="text-xs font-medium text-emerald-600 hover:text-emerald-700 flex items-center gap-1"
                            >
                              <Plus className="h-3 w-3" /> Добавить
                            </button>
                          )}
                    </div>
                    <div className="flex flex-wrap gap-3">
                      {projectMembers.map((member) => (
                        <div key={member.user_id} className="flex items-center gap-2 rounded-full border border-gray-100 bg-gray-50 px-3 py-1.5 pr-4">
                           <div className="flex h-6 w-6 items-center justify-center rounded-full bg-emerald-100 text-xs font-bold text-emerald-700">
                              {member.full_name?.[0] || '?'}
                           </div>
                           <div className="flex flex-col">
                              <span className="text-xs font-medium text-gray-700">{member.full_name || 'Unknown'}</span>
                              <span className="text-[10px] text-gray-400">{member.role}</span>
                           </div>
                           {canCreateProject && (
                              <button 
                                onClick={() => handleRemoveMember(member.user_id)}
                                className="ml-2 text-gray-400 hover:text-rose-500"
                              >
                                <X className="h-3 w-3" />
                              </button>
                           )}
                        </div>
                      ))}
                      {projectMembers.length === 0 && (
                        <p className="text-sm text-gray-400 italic">Нет участников</p>
                      )}
                    </div>
                  </div>

                  {/* Task Board */}
                  <div className="mt-6 grid gap-6 md:grid-cols-3">
                    {columns.map((col) => (
                      <div 
                        key={col.id} 
                        className="flex flex-col rounded-3xl bg-gray-50/50 p-4 transition-colors"
                        onDragOver={handleDragOver}
                        onDrop={(e) => handleDrop(e, col.id as keyof typeof STATUS_MAP)}
                      >
                        <div className="mb-4 flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <span className={`rounded-full px-2 py-0.5 text-xs font-semibold ${col.color}`}>
                              {tasksByStatus[col.id]?.length || 0}
                            </span>
                            <h3 className="text-sm font-semibold text-gray-700">{col.title}</h3>
                          </div>
                          {col.id === 'TODO' && (
                            <button 
                              onClick={() => handleOpenCreateTask(STATUS_MAP[col.id as keyof typeof STATUS_MAP])}
                              className="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-100 text-emerald-600 opacity-80 transition-all hover:scale-110 hover:opacity-100 hover:shadow-md"
                              title="Добавить задачу"
                            >
                              <Plus className="h-5 w-5" />
                            </button>
                          )}
                        </div>
                        
                        <div className="flex-1 space-y-3">
                          {tasksByStatus[col.id]?.map((task) => (
                            <div
                              key={task.id}
                              draggable={!isManager}
                              onDragStart={(e) => handleDragStart(e, task.id)}
                              className={`group relative rounded-2xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:shadow-md ${
                                !isManager ? 'cursor-grab active:cursor-grabbing' : 'cursor-default'
                              } ${
                                draggedTaskId === task.id ? 'opacity-50' : ''
                              }`}
                            >
                              <div className="mb-2 flex items-start justify-between">
                                <span className={`rounded-full px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider ${getPriorityBadge(task.priority)}`}>
                                  {getPriorityLabel(task.priority)}
                                </span>
                                <div className="flex gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                                  <button 
                                    onClick={() => handleOpenEditTask(task)}
                                    className="text-gray-400 hover:text-emerald-600"
                                    title="Редактировать"
                                  >
                                    <Pencil className="h-3.5 w-3.5" />
                                  </button>
                                  <button 
                                    onClick={() => handleDeleteTask(task.id)}
                                    className="text-gray-400 hover:text-rose-600"
                                    title="Удалить"
                                  >
                                    <Trash2 className="h-3.5 w-3.5" />
                                  </button>
                                </div>
                              </div>
                              <h4 className="font-medium text-gray-900">{task.title}</h4>
                              <p className="mt-1 text-xs text-gray-500 line-clamp-2">{task.description}</p>
                              
                              {task.assignee_id && (
                                <div className="mt-2 flex items-center gap-1 text-xs text-gray-500">
                                  <div className="h-4 w-4 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-[8px] font-bold">
                                    {(projectMembers.find(m => m.user_id === task.assignee_id)?.full_name?.[0]) || '?'}
                                  </div>
                                  <span>{projectMembers.find(m => m.user_id === task.assignee_id)?.full_name || 'Unknown'}</span>
                                </div>
                              )}
                              
                              <div className="mt-4 flex items-center justify-between border-t border-gray-50 pt-3">
                                <div className="flex items-center gap-2 text-xs text-gray-400">
                                  <Clock className="h-3.5 w-3.5" />
                                  <span>{safeDate(task.due_date) ? new Date(safeDate(task.due_date)).toLocaleDateString() : 'No date'}</span>
                                </div>
                              </div>
                            </div>
                          ))}
                          {tasksByStatus[col.id]?.length === 0 && col.id === 'TODO' && (
                            <button 
                              onClick={() => handleOpenCreateTask(STATUS_MAP[col.id as keyof typeof STATUS_MAP])}
                              className="flex h-24 w-full items-center justify-center rounded-2xl border border-dashed border-gray-200 text-xs text-gray-400 hover:border-emerald-200 hover:bg-emerald-50/50 hover:text-emerald-600 transition-all"
                            >
                              + Добавить задачу
                            </button>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </>
            ) : (
              <div className="flex h-full flex-col items-center justify-center rounded-4xl border border-gray-100 bg-white/95 p-12 text-center shadow-[0_25px_80px_rgba(6,95,70,0.08)]">
                <FolderKanban className="mb-4 h-16 w-16 text-emerald-100" />
                <h3 className="text-xl font-semibold text-gray-900">Выберите проект</h3>
                <p className="mt-2 text-gray-500">
                  Выберите проект из списка слева, чтобы увидеть задачи и детали.
                </p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Add Member Modal */}
      {isAddMemberModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/20 backdrop-blur-sm p-4">
          <div className="w-full max-w-md rounded-3xl bg-white p-6 shadow-2xl animate-in fade-in zoom-in-95 duration-200">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-lg font-semibold text-gray-900">Добавить участника</h3>
              <button 
                onClick={() => setIsAddMemberModalOpen(false)}
                className="rounded-full p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="relative mb-4">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
              <input
                type="text"
                placeholder="Поиск сотрудника..."
                value={memberSearchQuery}
                onChange={(e) => setMemberSearchQuery(e.target.value)}
                className="w-full rounded-xl border border-gray-200 bg-gray-50 py-2 pl-10 pr-4 text-sm focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
              />
            </div>

            {/* Filters */}
            <div className="mb-4 space-y-3">
              {/* Department Select */}
              <select
                value={selectedDepartmentId}
                onChange={(e) => setSelectedDepartmentId(e.target.value)}
                className="w-full rounded-xl border border-gray-200 bg-gray-50 py-2 px-3 text-sm focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
              >
                <option value="">Все отделы</option>
                {departments.map(dept => (
                  <option key={dept.id} value={dept.id}>{dept.name}</option>
                ))}
              </select>

              {/* Skills Chips */}
              {skills.length > 0 && (
                <div className="flex flex-wrap gap-2 max-h-24 overflow-y-auto">
                  {skills.map(skill => (
                    <button
                      key={skill.id}
                      onClick={() => toggleSkillFilter(skill.id)}
                      className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                        selectedSkillIds.includes(skill.id)
                          ? 'bg-emerald-100 text-emerald-700 border border-emerald-200'
                          : 'bg-gray-100 text-gray-600 border border-gray-200 hover:bg-gray-200'
                      }`}
                    >
                      {skill.name}
                    </button>
                  ))}
                </div>
              )}
            </div>

            <div className="max-h-[300px] overflow-y-auto space-y-2 pr-1">
              {isLoadingEmployees ? (
                <div className="flex justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin text-emerald-500" />
                </div>
              ) : filteredEmployees.length > 0 ? (
                filteredEmployees.map((employee) => (
                  <div key={employee.id} className="flex items-center justify-between rounded-xl border border-gray-100 p-3 hover:bg-gray-50">
                    <div className="flex items-center gap-3">
                      <div className="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-100 text-emerald-700 text-xs font-bold">
                        {employee.first_name[0]}{employee.last_name[0]}
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-900">
                          {employee.first_name} {employee.last_name}
                        </p>
                        <p className="text-xs text-gray-500">{employee.position?.name || 'Сотрудник'}</p>
                      </div>
                    </div>
                    <button
                      onClick={() => handleAddMember(employee.id)}
                      className="rounded-full bg-emerald-50 p-2 text-emerald-600 hover:bg-emerald-100 transition-colors"
                      title="Добавить"
                    >
                      <Plus className="h-4 w-4" />
                    </button>
                  </div>
                ))
              ) : (
                <p className="text-center text-sm text-gray-500 py-4">Сотрудники не найдены</p>
              )}
            </div>
          </div>
        </div>
      )}

      <TaskModal
        isOpen={isTaskModalOpen}
        onClose={() => setIsTaskModalOpen(false)}
        onSave={handleSaveTask}
        task={editingTask}
        initialStatus={taskModalInitialStatus}
        members={projectMembers}
      />
    </div>
  );
}
