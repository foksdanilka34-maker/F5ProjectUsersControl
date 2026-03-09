import { useCallback, useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Building2,
  ChevronLeft,
  ChevronRight,
  Edit2,
  Layers3,
  Loader2,
  Plus,
  Save,
  Search,
  Star,
  Trash2,
  User,
  UserPlus,
  X,
} from 'lucide-react';
import Avatar from '../components/Avatar';
import ConfirmDialog from '../components/ConfirmDialog';
import { useAuth } from '../context/AuthContext';

import {
  listProfiles,
  listDepartments,
  listPositions,
  listSkills,
  createProfile,
  updateProfile,
  createDepartment,
  updateDepartment,
  deleteDepartment,
  createPosition,
  updatePosition,
  deletePosition,
  createSkill,
  deleteSkill,
} from '../services/employeeService';
import type { ProfileDTO, DepartmentDTO, PositionDTO, SkillDTO, CreateProfileRequest } from '../services/types';

const PAGE_SIZE = 10;

const tabs = ['Профили', 'Оргструктура', 'Навыки'] as const;
type TabType = typeof tabs[number];

const statusStyles: Record<string, string> = {
  active: 'bg-emerald-50 text-emerald-700',
  blocked: 'bg-rose-50 text-rose-700',
};

const roles = [
  { label: 'Администратор', value: 'admin' },
  { label: 'Директор', value: 'director' },
  { label: 'Менеджер', value: 'manager' },
  { label: 'Сотрудник', value: 'employee' },
];

type FormState = {
  firstName: string;
  lastName: string;
  email: string;
  departmentId: string;
  positionId: string;
  hireDate: string;
  login: string;
  password: string;
  role: string;
};

const emptyForm: FormState = {
  firstName: '',
  lastName: '',
  email: '',
  departmentId: '',
  positionId: '',
  hireDate: new Date().toISOString().split('T')[0],
  login: '',
  password: '',
  role: 'employee',
};

