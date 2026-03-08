import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  ArrowRight,
  Calendar,
  FolderKanban,
  Grid3X3,
  LayoutList,
  Loader2,
  PlusCircle,
  Search,
  User,
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { getProjects, type Project } from '../services/projectService';
import { projectStatusLabels, projectStatusColors } from '../lib/constants';
import { listProfiles } from '../services/employeeService';
import type { ProfileDTO } from '../services/types';

type ViewMode = 'grid' | 'list';
type StatusFilter = 'all' | 'ACTIVE' | 'ON_HOLD' | 'ARCHIVED';

export default function ProjectsHubPage() {
  const { user } = useAuth();
  const [projects, setProjects] = useState<Project[]>([]);
  const [profiles, setProfiles] = useState<ProfileDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');

  const canCreate = user && ['manager', 'developer', 'director', 'admin'].includes(user.role);

  const isManagerOrHigher = user && ['admin', 'director', 'manager', 'developer'].includes(user.role);

  useEffect(() => {
    const loadData = async () => {
      try {

        const params = isManagerOrHigher ? {} : { member_id: user?.id };
        const [projectsData, profilesData] = await Promise.all([
          getProjects(params),
          listProfiles({ pageSize: 100 }),
        ]);
        setProjects(projectsData);
        setProfiles(profilesData.profiles || []);
      } catch (err) {
        console.error('Failed to load projects:', err);
        setError('Не удалось загрузить проекты');
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, [isManagerOrHigher, user?.id]);

  const getManagerName = (managerId: number) => {
    const manager = profiles.find(p => p.id === managerId);
    return manager ? `${manager.first_name} ${manager.last_name}` : '—';
  };

  const filteredProjects = projects.filter((project) => {

    if (statusFilter !== 'all' && project.status !== statusFilter) return false;

    if (!searchTerm.trim()) return true;
    const term = searchTerm.toLowerCase();
    return (
      project.name.toLowerCase().includes(term) ||
      project.description?.toLowerCase().includes(term)
    );
  });

  const statusCounts = {
    all: projects.length,
    ACTIVE: projects.filter(p => p.status === 'ACTIVE').length,
    ON_HOLD: projects.filter(p => p.status === 'ON_HOLD').length,
    ARCHIVED: projects.filter(p => p.status === 'ARCHIVED').length,
  };

  const statusFilterOptions = [
    { key: 'all' as const, label: 'Все', color: 'bg-gray-100 text-gray-700' },
    { key: 'ACTIVE' as const, label: 'Активные', color: 'bg-emerald-100 text-emerald-700' },
    { key: 'ON_HOLD' as const, label: 'На паузе', color: 'bg-amber-100 text-amber-700' },
    { key: 'ARCHIVED' as const, label: 'В архиве', color: 'bg-gray-100 text-gray-600' },
  ];

  return (
    <div className="min-h-screen bg-gradient-to-br from-white via-emerald-50/30 to-white text-gray-900">
      <div className="mx-auto max-w-7xl px-4 py-6">
        {}
        <header className="flex flex-wrap items-center justify-between gap-4 mb-6">
          <div>
            <p className="text-xs font-semibold uppercase tracking-widest text-emerald-500">Управление</p>
            <h1 className="text-2xl font-bold text-gray-900">Проекты</h1>
          </div>
          {canCreate && (
            <Link
              to="/projects/new"
              className="inline-flex items-center gap-2 rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600 transition-colors"
            >
              <PlusCircle className="h-4 w-4" />
              Новый проект
            </Link>
          )}
        </header>

        {}
        <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm mb-6">
          <div className="flex flex-wrap items-center justify-between gap-4">
            {}
            <div className="flex flex-wrap gap-2">
              {statusFilterOptions.map((opt) => (
                <button
                  key={opt.key}
                  type="button"
                  onClick={() => setStatusFilter(opt.key)}
                  className={`inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-sm font-medium transition-all ${
                    statusFilter === opt.key
                      ? opt.color
                      : 'bg-white text-gray-500 hover:bg-gray-50'
                  }`}
                >
                  {opt.label}
                  <span className={`ml-1 rounded-full px-1.5 py-0.5 text-xs ${
                    statusFilter === opt.key ? 'bg-white/30' : 'bg-gray-100'
                  }`}>
                    {statusCounts[opt.key]}
                  </span>
                </button>
              ))}
            </div>

            {}
            <div className="flex items-center gap-3">
              <div className="flex rounded-lg border border-gray-200 p-0.5">
                <button
                  type="button"
                  onClick={() => setViewMode('grid')}
                  className={`rounded-md p-1.5 transition-colors ${
                    viewMode === 'grid' ? 'bg-emerald-100 text-emerald-600' : 'text-gray-400 hover:text-gray-600'
                  }`}
                >
                  <Grid3X3 className="h-4 w-4" />
                </button>
                <button
                  type="button"
                  onClick={() => setViewMode('list')}
                  className={`rounded-md p-1.5 transition-colors ${
                    viewMode === 'list' ? 'bg-emerald-100 text-emerald-600' : 'text-gray-400 hover:text-gray-600'
                  }`}
                >
                  <LayoutList className="h-4 w-4" />
                </button>
              </div>

              <div className="relative">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                <input
                  type="text"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  placeholder="Поиск проектов..."
                  className="w-64 rounded-xl border border-gray-200 bg-white pl-9 pr-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                />
              </div>
            </div>
          </div>
        </div>

        {}
        {loading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="h-8 w-8 animate-spin text-emerald-500" />
          </div>
        ) : error ? (
          <div className="rounded-2xl border border-red-100 bg-red-50 px-6 py-8 text-center">
            <p className="text-red-700">{error}</p>
            <button
              onClick={() => window.location.reload()}
              className="mt-4 text-sm font-medium text-red-600 hover:text-red-700"
            >
              Попробовать снова
            </button>
          </div>
        ) : filteredProjects.length > 0 ? (
          viewMode === 'grid' ? (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {filteredProjects.map((project) => (
                <Link
                  to={`/projects/${project.id}`}
                  key={project.id}
                  className="group rounded-2xl border border-gray-100 bg-white p-5 shadow-sm hover:border-emerald-200 hover:shadow-lg transition-all"
                >
                  <div className="flex items-start justify-between mb-3">
                    <div className="h-12 w-12 rounded-xl bg-gradient-to-br from-emerald-400 to-lime-400 flex items-center justify-center text-white text-lg font-bold shadow-lg shadow-emerald-200/50">
                      {project.name.charAt(0).toUpperCase()}
                    </div>
                    <span className={`rounded-full px-2.5 py-1 text-xs font-medium ${projectStatusColors[project.status] || 'bg-gray-100 text-gray-600'}`}>
                      {projectStatusLabels[project.status] || project.status}
                    </span>
                  </div>
                  
                  <h3 className="font-semibold text-gray-900 mb-1 group-hover:text-emerald-600 transition-colors">
                    {project.name}
                  </h3>
                  <p className="text-sm text-gray-500 line-clamp-2 mb-4">
                    {project.description || 'Без описания'}
                  </p>
                  
                  <div className="flex items-center justify-between pt-3 border-t border-gray-100">
                    <div className="flex items-center gap-1.5 text-xs text-gray-400">
                      <User className="h-3.5 w-3.5" />
                      <span>{getManagerName(project.manager_id)}</span>
                    </div>
                    <div className="flex items-center gap-1.5 text-xs text-gray-400">
                      <Calendar className="h-3.5 w-3.5" />
                      <span>{new Date(project.created_at).toLocaleDateString('ru-RU')}</span>
                    </div>
                  </div>
                  
                  <div className="mt-3 flex items-center gap-1 text-xs font-medium text-emerald-600 opacity-0 group-hover:opacity-100 transition-opacity">
                    Открыть проект
                    <ArrowRight className="h-3.5 w-3.5" />
                  </div>
                </Link>
              ))}
            </div>
          ) : (
            <div className="rounded-2xl border border-gray-100 bg-white shadow-sm overflow-hidden">
              <table className="w-full">
                <thead className="bg-gray-50 text-xs uppercase text-gray-500">
                  <tr>
                    <th className="px-4 py-3 text-left">Проект</th>
                    <th className="px-4 py-3 text-left hidden md:table-cell">Описание</th>
                    <th className="px-4 py-3 text-left hidden sm:table-cell">Менеджер</th>
                    <th className="px-4 py-3 text-left">Статус</th>
                    <th className="px-4 py-3 text-left hidden sm:table-cell">Создан</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {filteredProjects.map((project) => (
                    <tr key={project.id} className="hover:bg-gray-50/50 transition-colors">
                      <td className="px-4 py-3">
                        <Link 
                          to={`/projects/${project.id}`}
                          className="flex items-center gap-3 group"
                        >
                          <div className="h-9 w-9 rounded-lg bg-gradient-to-br from-emerald-400 to-lime-400 flex items-center justify-center text-white text-sm font-semibold">
                            {project.name.charAt(0).toUpperCase()}
                          </div>
                          <span className="font-medium text-gray-900 group-hover:text-emerald-600 transition-colors">
                            {project.name}
                          </span>
                        </Link>
                      </td>
                      <td className="px-4 py-3 hidden md:table-cell">
                        <p className="text-sm text-gray-500 line-clamp-1 max-w-xs">
                          {project.description || '—'}
                        </p>
                      </td>
                      <td className="px-4 py-3 hidden sm:table-cell text-sm text-gray-600">
                        {getManagerName(project.manager_id)}
                      </td>
                      <td className="px-4 py-3">
                        <span className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${projectStatusColors[project.status] || 'bg-gray-100 text-gray-600'}`}>
                          {projectStatusLabels[project.status] || project.status}
                        </span>
                      </td>
                      <td className="px-4 py-3 hidden sm:table-cell text-sm text-gray-400">
                        {new Date(project.created_at).toLocaleDateString('ru-RU')}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )
        ) : (
          <div className="rounded-2xl border-2 border-dashed border-gray-200 bg-white px-6 py-16 text-center">
            <FolderKanban className="mx-auto h-12 w-12 text-gray-300" />
            <p className="mt-4 text-gray-500">
              {searchTerm || statusFilter !== 'all' 
                ? 'Нет проектов по заданным фильтрам' 
                : 'Проектов пока нет'}
            </p>
            {!searchTerm && statusFilter === 'all' && (
              <Link
                to="/projects/new"
                className="mt-4 inline-flex items-center gap-2 text-sm font-medium text-emerald-600 hover:text-emerald-700"
              >
                <PlusCircle className="h-4 w-4" />
                Создать первый проект
              </Link>
            )}
          </div>
        )}
      </div>
    </div>
  );
}


