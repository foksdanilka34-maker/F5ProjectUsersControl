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
  Loader2
} from 'lucide-react';
import { employeeService } from '../api/services/employee.service';
import type { Profile, Department, Position, Skill } from '../api/types';
import Modal from '../components/Modal';

const tabs = ['Профили', 'Оргструктура', 'Навыки'];

export default function EmployeesHubPage() {
  const [activeTab, setActiveTab] = useState('Профили');
  const [searchTerm, setSearchTerm] = useState('');
  const [isLoading, setIsLoading] = useState(true);

  // Modals state
  const [isDeptModalOpen, setIsDeptModalOpen] = useState(false);
  const [isPosModalOpen, setIsPosModalOpen] = useState(false);
  const [isSkillModalOpen, setIsSkillModalOpen] = useState(false);
  const [newItemName, setNewItemName] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Data state
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [skills, setSkills] = useState<Skill[]>([]);

  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true);
      try {
        const [profilesData, depsData, posData, skillsData] = await Promise.all([
          employeeService.listProfiles({ page_size: 100 }),
          employeeService.listDepartments(),
          employeeService.listPositions(),
          employeeService.listSkills()
        ]);

        setProfiles(profilesData.profiles);
        setDepartments(depsData);
        setPositions(posData);
        setSkills(skillsData);
      } catch (error) {
        console.error('Failed to fetch employees data:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, []);

  const openModal = (type: 'dept' | 'pos' | 'skill') => {
    setNewItemName('');
    if (type === 'dept') setIsDeptModalOpen(true);
    if (type === 'pos') setIsPosModalOpen(true);
    if (type === 'skill') setIsSkillModalOpen(true);
  };

  const profileStats = useMemo(
    () => [
      { label: 'Профилей в базе', value: profiles.length },
      { label: 'Активные', value: profiles.filter((p) => p.is_active !== false).length },
      { label: 'Отделов', value: departments.length },
      { label: 'Навыков', value: skills.length },
    ],
    [profiles, departments, skills],
  );

  const filteredProfiles = useMemo(() => {
    if (!searchTerm.trim()) {
      return profiles;
    }
    const normalized = searchTerm.toLowerCase();
    return profiles.filter((profile) => {
      const fullName = `${profile.first_name} ${profile.last_name}`.toLowerCase();
      const departmentName = profile.department?.name.toLowerCase() || '';
      const positionName = profile.position?.name.toLowerCase() || '';
      
      return (
        fullName.includes(normalized) ||
        departmentName.includes(normalized) ||
        positionName.includes(normalized) ||
        profile.email.toLowerCase().includes(normalized)
      );
    });
  }, [searchTerm, profiles]);

  const handleCreateDepartment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newItemName.trim()) return;

    setIsSubmitting(true);
    try {
      const newDept = await employeeService.createDepartment({ name: newItemName });
      setDepartments([...departments, newDept]);
      setNewItemName('');
      setIsDeptModalOpen(false);
    } catch (error) {
      console.error('Failed to create department:', error);
      alert('Не удалось создать отдел');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCreatePosition = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newItemName.trim()) return;

    setIsSubmitting(true);
    try {
      const newPos = await employeeService.createPosition({ name: newItemName });
      setPositions([...positions, newPos]);
      setNewItemName('');
      setIsPosModalOpen(false);
    } catch (error) {
      console.error('Failed to create position:', error);
      alert('Не удалось создать должность');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCreateSkill = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newItemName.trim()) return;

    setIsSubmitting(true);
    try {
      const newSkill = await employeeService.createSkill({ name: newItemName });
      setSkills([...skills, newSkill]);
      setNewItemName('');
      setIsSkillModalOpen(false);
    } catch (error) {
      console.error('Failed to create skill:', error);
      alert('Не удалось создать навык');
    } finally {
      setIsSubmitting(false);
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

        <section className="mt-8 grid gap-6 lg:grid-cols-[1fr]">
          <div className="space-y-6">
            <div className="rounded-4xl border border-gray-100 bg-white/95 p-6 shadow-[0_25px_80px_rgba(6,95,70,0.08)]">
              <div className="flex flex-wrap items-center justify-between gap-4">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Быстрые действия</p>
                  <h2 className="mt-2 text-xl font-semibold">Что вы хотите сделать?</h2>
                </div>
                <label className="relative flex items-center">
                  <Search className="absolute left-3 h-4 w-4 text-gray-400" />
                  <input
                    type="text"
                    value={searchTerm}
                    onChange={(event) => setSearchTerm(event.target.value)}
                    placeholder="Поиск сотрудника..."
                    className="w-64 rounded-full border border-gray-200 bg-white px-9 py-2 text-sm text-gray-600 focus:border-emerald-400 focus:outline-none"
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

            {activeTab === 'Профили' && (
              <div className="rounded-4xl border border-gray-100 bg-white/95 p-6 shadow-[0_20px_70px_rgba(6,95,70,0.08)]">
                <div className="flex flex-wrap items-center justify-between gap-4">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">Сотрудники</p>
                    <h2 className="mt-2 text-xl font-semibold">Список профилей</h2>
                  </div>
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
                            <td className="px-6 py-4">{profile.position?.name || '-'}</td>
                            <td className="px-6 py-4">{profile.department?.name || '-'}</td>
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
                              <button type="button" className="rounded-full border border-gray-200 px-3 py-1 text-xs text-gray-600 hover:bg-gray-100">
                                <Edit className="mr-1 inline h-3.5 w-3.5" /> Редактировать
                              </button>
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
                      <li key={dep.id} className="rounded-3xl border border-gray-100 bg-gray-50/80 px-4 py-3">
                        <div className="flex items-center justify-between">
                          <p className="font-semibold text-gray-900">{dep.name}</p>
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
                      <li key={position.id} className="rounded-3xl border border-gray-100 bg-gray-50/80 px-4 py-3">
                        <div className="flex items-center justify-between">
                          <p className="font-semibold text-gray-900">{position.name}</p>
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
                    <div key={skill.id} className="rounded-3xl border border-gray-100 bg-gray-50/80 p-4">
                      <p className="text-sm font-semibold text-gray-900">{skill.name}</p>
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
        onClose={() => setIsDeptModalOpen(false)}
        title="Новый отдел"
      >
        <form onSubmit={handleCreateDepartment} className="space-y-4">
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
              onClick={() => setIsDeptModalOpen(false)}
              className="rounded-xl px-4 py-2 text-sm font-medium text-gray-600 hover:bg-gray-50"
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={!newItemName.trim() || isSubmitting}
              className="rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-50"
            >
              {isSubmitting ? 'Создание...' : 'Создать'}
            </button>
          </div>
        </form>
      </Modal>

      <Modal
        isOpen={isPosModalOpen}
        onClose={() => setIsPosModalOpen(false)}
        title="Новая должность"
      >
        <form onSubmit={handleCreatePosition} className="space-y-4">
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
              onClick={() => setIsPosModalOpen(false)}
              className="rounded-xl px-4 py-2 text-sm font-medium text-gray-600 hover:bg-gray-50"
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={!newItemName.trim() || isSubmitting}
              className="rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-50"
            >
              {isSubmitting ? 'Создание...' : 'Создать'}
            </button>
          </div>
        </form>
      </Modal>

      <Modal
        isOpen={isSkillModalOpen}
        onClose={() => setIsSkillModalOpen(false)}
        title="Новый навык"
      >
        <form onSubmit={handleCreateSkill} className="space-y-4">
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
              onClick={() => setIsSkillModalOpen(false)}
              className="rounded-xl px-4 py-2 text-sm font-medium text-gray-600 hover:bg-gray-50"
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={!newItemName.trim() || isSubmitting}
              className="rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-50"
            >
              {isSubmitting ? 'Создание...' : 'Создать'}
            </button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