export default function EmployeesHubPage() {
  const [activeTab, setActiveTab] = useState<TabType>('Профили');
  const [searchTerm, setSearchTerm] = useState('');
  const [currentPage, setCurrentPage] = useState(1);

  const [profiles, setProfiles] = useState<ProfileDTO[]>([]);
  const [departments, setDepartments] = useState<DepartmentDTO[]>([]);
  const [positions, setPositions] = useState<PositionDTO[]>([]);
  const [skills, setSkills] = useState<SkillDTO[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [loading, setLoading] = useState(true);

  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const { user } = useAuth();
  const isAdmin = user?.role === 'admin' || user?.role === 'developer';

  const [confirmDialog, setConfirmDialog] = useState<{
    isOpen: boolean;
    title: string;
    message: string;
    onConfirm: () => void;
  }>({ isOpen: false, title: '', message: '', onConfirm: () => {} });

  const closeConfirm = () => setConfirmDialog(prev => ({ ...prev, isOpen: false }));

  const [editingDept, setEditingDept] = useState<{ id: number; name: string } | null>(null);
  const [newDeptName, setNewDeptName] = useState('');
  const [editingPos, setEditingPos] = useState<{ id: number; name: string } | null>(null);
  const [newPosName, setNewPosName] = useState('');
  const [newSkillName, setNewSkillName] = useState('');

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [profilesRes, depsRes, posRes, skillsRes] = await Promise.all([
        listProfiles({ pageSize: PAGE_SIZE, pageNumber: currentPage }),
        listDepartments(),
        listPositions(),
        listSkills(),
      ]);
      setProfiles(profilesRes.profiles || []);
      setTotalCount(profilesRes.total_count || 0);
      setDepartments(depsRes.departments || []);
      setPositions(posRes.positions || []);
      setSkills(skillsRes.skills || []);
    } catch (err) {
      console.error('Failed to load data:', err);
    } finally {
      setLoading(false);
    }
  }, [currentPage]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const filteredProfiles = useMemo(() => {
    if (!searchTerm.trim()) return profiles;
    const normalized = searchTerm.toLowerCase();
    return profiles.filter((profile) =>
      `${profile.first_name} ${profile.last_name}`.toLowerCase().includes(normalized) ||
      profile.email?.toLowerCase().includes(normalized)
    );
  }, [profiles, searchTerm]);

  const totalPages = Math.ceil(totalCount / PAGE_SIZE);

  const handleFormChange = (field: keyof FormState) => (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    setForm(prev => ({ ...prev, [field]: e.target.value }));
    setError(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    setSuccess(null);

    try {
      if (editingId) {

        await updateProfile(editingId, {
          first_name: form.firstName.trim(),
          last_name: form.lastName.trim(),
          email: form.email.trim(),
          position_id: Number(form.positionId),
          department_id: form.departmentId ? Number(form.departmentId) : undefined,
        });
        setSuccess('Профиль успешно обновлён');
      } else {

        const payload: CreateProfileRequest = {
          first_name: form.firstName.trim(),
          last_name: form.lastName.trim(),
          email: form.email.trim(),
          position_id: Number(form.positionId),
          hire_date: new Date(form.hireDate).toISOString(),
          login: form.login.trim(),
          password: form.password,
          role: form.role,
          ...(form.departmentId ? { department_id: Number(form.departmentId) } : {}),
        };
        await createProfile(payload);
        setSuccess(`Сотрудник ${form.firstName} ${form.lastName} создан`);
      }
      setForm(emptyForm);
      setShowForm(false);
      setEditingId(null);
      loadData();
    } catch (err) {
      console.error('Submit error:', err);
      setError('Не удалось сохранить. Проверьте данные.');
    } finally {
      setSubmitting(false);
    }
  };

  const handleCreateDepartment = async () => {
    if (!newDeptName.trim()) return;
    setSubmitting(true);
    try {
      await createDepartment(newDeptName.trim());
      setNewDeptName('');
      setSuccess('Отдел создан');
      loadData();
    } catch {
      setError('Не удалось создать отдел');
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdateDepartment = async () => {
    if (!editingDept || !editingDept.name.trim()) return;
    setSubmitting(true);
    try {
      await updateDepartment(editingDept.id, editingDept.name.trim());
      setEditingDept(null);
      setSuccess('Отдел обновлён');
      loadData();
    } catch {
      setError('Не удалось обновить отдел');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteDepartment = async (id: number) => {
    setConfirmDialog({
      isOpen: true,
      title: 'Удалить отдел',
      message: 'Вы уверены, что хотите удалить этот отдел? Это действие нельзя отменить.',
      onConfirm: async () => {
        closeConfirm();
        try {
          await deleteDepartment(id);
          setSuccess('Отдел удалён');
          loadData();
        } catch {
          setError('Не удалось удалить отдел');
        }
      },
    });
  };

  const handleCreatePosition = async () => {
    if (!newPosName.trim()) return;
    setSubmitting(true);
    try {
      await createPosition(newPosName.trim());
      setNewPosName('');
      setSuccess('Должность создана');
      loadData();
    } catch {
      setError('Не удалось создать должность');
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdatePosition = async () => {
    if (!editingPos || !editingPos.name.trim()) return;
    setSubmitting(true);
    try {
      await updatePosition(editingPos.id, editingPos.name.trim());
      setEditingPos(null);
      setSuccess('Должность обновлена');
      loadData();
    } catch {
      setError('Не удалось обновить должность');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeletePosition = async (id: number) => {
    setConfirmDialog({
      isOpen: true,
      title: 'Удалить должность',
      message: 'Вы уверены, что хотите удалить эту должность? Это действие нельзя отменить.',
      onConfirm: async () => {
        closeConfirm();
        try {
          await deletePosition(id);
          setSuccess('Должность удалена');
          loadData();
        } catch {
          setError('Не удалось удалить должность');
        }
      },
    });
  };

  const handleCreateSkill = async () => {
    if (!newSkillName.trim()) return;
    setSubmitting(true);
    try {
      await createSkill(newSkillName.trim());
      setNewSkillName('');
      setSuccess('Навык создан');
      loadData();
    } catch {
      setError('Не удалось создать навык');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteSkill = async (skillId: number, skillName: string) => {
    setConfirmDialog({
      isOpen: true,
      title: 'Удалить навык',
      message: `Вы уверены, что хотите удалить навык «${skillName}»?`,
      onConfirm: async () => {
        closeConfirm();
        setSubmitting(true);
        try {
          await deleteSkill(skillId);
          setSuccess('Навык удалён');
          loadData();
        } catch {
          setError('Не удалось удалить навык');
        } finally {
          setSubmitting(false);
        }
      },
    });
  };

  const getPositionName = (positionId: number) => 
    positions.find(p => p.id === positionId)?.name || '—';

  const openCreateForm = () => {
    setForm(emptyForm);
    setEditingId(null);
    setShowForm(true);
    setError(null);
    setSuccess(null);
  };

  const openEditForm = (profile: ProfileDTO) => {
    setForm({
      firstName: profile.first_name,
      lastName: profile.last_name,
      email: profile.email,
      departmentId: profile.department?.id?.toString() || '',
      positionId: profile.position_id.toString(),
      hireDate: profile.hire_date.split('T')[0],
      login: profile.login,
      password: '',
      role: profile.role,
    });
    setEditingId(profile.id);
    setShowForm(true);
    setError(null);
    setSuccess(null);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-white via-emerald-50/30 to-white text-gray-900">
      <div className="mx-auto max-w-7xl px-4 py-6">
        {}
        <header className="flex flex-wrap items-center justify-between gap-4 mb-6">
          <div>
            <p className="text-xs font-semibold uppercase tracking-widest text-emerald-500">Управление</p>
            <h1 className="text-2xl font-bold text-gray-900">Сотрудники</h1>
          </div>
        </header>

        {}
        {success && (
          <div className="mb-4 rounded-xl border border-emerald-100 bg-emerald-50 px-4 py-2 text-sm text-emerald-700">
            {success}
          </div>
        )}
        {error && (
          <div className="mb-4 rounded-xl border border-red-100 bg-red-50 px-4 py-2 text-sm text-red-700">
            {error}
          </div>
        )}

        {}
        <div className="flex flex-wrap gap-2 mb-6">
          {tabs.map((tab) => (
            <button
              key={tab}
              type="button"
              onClick={() => setActiveTab(tab)}
              className={`rounded-full px-4 py-2 text-sm font-medium transition-all ${
                activeTab === tab
                  ? 'bg-gray-900 text-white'
                  : 'border border-gray-200 bg-white text-gray-600 hover:text-emerald-600'
              }`}
            >
              {tab}
            </button>
          ))}
        </div>

        <div className="grid gap-6 lg:grid-cols-3">
          {}
          <div className="lg:col-span-2">
            {activeTab === 'Профили' && (
              <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm">
                {}
                <div className="flex flex-wrap items-center justify-between gap-3 mb-4">
                  <div className="relative flex-1 min-w-[200px] max-w-sm">
                    <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                    <input
                      type="text"
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      placeholder="Поиск по имени или email..."
                      className="w-full rounded-xl border border-gray-200 bg-white pl-9 pr-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                  </div>
                  <button
                    type="button"
                    onClick={openCreateForm}
                    className="inline-flex items-center gap-2 rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600"
                  >
                    <UserPlus className="h-4 w-4" />
                    Добавить
                  </button>
                </div>

                {}
                <div className="overflow-hidden rounded-xl border border-gray-100">
                  {loading ? (
                    <div className="flex items-center justify-center py-12">
                      <Loader2 className="h-6 w-6 animate-spin text-emerald-500" />
                    </div>
                  ) : filteredProfiles.length ? (
                    <table className="w-full text-left text-sm">
                      <thead className="bg-gray-50 text-xs uppercase text-gray-500">
                        <tr>
                          <th className="px-4 py-3">Сотрудник</th>
                          <th className="px-4 py-3 hidden sm:table-cell">Должность</th>
                          <th className="px-4 py-3 hidden md:table-cell">Отдел</th>
                          <th className="px-4 py-3">Статус</th>
                          <th className="px-4 py-3 w-24">Действия</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-100">
                        {filteredProfiles.map((profile) => (
                          <tr key={profile.id} className="hover:bg-gray-50/50">
                            <td className="px-4 py-3">
                              <div className="flex items-center gap-3">
                                <Avatar
                                  src={profile.avatar_url}
                                  name={`${profile.first_name} ${profile.last_name}`}
                                  size="sm"
                                />
                                <div>
                                  <p className="font-medium text-gray-900">
                                    {profile.first_name} {profile.last_name}
                                  </p>
                                  <p className="text-xs text-gray-400">{profile.email}</p>
                                </div>
                              </div>
                            </td>
                            <td className="px-4 py-3 hidden sm:table-cell text-gray-600">
                              {getPositionName(profile.position_id)}
                            </td>
                            <td className="px-4 py-3 hidden md:table-cell text-gray-600">
                              {profile.department?.name || '—'}
                            </td>
                            <td className="px-4 py-3">
                              <span className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${
                                profile.is_active ? statusStyles.active : statusStyles.blocked
                              }`}>
                                {profile.is_active ? 'Активен' : 'Заблокирован'}
                              </span>
                            </td>
                            <td className="px-4 py-3">
                              <div className="flex items-center gap-1">
                                <Link
                                  to={`/profile/${profile.id}`}
                                  className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-emerald-600"
                                  title="Открыть профиль"
                                >
                                  <User className="h-4 w-4" />
                                </Link>
                                <button
                                  type="button"
                                  onClick={() => openEditForm(profile)}
                                  className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-emerald-600"
                                  title="Редактировать"
                                >
                                  <Edit2 className="h-4 w-4" />
                                </button>
                              </div>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  ) : (
                    <div className="py-12 text-center text-sm text-gray-500">
                      {searchTerm ? `Нет результатов для «${searchTerm}»` : 'Сотрудников пока нет'}
                    </div>
                  )}
                </div>

                {}
                {totalPages > 1 && (
                  <div className="flex items-center justify-between mt-4 text-sm">
                    <span className="text-gray-500">
                      Страница {currentPage} из {totalPages} ({totalCount} записей)
                    </span>
                    <div className="flex gap-1">
                      <button
                        type="button"
                        onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                        disabled={currentPage === 1}
                        className="rounded-lg border border-gray-200 p-2 hover:bg-gray-50 disabled:opacity-50"
                      >
                        <ChevronLeft className="h-4 w-4" />
                      </button>
                      <button
                        type="button"
                        onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                        disabled={currentPage === totalPages}
                        className="rounded-lg border border-gray-200 p-2 hover:bg-gray-50 disabled:opacity-50"
                      >
                        <ChevronRight className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}

            {activeTab === 'Оргструктура' && (
              <div className="grid gap-4 sm:grid-cols-2">
                {}
                <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <Building2 className="h-4 w-4 text-emerald-500" />
                      <h3 className="font-semibold text-gray-900">Отделы</h3>
                    </div>
                    <span className="text-xs text-gray-400">{departments.length}</span>
                  </div>

                  {isAdmin && (
                    <div className="flex gap-2 mb-3">
                      <input
                        type="text"
                        value={newDeptName}
                        onChange={(e) => setNewDeptName(e.target.value)}
                        placeholder="Название отдела"
                        className="flex-1 rounded-lg border border-gray-200 px-3 py-1.5 text-sm focus:border-emerald-400 focus:outline-none"
                        onKeyDown={(e) => e.key === 'Enter' && handleCreateDepartment()}
                      />
                      <button
                        type="button"
                        onClick={handleCreateDepartment}
                        disabled={!newDeptName.trim() || submitting}
                        className="rounded-lg bg-emerald-500 px-3 py-1.5 text-white hover:bg-emerald-600 disabled:opacity-50"
                      >
                        <Plus className="h-4 w-4" />
                      </button>
                    </div>
                  )}

                  <div className="space-y-2 max-h-[300px] overflow-y-auto">
                    {departments.length ? departments.map((dep) => (
                      <div key={dep.id} className="flex items-center justify-between rounded-lg bg-gray-50 px-3 py-2 text-sm group">
                        {editingDept?.id === dep.id ? (
                          <input
                            type="text"
                            value={editingDept.name}
                            onChange={(e) => setEditingDept({ ...editingDept, name: e.target.value })}
                            className="flex-1 rounded border border-emerald-300 px-2 py-1 text-sm focus:outline-none"
                            autoFocus
                            onKeyDown={(e) => {
                              if (e.key === 'Enter') handleUpdateDepartment();
                              if (e.key === 'Escape') setEditingDept(null);
                            }}
                          />
                        ) : (
                          <span className="font-medium text-gray-700">{dep.name}</span>
                        )}
                        {isAdmin && (
                          <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                            {editingDept?.id === dep.id ? (
                              <>
                                <button
                                  onClick={handleUpdateDepartment}
                                  className="p-1 text-emerald-600 hover:bg-emerald-100 rounded"
                                >
                                  <Save className="h-3.5 w-3.5" />
                                </button>
                                <button
                                  onClick={() => setEditingDept(null)}
                                  className="p-1 text-gray-400 hover:bg-gray-200 rounded"
                                >
                                  <X className="h-3.5 w-3.5" />
                                </button>
                              </>
                            ) : (
                              <>
                                <button
                                  onClick={() => setEditingDept({ id: dep.id, name: dep.name })}
                                  className="p-1 text-gray-400 hover:bg-gray-200 rounded"
                                >
                                  <Edit2 className="h-3.5 w-3.5" />
                                </button>
                                <button
                                  onClick={() => handleDeleteDepartment(dep.id)}
                                  className="p-1 text-red-400 hover:bg-red-100 rounded"
                                >
                                  <Trash2 className="h-3.5 w-3.5" />
                                </button>
                              </>
                            )}
                          </div>
                        )}
                      </div>
                    )) : (
                      <p className="text-sm text-gray-400 py-4 text-center">Отделов пока нет</p>
                    )}
                  </div>
                </div>

                {}
                <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <Layers3 className="h-4 w-4 text-emerald-500" />
                      <h3 className="font-semibold text-gray-900">Должности</h3>
                    </div>
                    <span className="text-xs text-gray-400">{positions.length}</span>
                  </div>

                  {isAdmin && (
                    <div className="flex gap-2 mb-3">
                      <input
                        type="text"
                        value={newPosName}
                        onChange={(e) => setNewPosName(e.target.value)}
                        placeholder="Название должности"
                        className="flex-1 rounded-lg border border-gray-200 px-3 py-1.5 text-sm focus:border-emerald-400 focus:outline-none"
                        onKeyDown={(e) => e.key === 'Enter' && handleCreatePosition()}
                      />
                      <button
                        type="button"
                        onClick={handleCreatePosition}
                        disabled={!newPosName.trim() || submitting}
                        className="rounded-lg bg-emerald-500 px-3 py-1.5 text-white hover:bg-emerald-600 disabled:opacity-50"
                      >
                        <Plus className="h-4 w-4" />
                      </button>
                    </div>
                  )}

                  <div className="space-y-2 max-h-[300px] overflow-y-auto">
                    {positions.length ? positions.map((pos) => (
                      <div key={pos.id} className="flex items-center justify-between rounded-lg bg-gray-50 px-3 py-2 text-sm group">
                        {editingPos?.id === pos.id ? (
                          <input
                            type="text"
                            value={editingPos.name}
                            onChange={(e) => setEditingPos({ ...editingPos, name: e.target.value })}
                            className="flex-1 rounded border border-emerald-300 px-2 py-1 text-sm focus:outline-none"
                            autoFocus
                            onKeyDown={(e) => {
                              if (e.key === 'Enter') handleUpdatePosition();
                              if (e.key === 'Escape') setEditingPos(null);
                            }}
                          />
                        ) : (
                          <span className="font-medium text-gray-700">{pos.name}</span>
                        )}
                        {isAdmin && (
                          <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                            {editingPos?.id === pos.id ? (
                              <>
                                <button
                                  onClick={handleUpdatePosition}
                                  className="p-1 text-emerald-600 hover:bg-emerald-100 rounded"
                                >
                                  <Save className="h-3.5 w-3.5" />
                                </button>
                                <button
                                  onClick={() => setEditingPos(null)}
                                  className="p-1 text-gray-400 hover:bg-gray-200 rounded"
                                >
                                  <X className="h-3.5 w-3.5" />
                                </button>
                              </>
                            ) : (
                              <>
                                <button
                                  onClick={() => setEditingPos({ id: pos.id, name: pos.name })}
                                  className="p-1 text-gray-400 hover:bg-gray-200 rounded"
                                >
                                  <Edit2 className="h-3.5 w-3.5" />
                                </button>
                                <button
                                  onClick={() => handleDeletePosition(pos.id)}
                                  className="p-1 text-red-400 hover:bg-red-100 rounded"
                                >
                                  <Trash2 className="h-3.5 w-3.5" />
                                </button>
                              </>
                            )}
                          </div>
                        )}
                      </div>
                    )) : (
                      <p className="text-sm text-gray-400 py-4 text-center">Должностей пока нет</p>
                    )}
                  </div>
                </div>
              </div>
            )}

            {activeTab === 'Навыки' && (
              <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <Star className="h-4 w-4 text-emerald-500" />
                    <h3 className="font-semibold text-gray-900">Библиотека навыков</h3>
                  </div>
                  <span className="text-xs text-gray-400">{skills.length} навыков</span>
                </div>

                {isAdmin && (
                  <div className="flex gap-2 mb-4">
                    <input
                      type="text"
                      value={newSkillName}
                      onChange={(e) => setNewSkillName(e.target.value)}
                      placeholder="Название навыка"
                      className="flex-1 max-w-xs rounded-lg border border-gray-200 px-3 py-1.5 text-sm focus:border-emerald-400 focus:outline-none"
                      onKeyDown={(e) => e.key === 'Enter' && handleCreateSkill()}
                    />
                    <button
                      type="button"
                      onClick={handleCreateSkill}
                      disabled={!newSkillName.trim() || submitting}
                      className="rounded-lg bg-emerald-500 px-3 py-1.5 text-white hover:bg-emerald-600 disabled:opacity-50"
                    >
                      <Plus className="h-4 w-4" />
                    </button>
                  </div>
                )}

                <div className="flex flex-wrap gap-2 max-h-[300px] overflow-y-auto">
                  {skills.length ? skills.map((skill) => (
                    <span key={skill.id} className="group inline-flex items-center gap-1 rounded-full bg-emerald-50 px-3 py-1 text-sm font-medium text-emerald-700">
                      {skill.name}
                      {isAdmin && (
                        <button
                          type="button"
                          onClick={() => handleDeleteSkill(skill.id, skill.name)}
                          disabled={submitting}
                          className="ml-1 rounded-full p-0.5 text-emerald-400 hover:bg-emerald-100 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity"
                          title="Удалить навык"
                        >
                          <X className="h-3 w-3" />
                        </button>
                      )}
                    </span>
                  )) : (
                    <p className="text-sm text-gray-400 py-4 w-full text-center">Навыков пока нет</p>
                  )}
                </div>
              </div>
            )}
          </div>

          {}
          <div className="lg:col-span-1">
            {showForm ? (
              <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm sticky top-6">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="font-semibold text-gray-900">
                    {editingId ? 'Редактировать' : 'Новый сотрудник'}
                  </h3>
                  <button
                    type="button"
                    onClick={() => setShowForm(false)}
                    className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100"
                  >
                    <X className="h-4 w-4" />
                  </button>
                </div>

                <form onSubmit={handleSubmit} className="space-y-3">
                  <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-1">
                    <div>
                      <label className="text-xs font-medium text-gray-500">Имя *</label>
                      <input
                        value={form.firstName}
                        onChange={handleFormChange('firstName')}
                        placeholder="Иван"
                        required
                        className="mt-1 w-full rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                      />
                    </div>
                    <div>
                      <label className="text-xs font-medium text-gray-500">Фамилия *</label>
                      <input
                        value={form.lastName}
                        onChange={handleFormChange('lastName')}
                        placeholder="Петров"
                        required
                        className="mt-1 w-full rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="text-xs font-medium text-gray-500">Email *</label>
                    <input
                      type="email"
                      value={form.email}
                      onChange={handleFormChange('email')}
                      placeholder="email@company.ru"
                      required
                      className="mt-1 w-full rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                  </div>

                  <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-1">
                    <div>
                      <label className="text-xs font-medium text-gray-500">Отдел</label>
                      <select
                        value={form.departmentId}
                        onChange={handleFormChange('departmentId')}
                        className="mt-1 w-full rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                      >
                        <option value="">Не выбран</option>
                        {departments.map((d) => (
                          <option key={d.id} value={d.id}>{d.name}</option>
                        ))}
                      </select>
                    </div>
                    <div>
                      <label className="text-xs font-medium text-gray-500">Должность *</label>
                      <select
                        value={form.positionId}
                        onChange={handleFormChange('positionId')}
                        required
                        className="mt-1 w-full rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                      >
                        <option value="">Выберите</option>
                        {positions.map((p) => (
                          <option key={p.id} value={p.id}>{p.name}</option>
                        ))}
                      </select>
                    </div>
                  </div>

                  <div>
                    <label className="text-xs font-medium text-gray-500">Дата приёма *</label>
                    <input
                      type="date"
                      value={form.hireDate}
                      onChange={handleFormChange('hireDate')}
                      required
                      className="mt-1 w-full rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                  </div>

                  <hr className="border-gray-100" />

                  <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-1">
                    <div>
                      <label className="text-xs font-medium text-gray-500">Логин *</label>
                      <input
                        value={form.login}
                        onChange={handleFormChange('login')}
                        placeholder="ivanov"
                        required
                        className="mt-1 w-full rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                      />
                    </div>
                    <div>
                      <label className="text-xs font-medium text-gray-500">Роль *</label>
                      <select
                        value={form.role}
                        onChange={handleFormChange('role')}
                        required
                        className="mt-1 w-full rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                      >
                        {roles.map((r) => (
                          <option key={r.value} value={r.value}>{r.label}</option>
                        ))}
                      </select>
                    </div>
                  </div>

                  <div>
                    <label className="text-xs font-medium text-gray-500">
                      Пароль {editingId ? '(оставьте пустым, чтобы не менять)' : '*'}
                    </label>
                    <input
                      type="password"
                      value={form.password}
                      onChange={handleFormChange('password')}
                      placeholder="••••••••"
                      required={!editingId}
                      minLength={6}
                      className="mt-1 w-full rounded-xl border border-gray-200 px-3 py-2 text-sm focus:border-emerald-400 focus:outline-none"
                    />
                  </div>

                  <button
                    type="submit"
                    disabled={submitting}
                    className="w-full inline-flex items-center justify-center gap-2 rounded-xl bg-emerald-500 px-4 py-2.5 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-60"
                  >
                    {submitting ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Save className="h-4 w-4" />
                    )}
                    {editingId ? 'Сохранить изменения' : 'Создать сотрудника'}
                  </button>
                </form>
              </div>
            ) : (
              <div className="rounded-2xl border border-dashed border-gray-200 bg-gray-50/50 p-6 text-center">
                <User className="mx-auto h-8 w-8 text-gray-300" />
                <p className="mt-2 text-sm text-gray-500">
                  Нажмите «Добавить», чтобы создать нового сотрудника
                </p>
                <button
                  type="button"
                  onClick={openCreateForm}
                  className="mt-4 inline-flex items-center gap-2 rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600"
                >
                  <Plus className="h-4 w-4" />
                  Добавить сотрудника
                </button>
              </div>
            )}

            {}
            <div className="mt-4 rounded-2xl border border-gray-100 bg-white p-4 shadow-sm">
              <h4 className="text-xs font-semibold uppercase tracking-wide text-gray-400 mb-3">Статистика</h4>
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Всего сотрудников</span>
                  <span className="font-semibold text-gray-900">{totalCount}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Отделов</span>
                  <span className="font-semibold text-gray-900">{departments.length}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Должностей</span>
                  <span className="font-semibold text-gray-900">{positions.length}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Навыков</span>
                  <span className="font-semibold text-gray-900">{skills.length}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <ConfirmDialog
        isOpen={confirmDialog.isOpen}
        title={confirmDialog.title}
        message={confirmDialog.message}
        confirmLabel="Удалить"
        variant="danger"
        onConfirm={confirmDialog.onConfirm}
        onCancel={closeConfirm}
      />
    </div>
  );
}


