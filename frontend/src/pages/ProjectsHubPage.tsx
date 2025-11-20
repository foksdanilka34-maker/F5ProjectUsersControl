import { useState, useEffect, useMemo } from 'react';
import { Link } from 'react-router-dom';
import {
  ArrowLeft,
  Calendar,
  Clock,
  FolderKanban,
  Layout,
  MoreHorizontal,
  Plus,
  Users,
  Loader2,
  Trash2,
  X,
  Search
} from 'lucide-react';
import { projectService } from '../api/services/project.service';
import { employeeService } from '../api/services/employee.service';
import type { Project, Task, Profile, Department, Skill } from '../api/types';
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

export default function ProjectsHubPage() {
  const { user } = useAuth();
  const { showToast } = useToast();
  const [activeTab, setActiveTab] = useState('Доска задач');
  const [isLoading, setIsLoading] = useState(true);
  
  // Data state
  const [projects, setProjects] = useState<Project[]>([]);
  const [selectedProject, setSelectedProject] = useState<Project | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);

  // Add Member Modal State
  const [isAddMemberModalOpen, setIsAddMemberModalOpen] = useState(false);
  const [availableEmployees, setAvailableEmployees] = useState<Profile[]>([]);
  const [isLoadingEmployees, setIsLoadingEmployees] = useState(false);
  const [memberSearchQuery, setMemberSearchQuery] = useState('');

  // Filters
  const [departments, setDepartments] = useState<Department[]>([]);
  const [skills, setSkills] = useState<Skill[]>([]);
  const [selectedDepartmentId, setSelectedDepartmentId] = useState<string>('');
  const [selectedSkillIds, setSelectedSkillIds] = useState<string[]>([]);

  // Check if user can create projects
  const canCreateProject = user?.role === 'manager' || user?.role === 'director';

  const fetchProjects = async () => {
    setIsLoading(true);
    try {
      const data = await projectService.listProjects({ page_size: 100 }).catch(err => {
        console.error('Error fetching projects:', err);
        return { projects: [], meta: {} };
      });
      
      const projectsList = Array.isArray(data?.projects) ? data.projects : [];
      setProjects(projectsList);
      
      // Only set selected project if none is selected and we have projects
      // or if the currently selected project is not in the new list (handled by effect dependencies usually, but here we do it manually if needed)
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

  // Fetch projects on mount
  useEffect(() => {
    fetchProjects();
  }, []);

  const fetchTasks = async () => {
    if (!selectedProject) return;
    try {
      const tasksData = await projectService.listTasksByProject(selectedProject.id).catch(err => {
        console.error('Error fetching tasks:', err);
        return [];
      });
      
      setTasks(Array.isArray(tasksData) ? tasksData : []);
    } catch (error) {
      console.error('Failed to fetch project details:', error);
      setTasks([]);
    }
  };

  // Fetch tasks when selected project changes
  useEffect(() => {
    fetchTasks();
  }, [selectedProject]);

  const handleMoveTask = async (taskId: string, newStatusKey: keyof typeof STATUS_MAP) => {
    if (!selectedProject) return;
    
    const newStatus = STATUS_MAP[newStatusKey];
    
    try {
      // Optimistic update
      setTasks((prev) =>
        prev.map((t) => (t.id === taskId ? { ...t, status: newStatus } : t))
      );
      
      await projectService.moveTask(selectedProject.id, taskId, {
        new_status: newStatus,
        new_order_index: 0 // Default to top for now
      });
      await fetchTasks();
    } catch (error) {
      console.error('Failed to move task:', error);
      // Revert on failure
      await fetchTasks();
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
    // Reset filters
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
        setDepartments(deptsData || []);
        setSkills(skillsData || []);
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

  const filteredEmployees = useMemo(() => {
    return availableEmployees.filter(emp => {
      // 1. Text Search
      if (memberSearchQuery) {
        const lowerQuery = memberSearchQuery.toLowerCase();
        const matchesName = emp.first_name.toLowerCase().includes(lowerQuery) || 
                            emp.last_name.toLowerCase().includes(lowerQuery) ||
                            emp.email.toLowerCase().includes(lowerQuery);
        if (!matchesName) return false;
      }

      // 2. Department Filter
      if (selectedDepartmentId && emp.department?.id !== selectedDepartmentId) {
        return false;
      }

      // 3. Skills Filter
      if (selectedSkillIds.length > 0) {
        const empSkillIds = emp.skills?.map(s => s.id) || [];
        // Check if employee has ALL selected skills
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
      const statusKey = REVERSE_STATUS_MAP[task.status] || 'TODO';
      if (grouped[statusKey]) {
        grouped[statusKey].push(task);
      } else {
        grouped['TODO'].push(task);
      }
    });
    return grouped;
  }, [tasks]);

  const getPriorityBadge = (priority: number) => {
    // TaskPriority: 0=UNSPECIFIED, 1=LOW, 2=MEDIUM, 3=HIGH, 4=CRITICAL
    switch (priority) {
      case 3: // HIGH
      case 4: // CRITICAL
        return 'bg-rose-50 text-rose-600';
      case 2: // MEDIUM
        return 'bg-amber-50 text-amber-600';
      default: // LOW or UNSPECIFIED
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
              Канбан-доска, списки задач и управление участниками проектов.
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
                      <div className="flex -space-x-2">
                        {/* Placeholder for members if we had them */}
                        <div className="flex h-8 w-8 items-center justify-center rounded-full border-2 border-white bg-gray-100 text-xs font-medium text-gray-600">
                          +
                        </div>
                      </div>
                      {canCreateProject && (
                        <button 
                          onClick={handleOpenAddMemberModal}
                          className="rounded-full border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-600 hover:text-emerald-600 hover:border-emerald-200 transition-colors"
                        >
                          <Users className="mr-2 inline h-4 w-4" />
                          Участники
                        </button>
                      )}
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

                  <div className="mt-6 flex border-b border-gray-100">
                    {['Доска задач', 'Список', 'Календарь'].map((tab) => (
                      <button
                        key={tab}
                        onClick={() => setActiveTab(tab)}
                        className={`mr-6 border-b-2 pb-3 text-sm font-medium transition-colors ${
                          activeTab === tab
                            ? 'border-emerald-500 text-emerald-600'
                            : 'border-transparent text-gray-500 hover:text-gray-700'
                        }`}
                      >
                        {tab}
                      </button>
                    ))}
                  </div>

                  {activeTab === 'Доска задач' && (
                    <div className="mt-6 grid gap-6 md:grid-cols-3">
                      {columns.map((col) => (
                        <div key={col.id} className="flex flex-col rounded-3xl bg-gray-50/50 p-4">
                          <div className="mb-4 flex items-center justify-between">
                            <div className="flex items-center gap-2">
                              <span className={`rounded-full px-2 py-0.5 text-xs font-semibold ${col.color}`}>
                                {tasksByStatus[col.id]?.length || 0}
                              </span>
                              <h3 className="text-sm font-semibold text-gray-700">{col.title}</h3>
                            </div>
                            <button className="text-gray-400 hover:text-gray-600">
                              <Plus className="h-4 w-4" />
                            </button>
                          </div>
                          
                          <div className="flex-1 space-y-3">
                            {tasksByStatus[col.id]?.map((task) => (
                              <div
                                key={task.id}
                                className="group relative rounded-2xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:shadow-md"
                              >
                                <div className="mb-2 flex items-start justify-between">
                                  <span className={`rounded-full px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider ${getPriorityBadge(task.priority)}`}>
                                    {getPriorityLabel(task.priority)}
                                  </span>
                                  <button className="opacity-0 transition-opacity group-hover:opacity-100">
                                    <MoreHorizontal className="h-4 w-4 text-gray-400" />
                                  </button>
                                </div>
                                <h4 className="font-medium text-gray-900">{task.title}</h4>
                                <p className="mt-1 text-xs text-gray-500 line-clamp-2">{task.description}</p>
                                
                                <div className="mt-4 flex items-center justify-between border-t border-gray-50 pt-3">
                                  <div className="flex items-center gap-2 text-xs text-gray-400">
                                    <Clock className="h-3.5 w-3.5" />
                                    <span>{task.due_date ? new Date(task.due_date).toLocaleDateString() : 'No date'}</span>
                                  </div>
                                  {/* Simple move controls for now */}
                                  <div className="flex gap-1">
                                    {col.id !== 'TODO' && (
                                      <button 
                                        onClick={() => handleMoveTask(task.id, 'TODO')}
                                        className="text-[10px] text-gray-400 hover:text-emerald-600"
                                        title="Move to Todo"
                                      >
                                        ←
                                      </button>
                                    )}
                                    {col.id !== 'IN_PROGRESS' && (
                                      <button 
                                        onClick={() => handleMoveTask(task.id, 'IN_PROGRESS')}
                                        className="text-[10px] text-gray-400 hover:text-emerald-600"
                                        title="Move to In Progress"
                                      >
                                        ↔
                                      </button>
                                    )}
                                    {col.id !== 'DONE' && (
                                      <button 
                                        onClick={() => handleMoveTask(task.id, 'DONE')}
                                        className="text-[10px] text-gray-400 hover:text-emerald-600"
                                        title="Move to Done"
                                      >
                                        →
                                      </button>
                                    )}
                                  </div>
                                </div>
                              </div>
                            ))}
                            {tasksByStatus[col.id]?.length === 0 && (
                              <div className="flex h-24 items-center justify-center rounded-2xl border border-dashed border-gray-200 text-xs text-gray-400">
                                Нет задач
                              </div>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  )}

                  {activeTab === 'Список' && (
                    <div className="mt-6 text-center py-12 text-gray-500">
                      <Layout className="mx-auto h-12 w-12 text-gray-300 mb-3" />
                      <p>Режим списка в разработке</p>
                    </div>
                  )}

                  {activeTab === 'Календарь' && (
                    <div className="mt-6 text-center py-12 text-gray-500">
                      <Calendar className="mx-auto h-12 w-12 text-gray-300 mb-3" />
                      <p>Календарь в разработке</p>
                    </div>
                  )}
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
    </div>
  );
}
