import { useMemo, useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import {
  ArrowLeft,
  BadgeCheck,
  Building2,
  Edit,
  Layers3,
  ListChecks,
  Search,
  UserPlus,
  Loader2,
  Trash2,
  CheckCircle,
  Ban
} from 'lucide-react';
import { employeeService } from '../api/services/employee.service';
import type { Profile, Department, Position, Skill, UpdateProfileRequest } from '../api/types';
import Modal from '../components/Modal';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';

const tabs = ['Профили', 'Оргструктура', 'Навыки'];

export default function EmployeesHubPage() {
  const { user } = useAuth();
  const { showToast } = useToast();
  const [activeTab, setActiveTab] = useState('Профили');
  const [searchTerm, setSearchTerm] = useState('');
  const [isLoading, setIsLoading] = useState(true);

  // Modals state
  const [isDeptModalOpen, setIsDeptModalOpen] = useState(false);
  const [isPosModalOpen, setIsPosModalOpen] = useState(false);
  const [isSkillModalOpen, setIsSkillModalOpen] = useState(false);
  const [isEditProfileModalOpen, setIsEditProfileModalOpen] = useState(false);
  const [isManageSkillsModalOpen, setIsManageSkillsModalOpen] = useState(false);
  const [selectedProfileForSkills, setSelectedProfileForSkills] = useState<Profile | null>(null);
  const [skillIdToAdd, setSkillIdToAdd] = useState<string>("");
  
  const [newItemName, setNewItemName] = useState('');
  const [editingItem, setEditingItem] = useState<Department | Position | Skill | null>(null);
  const [editingProfile, setEditingProfile] = useState<Profile | null>(null);
  const [profileForm, setProfileForm] = useState<UpdateProfileRequest>({});
  
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Data state
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [totalProfilesCount, setTotalProfilesCount] = useState(0);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [skills, setSkills] = useState<Skill[]>([]);

  const fetchData = async () => {
    // Don't set loading to true here to avoid flickering on updates
    // or handle it more gracefully if needed
    try {
      const [profilesData, depsData, posData, skillsData] = await Promise.all([
        employeeService.listProfiles({ page_size: 100 }).catch(err => {
          console.error('Error fetching profiles:', err);
          return { profiles: [], meta: { total_count: 0 } };
        }),
        employeeService.listDepartments().catch(err => {
          console.error('Error fetching departments:', err);
          return [];
        }),
        employeeService.listPositions().catch(err => {
          console.error('Error fetching positions:', err);
          return [];
        }),
        employeeService.listSkills().catch(err => {
          console.error('Error fetching skills:', err);
          return [];
        })
      ]);

      setProfiles(Array.isArray(profilesData?.profiles) ? profilesData.profiles : []);
      setTotalProfilesCount(profilesData?.meta?.total_count || (Array.isArray(profilesData?.profiles) ? profilesData.profiles.length : 0));
      setDepartments(Array.isArray(depsData) ? depsData : []);
      setPositions(Array.isArray(posData) ? posData : []);
      setSkills(Array.isArray(skillsData) ? skillsData : []);
    } catch (error) {
      console.error('Failed to fetch employees data:', error);
    }
  };

  useEffect(() => {
    const init = async () => {
      setIsLoading(true);
      await fetchData();
      setIsLoading(false);
    };
    init();
  }, []);

  const openModal = (type: 'dept' | 'pos' | 'skill', item?: Department | Position | Skill) => {
    setNewItemName(item ? item.name : '');
    setEditingItem(item || null);
    
    if (type === 'dept') setIsDeptModalOpen(true);
    if (type === 'pos') setIsPosModalOpen(true);
    if (type === 'skill') setIsSkillModalOpen(true);
  };

  const openEditProfileModal = (profile: Profile) => {
    setEditingProfile(profile);
    setProfileForm({
      first_name: profile.first_name,
      last_name: profile.last_name,
      email: profile.email,
      position_id: profile.position_id,
      department_id: profile.department_id,
      avatar_url: profile.avatar_url
    });
    setIsEditProfileModalOpen(true);
  };

  const profileStats = useMemo(
    () => [
      { label: 'Профилей в базе', value: totalProfilesCount },
      { label: 'Активные', value: profiles.filter((p) => p.is_active !== false).length },
      { label: 'Отделов', value: departments.length },
      { label: 'Навыков', value: skills.length },
    ],
    [totalProfilesCount, profiles, departments, skills],
  );

  const filteredProfiles = useMemo(() => {
    if (!searchTerm.trim()) {
      return profiles;
    }
    const normalized = searchTerm.toLowerCase();
    return profiles.filter((profile) => {
      const fullName = `${profile.first_name} ${profile.last_name}`.toLowerCase();
      
      const deptName = profile.department?.name || 
                       (profile.department?.id ? departments.find(d => d.id === profile.department!.id)?.name : '') || 
                       (profile.department_id ? departments.find(d => d.id === profile.department_id)?.name : '') || 
                       '';
                       
      const posName = profile.position?.name || 
                      (profile.position_id && positions.find(p => p.id === profile.position_id)?.name) || 
                      '';
      
      return (
        fullName.includes(normalized) ||
        deptName.toLowerCase().includes(normalized) ||
        posName.toLowerCase().includes(normalized) ||
        profile.email.toLowerCase().includes(normalized)
      );
    });
  }, [searchTerm, profiles, departments, positions]);

  const handleCreateOrUpdateDepartment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newItemName.trim()) return;

    setIsSubmitting(true);
    try {
      if (editingItem) {
        await employeeService.updateDepartment(editingItem.id, { name: newItemName });
        showToast('Отдел обновлен', 'success');
      } else {
        await employeeService.createDepartment({ name: newItemName });
        showToast('Отдел создан', 'success');
      }
      await fetchData();
      setNewItemName('');
      setEditingItem(null);
      setIsDeptModalOpen(false);
    } catch (error) {
      console.error('Failed to save department:', error);
      showToast('Не удалось сохранить отдел', 'error');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteDepartment = async (id: string) => {
    if (!window.confirm('Вы уверены? Это действие нельзя отменить.')) return;
    try {
      await employeeService.deleteDepartment(id);
      await fetchData();
      showToast('Отдел удален', 'success');
    } catch (error) {
      console.error('Failed to delete department:', error);
      showToast('Не удалось удалить отдел', 'error');
    }
  };

  const handleCreateOrUpdatePosition = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newItemName.trim()) return;

    setIsSubmitting(true);
    try {
      if (editingItem) {
        await employeeService.updatePosition(editingItem.id, { name: newItemName });
        showToast('Должность обновлена', 'success');
      } else {
        await employeeService.createPosition({ name: newItemName });
        showToast('Должность создана', 'success');
      }
      await fetchData();
      setNewItemName('');
      setEditingItem(null);
      setIsPosModalOpen(false);
    } catch (error) {
      console.error('Failed to save position:', error);
      showToast('Не удалось сохранить должность', 'error');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeletePosition = async (id: string) => {
    if (!window.confirm('Вы уверены? Это действие нельзя отменить.')) return;
    try {
      await employeeService.deletePosition(id);
      await fetchData();
      showToast('Должность удалена', 'success');
    } catch (error) {
      console.error('Failed to delete position:', error);
      showToast('Не удалось удалить должность', 'error');
    }
  };

  const handleCreateOrUpdateSkill = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newItemName.trim()) return;

    setIsSubmitting(true);
    try {
      if (editingItem) {
        await employeeService.updateSkill(editingItem.id, { name: newItemName });
        showToast('Навык обновлен', 'success');
      } else {
        await employeeService.createSkill({ name: newItemName });
        showToast('Навык создан', 'success');
      }
      await fetchData();
      setNewItemName('');
      setEditingItem(null);
      setIsSkillModalOpen(false);
    } catch (error) {
      console.error('Failed to save skill:', error);
      showToast('Не удалось сохранить навык', 'error');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteSkill = async (id: string) => {
    if (!window.confirm('Вы уверены? Это действие нельзя отменить.')) return;
    try {
      await employeeService.deleteSkill(id);
      await fetchData();
      showToast('Навык удален', 'success');
    } catch (error) {
      console.error('Failed to delete skill:', error);
      showToast('Не удалось удалить навык', 'error');
    }
  };

  const handleDeleteProfile = async (profile: Profile) => {
    if (!window.confirm(`Вы уверены, что хотите удалить профиль ${profile.first_name} ${profile.last_name}?`)) return;
    try {
      await employeeService.deleteProfile(profile.id);
      await fetchData();
      showToast('Профиль удален', 'success');
    } catch (error) {
      console.error('Failed to delete profile:', error);
      if (window.confirm('Удаление не поддерживается сервером. Деактивировать пользователя?')) {
         try {
            await employeeService.changeUserStatus(profile.id, { status: false });
            await fetchData();
            showToast('Пользователь деактивирован', 'success');
         } catch (e) {
            showToast('Не удалось деактивировать пользователя', 'error');
         }
      }
    }
  };

  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingProfile) return;

    setIsSubmitting(true);
    try {
      await employeeService.updateProfile(editingProfile.id, profileForm);
      await fetchData();
      showToast('Профиль обновлен', 'success');
      setIsEditProfileModalOpen(false);
      setEditingProfile(null);
    } catch (error) {
      console.error('Failed to update profile:', error);
      showToast('Не удалось обновить профиль', 'error');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleToggleUserStatus = async (profile: Profile) => {
    const newStatus = !(profile.is_active !== false);
    const action = newStatus ? 'активировать' : 'деактивировать';
    
    if (!window.confirm(`Вы уверены, что хотите ${action} сотрудника ${profile.first_name} ${profile.last_name}?`)) return;

    try {
      await employeeService.changeUserStatus(profile.id, { status: newStatus });
      await fetchData();
      showToast(`Сотрудник ${newStatus ? 'активирован' : 'деактивирован'}`, 'success');
    } catch (error) {
      console.error('Failed to change status:', error);
      showToast('Не удалось изменить статус', 'error');
    }
  };

  const openManageSkillsModal = (profile: Profile) => {
    setSelectedProfileForSkills(profile);
    setSkillIdToAdd("");
    setIsManageSkillsModalOpen(true);
  };

  const handleAddSkillToEmployee = async () => {
    if (!selectedProfileForSkills || !skillIdToAdd) return;
    try {
      await employeeService.addSkillToEmployee(selectedProfileForSkills.id, { skill_id: skillIdToAdd });
      showToast("Навык добавлен", "success");
      setSkillIdToAdd("");
      await fetchData();
      // Update local state for immediate feedback
      setSelectedProfileForSkills(prev => {
          if (!prev) return null;
          const skillToAdd = skills.find(s => s.id === skillIdToAdd);
          if (!skillToAdd) return prev;
          return {
              ...prev,
              skills: [...(prev.skills || []), skillToAdd]
          };
      });
    } catch (error) {
      console.error(error);
      showToast("Не удалось добавить навык", "error");
    }
  };

  const handleRemoveSkillFromEmployee = async (skillId: string) => {
     if (!selectedProfileForSkills) return;
     try {
         await employeeService.removeSkillFromEmployee(selectedProfileForSkills.id, skillId);
         showToast("Навык удален", "success");
         await fetchData();
         // Update local state
         setSelectedProfileForSkills(prev => {
             if (!prev) return null;
             return {
                 ...prev,
                 skills: prev.skills?.filter(s => s.id !== skillId)
             };
         });
     } catch (error) {
         console.error(error);
         showToast("Не удалось удалить навык", "error");
     }
  };

  if (isLoading) {
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
            <p className="text-xs font-semibold uppercase tracking-[0.35em] text-emerald-500">Сотрудники</p>
            <h1 className="mt-3 text-3xl font-semibold text-gray-900">Центр управления командами</h1>
            <p className="mt-2 text-sm text-gray-500">
              Управление профилями, отделами, должностями и навыками.
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

        <div className="mt-8 flex flex-wrap gap-3">
          {tabs.map((tab) => (
            <button
              key={tab}
              type="button"
              onClick={() => setActiveTab(tab)}
              className={`rounded-full px-5 py-2 text-sm font-semibold transition-all ${
                activeTab === tab ? 'bg-gray-900 text-white' : 'border border-gray-200 bg-white text-gray-600 hover:text-emerald-600'
              }`}
            >
              {tab}
            </button>
          ))}
        </div>

        {['admin', 'director'].includes(user?.role || '') && (
        <div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {profileStats.map((stat) => (
            <div
              key={stat.label}
              className="rounded-3xl border border-gray-100 bg-white/90 px-4 py-5 text-sm shadow-[0_15px_35px_rgba(6,95,70,0.08)]"
            >
              <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">{stat.label}</p>
              <p className="mt-2 text-2xl font-semibold text-gray-900">{stat.value}</p>
            </div>
          ))}
        </div>
        )}

        <section className="mt-8 grid gap-6 lg:grid-cols-[1fr]">
          <div className="space-y-6">
            {['admin', 'director'].includes(user?.role || '') && (
            <div className="rounded-4xl border border-gray-100 bg-white/95 p-6 shadow-[0_25px_80px_rgba(6,95,70,0.08)]">
              <div className="flex flex-wrap items-center justify-between gap-4">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Быстрые действия</p>
                  <h2 className="mt-2 text-xl font-semibold">Что вы хотите сделать?</h2>
                </div>
              </div>
              <div className="mt-5 flex gap-4">
                  <Link
                    to="/admin/employees/new"
                    className="inline-flex items-center gap-2 rounded-3xl border border-emerald-100 bg-emerald-50/70 px-6 py-4 text-sm font-semibold text-emerald-800 shadow-[0_15px_40px_rgba(16,185,129,0.25)] hover:bg-emerald-100 transition-colors"
                  >
                    <UserPlus className="h-5 w-5" />
                    Добавить профиль
                  </Link>
              </div>
            </div>
            )}

            {activeTab === 'Профили' && (
              <div className="rounded-4xl border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
                <div className="flex flex-wrap items-center justify-between gap-4">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Сотрудники</p>
                    <h2 className="mt-2 text-xl font-semibold">Список профилей</h2>
                  </div>
                  <label className="relative flex items-center">
                    <Search className="absolute left-3 h-4 w-4 text-gray-400" />
                    <input
                      type="text"
                      value={searchTerm}
                      onChange={(event) => setSearchTerm(event.target.value)}
                      placeholder="Поиск сотрудника..."
                      className="w-full min-w-[300px] rounded-full border border-gray-200 bg-white px-9 py-2 text-sm text-gray-600 focus:border-emerald-400 focus:outline-none"
                    />
                    {searchTerm && (
                      <button
                        type="button"
                        onClick={() => setSearchTerm('')}
                        className="absolute right-3 text-xs font-semibold text-emerald-600"
                      >
                        Очистить
                      </button>
                    )}
                  </label>
                </div>
                <div className="mt-5 overflow-hidden rounded-[28px] border border-gray-100">
                  {filteredProfiles.length ? (
                    <table className="w-full text-left text-sm text-gray-600">
                      <thead className="bg-gray-50 text-xs uppercase tracking-wide text-gray-400">
                        <tr>
                          <th className="px-6 py-4">Сотрудник</th>
                          <th className="px-6 py-4">Email</th>
                          <th className="px-6 py-4">Должность</th>
                          <th className="px-6 py-4">Отдел</th>
                          <th className="px-6 py-4">Статус</th>
                          <th className="px-6 py-4">Действия</th>
                        </tr>
                      </thead>
                      <tbody>
                        {filteredProfiles.map((profile) => (
                          <tr key={profile.id} className="border-t border-gray-100 bg-white/80 hover:bg-gray-50 transition-colors">
                            <td className="px-6 py-4 font-semibold text-gray-900">
                              {profile.first_name} {profile.last_name}
                            </td>
                            <td className="px-6 py-4">{profile.email}</td>
                            <td className="px-6 py-4">
                              {profile.position?.name || (profile.position_id && positions.find(p => p.id === profile.position_id)?.name) || '-'}
                            </td>
                            <td className="px-6 py-4">
                              {profile.department?.name || 
                               (profile.department?.id && departments.find(d => d.id === profile.department!.id)?.name) || 
                               (profile.department_id && departments.find(d => d.id === profile.department_id)?.name) || 
                               '-'}
                            </td>
                            <td className="px-6 py-4">
                              <span
                                className={`rounded-full px-3 py-1 text-xs font-semibold ${
                                  profile.is_active !== false
                                    ? 'bg-emerald-50 text-emerald-700'
                                    : 'bg-rose-50 text-rose-700'
                                }`}
                              >
                                {profile.is_active !== false ? 'Активен' : 'Неактивен'}
                              </span>
                            </td>
                            <td className="px-6 py-4">
                              <div className="flex items-center gap-2">
                                <button
                                  onClick={() => openManageSkillsModal(profile)}
                                  className="rounded-full p-2 text-gray-400 hover:bg-blue-50 hover:text-blue-600 transition-colors"
                                  title="Управление навыками"
                                >
                                  <ListChecks className="h-4 w-4" />
                                </button>
                                <button 
                                  onClick={() => openEditProfileModal(profile)}
                                  className="rounded-full p-2 text-gray-400 hover:bg-emerald-50 hover:text-emerald-600 transition-colors"
                                  title="Редактировать"
                                >
                                  <Edit className="h-4 w-4" />
                                </button>
                                <button 
                                  onClick={() => handleToggleUserStatus(profile)}
                                  className={`rounded-full p-2 transition-colors ${
                                    profile.is_active !== false 
                                      ? 'text-gray-400 hover:bg-rose-50 hover:text-rose-600' 
                                      : 'text-gray-400 hover:bg-emerald-50 hover:text-emerald-600'
                                  }`}
                                  title={profile.is_active !== false ? "Деактивировать" : "Активировать"}
                                >
                                  {profile.is_active !== false ? <Ban className="h-4 w-4" /> : <CheckCircle className="h-4 w-4" />}
                                </button>
                                <button 
                                  onClick={() => handleDeleteProfile(profile)}
                                  className="rounded-full p-2 text-gray-400 hover:bg-rose-50 hover:text-rose-600 transition-colors"
                                  title="Удалить"
                                >
                                  <Trash2 className="h-4 w-4" />
                                </button>
                              </div>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  ) : (
                    <div className="flex flex-col items-center justify-center bg-white/80 px-6 py-12 text-center text-sm text-gray-500">
                      <BadgeCheck className="mb-3 h-6 w-6 text-emerald-500" />
                      Нет совпадений по запросу «{searchTerm}».
                    </div>
                  )}
                </div>
              </div>
            )}

            {activeTab === 'Оргструктура' && (
              <div className="grid gap-6 lg:grid-cols-2">
                <div className="rounded-4xl border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Отделы</p>
                      <h2 className="mt-2 text-xl font-semibold">Управление департаментами</h2>
                    </div>
                    <button 
                      onClick={() => openModal('dept')}
                      className="rounded-full border border-gray-200 px-4 py-2 text-xs font-semibold text-gray-600 hover:bg-gray-50" 
                      type="button"
                    >
                      <Building2 className="mr-1 inline h-3.5 w-3.5" /> Добавить
                    </button>
                  </div>
                  <ul className="mt-5 space-y-3 text-sm text-gray-600">
                    {departments.map((dep) => (
                      <li key={dep.id} className="rounded-3xl border border-gray-100 bg-gray-50/80 px-4 py-3 group hover:bg-gray-50 transition-colors">
                        <div className="flex items-center justify-between">
                          <p className="font-semibold text-gray-900">{dep.name}</p>
                          <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                            <button 
                              onClick={() => openModal('dept', dep)}
                              className="p-1.5 text-gray-400 hover:text-emerald-600 rounded-full hover:bg-emerald-50"
                            >
                              <Edit className="h-3.5 w-3.5" />
                            </button>
                            <button 
                              onClick={() => handleDeleteDepartment(dep.id)}
                              className="p-1.5 text-gray-400 hover:text-rose-600 rounded-full hover:bg-rose-50"
                            >
                              <Trash2 className="h-3.5 w-3.5" />
                            </button>
                          </div>
                        </div>
                      </li>
                    ))}
                    {departments.length === 0 && (
                      <li className="text-center py-4 text-gray-500">Нет отделов</li>
                    )}
                  </ul>
                </div>
                <div className="rounded-4xl border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Должности</p>
                      <h2 className="mt-2 text-xl font-semibold">Каталог позиций</h2>
                    </div>
                    <button 
                      onClick={() => openModal('pos')}
                      className="rounded-full border border-gray-200 px-4 py-2 text-xs font-semibold text-gray-600 hover:bg-gray-50" 
                      type="button"
                    >
                      <Layers3 className="mr-1 inline h-3.5 w-3.5" /> Добавить
                    </button>
                  </div>
                  <ul className="mt-5 space-y-3 text-sm text-gray-600">
                    {positions.map((position) => (
                      <li key={position.id} className="rounded-3xl border border-gray-100 bg-gray-50/80 px-4 py-3 group hover:bg-gray-50 transition-colors">
                        <div className="flex items-center justify-between">
                          <p className="font-semibold text-gray-900">{position.name}</p>
                          <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                            <button 
                              onClick={() => openModal('pos', position)}
                              className="p-1.5 text-gray-400 hover:text-emerald-600 rounded-full hover:bg-emerald-50"
                            >
                              <Edit className="h-3.5 w-3.5" />
                            </button>
                            <button 
                              onClick={() => handleDeletePosition(position.id)}
                              className="p-1.5 text-gray-400 hover:text-rose-600 rounded-full hover:bg-rose-50"
                            >
                              <Trash2 className="h-3.5 w-3.5" />
                            </button>
                          </div>
                        </div>
                      </li>
                    ))}
                    {positions.length === 0 && (
                      <li className="text-center py-4 text-gray-500">Нет должностей</li>
                    )}
                  </ul>
                </div>
              </div>
            )}

            {activeTab === 'Навыки' && (
              <div className="rounded-4xl border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Навыки</p>
                    <h2 className="mt-2 text-xl font-semibold">Библиотека навыков</h2>
                  </div>
                  <div className="flex gap-2">
                    <button 
                      onClick={() => openModal('skill')}
                      className="rounded-full border border-gray-200 px-4 py-2 text-xs font-semibold text-gray-600 hover:bg-gray-50" 
                      type="button"
                    >
                      <ListChecks className="mr-1 inline h-3.5 w-3.5" /> Добавить навык
                    </button>
                  </div>
                </div>
                <div className="mt-5 grid gap-4 md:grid-cols-3 lg:grid-cols-4">
                  {skills.map((skill) => (
                    <div key={skill.id} className="group relative rounded-3xl border border-gray-100 bg-gray-50/80 p-4 hover:bg-white hover:shadow-md transition-all">
                      <p className="text-sm font-semibold text-gray-900">{skill.name}</p>
                      <div className="absolute top-2 right-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-all">
                        <button 
                          onClick={() => openModal('skill', skill)}
                          className="p-1.5 text-gray-400 hover:text-emerald-600 rounded-full hover:bg-emerald-50"
                          title="Редактировать навык"
                        >
                          <Edit className="h-3.5 w-3.5" />
                        </button>
                        <button 
                          onClick={() => handleDeleteSkill(skill.id)}
                          className="p-1.5 text-gray-400 hover:text-rose-600 rounded-full hover:bg-rose-50"
                          title="Удалить навык"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    </div>
                  ))}
                  {skills.length === 0 && (
                    <div className="col-span-full text-center py-8 text-gray-500">
                      Нет добавленных навыков
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>
        </section>
      </div>

      {/* Modals */}
      <Modal
        isOpen={isDeptModalOpen}
        onClose={() => { setIsDeptModalOpen(false); setEditingItem(null); }}
        title={editingItem ? "Редактировать отдел" : "Новый отдел"}
      >
        <form onSubmit={handleCreateOrUpdateDepartment} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Название отдела
            </label>
            <input
              type="text"
              value={newItemName}
              onChange={(e) => setNewItemName(e.target.value)}
              placeholder="Например: Отдел разработки"
              className="w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none"
              autoFocus
            />
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={() => { setIsDeptModalOpen(false); setEditingItem(null); }}
              className="rounded-xl px-4 py-2 text-sm font-medium text-gray-600 hover:bg-gray-50"
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={!newItemName.trim() || isSubmitting}
              className="rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-50"
            >
              {isSubmitting ? 'Сохранение...' : (editingItem ? 'Сохранить' : 'Создать')}
            </button>
          </div>
        </form>
      </Modal>

      <Modal
        isOpen={isPosModalOpen}
        onClose={() => { setIsPosModalOpen(false); setEditingItem(null); }}
        title={editingItem ? "Редактировать должность" : "Новая должность"}
      >
        <form onSubmit={handleCreateOrUpdatePosition} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Название должности
            </label>
            <input
              type="text"
              value={newItemName}
              onChange={(e) => setNewItemName(e.target.value)}
              placeholder="Например: Senior Frontend Developer"
              className="w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none"
              autoFocus
            />
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={() => { setIsPosModalOpen(false); setEditingItem(null); }}
              className="rounded-xl px-4 py-2 text-sm font-medium text-gray-600 hover:bg-gray-50"
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={!newItemName.trim() || isSubmitting}
              className="rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-50"
            >
              {isSubmitting ? 'Сохранение...' : (editingItem ? 'Сохранить' : 'Создать')}
            </button>
          </div>
        </form>
      </Modal>

      <Modal
        isOpen={isSkillModalOpen}
        onClose={() => { setIsSkillModalOpen(false); setEditingItem(null); }}
        title={editingItem ? "Редактировать навык" : "Новый навык"}
      >
        <form onSubmit={handleCreateOrUpdateSkill} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Название навыка
            </label>
            <input
              type="text"
              value={newItemName}
              onChange={(e) => setNewItemName(e.target.value)}
              placeholder="Например: React"
              className="w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none"
              autoFocus
            />
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={() => { setIsSkillModalOpen(false); setEditingItem(null); }}
              className="rounded-xl px-4 py-2 text-sm font-medium text-gray-600 hover:bg-gray-50"
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={!newItemName.trim() || isSubmitting}
              className="rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-50"
            >
              {isSubmitting ? 'Сохранение...' : (editingItem ? 'Сохранить' : 'Создать')}
            </button>
          </div>
        </form>
      </Modal>

      <Modal
        isOpen={isEditProfileModalOpen}
        onClose={() => { setIsEditProfileModalOpen(false); setEditingProfile(null); }}
        title="Редактирование профиля"
      >
        <form onSubmit={handleUpdateProfile} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Имя</label>
              <input
                type="text"
                value={profileForm.first_name || ''}
                onChange={(e) => setProfileForm({ ...profileForm, first_name: e.target.value })}
                className="w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Фамилия</label>
              <input
                type="text"
                value={profileForm.last_name || ''}
                onChange={(e) => setProfileForm({ ...profileForm, last_name: e.target.value })}
                className="w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none"
              />
            </div>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
            <input
              type="email"
              value={profileForm.email || ''}
              onChange={(e) => setProfileForm({ ...profileForm, email: e.target.value })}
              className="w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Отдел</label>
            <select
              value={profileForm.department_id || ''}
              onChange={(e) => setProfileForm({ ...profileForm, department_id: e.target.value })}
              className="w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none"
            >
              <option value="">Без отдела</option>
              {departments.map(d => (
                <option key={d.id} value={d.id}>{d.name}</option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Должность</label>
            <select
              value={profileForm.position_id || ''}
              onChange={(e) => setProfileForm({ ...profileForm, position_id: e.target.value })}
              className="w-full rounded-xl border border-gray-200 px-4 py-2.5 text-sm focus:border-emerald-400 focus:outline-none"
            >
              <option value="">Без должности</option>
              {positions.map(p => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </select>
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={() => { setIsEditProfileModalOpen(false); setEditingProfile(null); }}
              className="rounded-xl px-4 py-2 text-sm font-medium text-gray-600 hover:bg-gray-50"
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className="rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-50"
            >
              {isSubmitting ? 'Сохранение...' : 'Сохранить'}
            </button>
          </div>
        </form>
      </Modal>

      <Modal
        isOpen={isManageSkillsModalOpen}
        onClose={() => { setIsManageSkillsModalOpen(false); setSelectedProfileForSkills(null); }}
        title={`Навыки сотрудника: ${selectedProfileForSkills?.first_name} ${selectedProfileForSkills?.last_name}`}
      >
        <div className="space-y-6">
          <div>
            <h4 className="text-sm font-medium text-gray-700 mb-3">Текущие навыки</h4>
            <div className="flex flex-wrap gap-2">
              {selectedProfileForSkills?.skills && selectedProfileForSkills.skills.length > 0 ? (
                selectedProfileForSkills.skills.map(skill => (
                  <span key={skill.id} className="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-3 py-1 text-xs font-medium text-emerald-700 border border-emerald-100">
                    {skill.name}
                    <button
                      onClick={() => handleRemoveSkillFromEmployee(skill.id)}
                      className="ml-1 rounded-full p-0.5 hover:bg-emerald-200 text-emerald-600"
                      title="Удалить навык"
                    >
                      <Trash2 className="h-3 w-3" />
                    </button>
                  </span>
                ))
              ) : (
                <p className="text-sm text-gray-500 italic">Навыки не назначены</p>
              )}
            </div>
          </div>

          <div className="border-t border-gray-100 pt-4">
            <h4 className="text-sm font-medium text-gray-700 mb-3">Добавить навык</h4>
            <div className="flex gap-2">
              <select
                value={skillIdToAdd}
                onChange={(e) => setSkillIdToAdd(e.target.value)}
                className="flex-1 rounded-xl border border-gray-200 px-4 py-2 text-sm focus:border-emerald-400 focus:outline-none"
              >
                <option value="">Выберите навык...</option>
                {skills
                  .filter(s => !selectedProfileForSkills?.skills?.some(ps => ps.id === s.id))
                  .map(s => (
                    <option key={s.id} value={s.id}>{s.name}</option>
                  ))
                }
              </select>
              <button
                onClick={handleAddSkillToEmployee}
                disabled={!skillIdToAdd}
                className="rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Добавить
              </button>
            </div>
          </div>
          
          <div className="flex justify-end pt-2">
             <button
              onClick={() => { setIsManageSkillsModalOpen(false); setSelectedProfileForSkills(null); }}
              className="rounded-xl px-4 py-2 text-sm font-medium text-gray-600 hover:bg-gray-50"
            >
              Закрыть
            </button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
