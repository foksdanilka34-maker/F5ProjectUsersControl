import { useCallback, useEffect, useRef, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import {
  Award,
  Briefcase,
  Calendar,
  CheckCircle,
  Clock,
  Edit2,
  Mail,
  Save,
  TrendingUp,
  User,
  X,
} from 'lucide-react';
import {
  RadarChart,
  Radar,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  ResponsiveContainer,
  Tooltip,
} from 'recharts';
import Avatar from '../components/Avatar';

import { useAuth } from '../context/AuthContext';
import { useWebSocket } from '../context/WebSocketContext';
import { getProfile, updateProfile } from '../services/employeeService';
import { getEmployeeMetrics, type EmployeeMetrics } from '../services/analyticsService';
import type { ProfileDTO } from '../services/types';
import { roleLabels } from '../lib/constants';

export default function ProfilePage() {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const [profile, setProfile] = useState<ProfileDTO | null>(null);
  const [metrics, setMetrics] = useState<EmployeeMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [avatarUrl, setAvatarUrl] = useState<string | null>(null);

  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [email, setEmail] = useState('');

  const profileId = id ? parseInt(id, 10) : user?.id;
  const isOwnProfile = user?.id === profileId;
  const canEdit = isOwnProfile || ['admin', 'hr'].includes(user?.role || '');

  const loadData = useCallback(async () => {
    if (!profileId) return;
    setLoading(true);
    try {
      const [profileData, metricsData] = await Promise.all([
        getProfile(profileId),
        getEmployeeMetrics(profileId).catch((err) => {
          console.error('Failed to load metrics:', err);
          return null;
        }),
      ]);
      console.log('Profile loaded:', profileData);
      console.log('Metrics loaded:', metricsData);
      setProfile(profileData);
      setMetrics(metricsData);
      setFirstName(profileData.first_name);
      setLastName(profileData.last_name);
      setEmail(profileData.email);
      setAvatarUrl(profileData.avatar_url || null);
    } catch (err) {
      console.error('Failed to load profile:', err);
      setError('Не удалось загрузить профиль');
    } finally {
      setLoading(false);
    }
  }, [profileId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const { subscribe } = useWebSocket();

  const loadMetrics = useCallback(async () => {
    if (!profileId) return;
    try {
      const data = await getEmployeeMetrics(profileId);
      console.log('[Profile] Metrics reloaded:', data);
      setMetrics(data);
    } catch (err) {
      console.error('[Profile] Failed to reload metrics:', err);
    }
  }, [profileId]);

  const loadMetricsRef = useRef(loadMetrics);
  loadMetricsRef.current = loadMetrics;
  
  useEffect(() => {
    if (!profileId) return;

    const reloadMetrics = () => {
      console.log('[Profile] Task event received, reloading metrics...');
      loadMetricsRef.current();
    };

    const unsubCreated = subscribe('task:created', reloadMetrics);
    const unsubUpdated = subscribe('task:updated', reloadMetrics);
    const unsubDeleted = subscribe('task:deleted', reloadMetrics);
    const unsubMoved = subscribe('task:moved', reloadMetrics);
    const unsubAssigned = subscribe('task:assigned', reloadMetrics);

    return () => {
      unsubCreated();
      unsubUpdated();
      unsubDeleted();
      unsubMoved();
      unsubAssigned();
    };
  }, [profileId, subscribe]); // Убираем loadMetrics, используем ref

  const handleSave = async () => {
    if (!profileId || !profile) return;
    setSaving(true);
    setError(null);
    try {
      const payload: Record<string, unknown> = {
        first_name: firstName,
        last_name: lastName,
        email,
      };

      if (avatarUrl !== (profile.avatar_url || null)) {
        payload.avatar_url = avatarUrl || null;
      }
      await updateProfile(profileId, payload);
      setProfile({ ...profile, first_name: firstName, last_name: lastName, email, avatar_url: avatarUrl || undefined });
      setEditing(false);
    } catch (err) {
      console.error('Failed to update profile:', err);
      setError('Не удалось сохранить изменения');
    } finally {
      setSaving(false);
    }
  };

  const safePercent = (value: number) => {
    if (!Number.isFinite(value) || Number.isNaN(value)) return 0;
    return Math.max(0, Math.min(100, Math.round(value)));
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-white via-emerald-50/30 to-white">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-emerald-500 border-t-transparent" />
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-white via-emerald-50/30 to-white">
        <p className="text-gray-500">Профиль не найден</p>
        <Link to="/" className="mt-4 text-emerald-600 hover:underline">
          Вернуться на главную
        </Link>
      </div>
    );
  }

  const efficiency = safePercent(metrics?.completion_rate || 0);
  const onTimeRate = safePercent(metrics?.on_time_rate || 0);

  const taskCompletionPercent = metrics && metrics.assigned_tasks > 0
    ? safePercent((metrics.completed_tasks / metrics.assigned_tasks) * 100)
    : 0;

  return (
    <div className="min-h-screen bg-gradient-to-br from-white via-emerald-50/30 to-white text-gray-900">
      <div className="mx-auto max-w-4xl px-4 py-6">
        {}
        <header className="mb-6">
          <p className="text-xs font-semibold uppercase tracking-widest text-emerald-500">
            {isOwnProfile ? 'Мой профиль' : 'Профиль сотрудника'}
          </p>
          <h1 className="text-2xl font-bold text-gray-900">
            {profile.first_name} {profile.last_name}
          </h1>
        </header>

        {error && (
          <div className="mb-4 rounded-xl border border-red-100 bg-red-50 px-4 py-2 text-sm text-red-700">
            {error}
          </div>
        )}

        <div className="grid gap-6 lg:grid-cols-3">
          {}
          <div className="lg:col-span-2 rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
            <div className="flex items-start justify-between mb-6">
              <div className="flex items-center gap-4">
                <div className="relative">
                  <Avatar
                    src={avatarUrl}
                    name={`${profile.first_name} ${profile.last_name}`}
                    size="xl"
                  />
                  {canEdit && editing && avatarUrl && (
                    <button
                      type="button"
                      onClick={() => setAvatarUrl(null)}
                      className="absolute -bottom-2 -right-2 rounded-full bg-white p-1 shadow ring-1 ring-gray-200 text-gray-400 hover:text-rose-500"
                      title="Удалить фото"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  )}
                </div>
                <div>
                  {editing ? (
                    <div className="space-y-2">
                      <div className="flex flex-wrap gap-2">
                        <input
                          type="text"
                          value={firstName}
                          onChange={(e) => setFirstName(e.target.value)}
                          placeholder="Имя"
                          className="rounded-lg border border-gray-200 px-3 py-1.5 text-sm focus:border-emerald-400 focus:outline-none"
                        />
                        <input
                          type="text"
                          value={lastName}
                          onChange={(e) => setLastName(e.target.value)}
                          placeholder="Фамилия"
                          className="rounded-lg border border-gray-200 px-3 py-1.5 text-sm focus:border-emerald-400 focus:outline-none"
                        />
                      </div>
                      <input
                        type="email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        placeholder="Email"
                        className="w-full rounded-lg border border-gray-200 px-3 py-1.5 text-sm focus:border-emerald-400 focus:outline-none"
                      />
                      {canEdit && (
                        <div className="flex flex-wrap items-center gap-2">
                          <input
                            type="url"
                            value={avatarUrl || ''}
                            onChange={(e) => setAvatarUrl(e.target.value || null)}
                            placeholder="URL аватара (https://...)"
                            className="flex-1 rounded-lg border border-gray-200 px-3 py-1.5 text-sm focus:border-emerald-400 focus:outline-none"
                          />
                          {avatarUrl && (
                            <button
                              type="button"
                              onClick={() => setAvatarUrl(null)}
                              className="text-xs text-red-500 hover:underline"
                            >
                              Удалить
                            </button>
                          )}
                        </div>
                      )}
                    </div>
                  ) : (
                    <>
                      <h2 className="text-xl font-semibold text-gray-900">
                        {profile.first_name} {profile.last_name}
                      </h2>
                      <p className="text-sm text-gray-500">{profile.email}</p>
                    </>
                  )}
                </div>
              </div>
              {canEdit && (
                <div className="flex gap-2">
                  {editing ? (
                    <>
                      <button
                        type="button"
                        onClick={() => {
                          setEditing(false);
                          setFirstName(profile.first_name);
                          setLastName(profile.last_name);
                          setEmail(profile.email);
                          setAvatarUrl(profile.avatar_url || null);
                        }}
                        className="rounded-lg p-2 text-gray-400 hover:bg-gray-100"
                      >
                        <X className="h-4 w-4" />
                      </button>
                      <button
                        type="button"
                        onClick={handleSave}
                        disabled={saving}
                        className="rounded-lg bg-emerald-500 p-2 text-white hover:bg-emerald-600 disabled:opacity-50"
                      >
                        <Save className="h-4 w-4" />
                      </button>
                    </>
                  ) : (
                    <button
                      type="button"
                      onClick={() => setEditing(true)}
                      className="rounded-lg p-2 text-gray-400 hover:bg-gray-100 hover:text-emerald-600"
                    >
                      <Edit2 className="h-4 w-4" />
                    </button>
                  )}
                </div>
              )}
            </div>

            {}
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="flex items-center gap-3 rounded-xl border border-gray-100 p-4">
                <div className="rounded-lg bg-blue-50 p-2">
                  <Briefcase className="h-5 w-5 text-blue-600" />
                </div>
                <div>
                  <p className="text-xs text-gray-500">Роль</p>
                  <p className="font-medium text-gray-900">{roleLabels[profile.role] || profile.role}</p>
                </div>
              </div>

              <div className="flex items-center gap-3 rounded-xl border border-gray-100 p-4">
                <div className="rounded-lg bg-purple-50 p-2">
                  <User className="h-5 w-5 text-purple-600" />
                </div>
                <div>
                  <p className="text-xs text-gray-500">Отдел</p>
                  <p className="font-medium text-gray-900">{profile.department?.name || '—'}</p>
                </div>
              </div>

              <div className="flex items-center gap-3 rounded-xl border border-gray-100 p-4">
                <div className="rounded-lg bg-amber-50 p-2">
                  <Calendar className="h-5 w-5 text-amber-600" />
                </div>
                <div>
                  <p className="text-xs text-gray-500">Дата найма</p>
                  <p className="font-medium text-gray-900">
                    {new Date(profile.hire_date).toLocaleDateString('ru-RU')}
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-3 rounded-xl border border-gray-100 p-4">
                <div className="rounded-lg bg-emerald-50 p-2">
                  <Mail className="h-5 w-5 text-emerald-600" />
                </div>
                <div>
                  <p className="text-xs text-gray-500">Логин</p>
                  <p className="font-medium text-gray-900">{profile.login}</p>
                </div>
              </div>
            </div>

            {}
            {profile.skills && profile.skills.length > 0 && (
              <div className="mt-6">
                <h3 className="text-sm font-medium text-gray-700 mb-3">Навыки</h3>
                <div className="flex flex-wrap gap-2">
                  {profile.skills.map((skill) => (
                    <span
                      key={skill.id}
                      className="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-3 py-1 text-sm text-emerald-700"
                    >
                      <Award className="h-3.5 w-3.5" />
                      {skill.name}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>

          {}
          <div className="space-y-4">
            {/* Radar chart — employee performance */}
            <div className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
              <div className="flex items-center gap-3 mb-3">
                <div className="rounded-xl bg-emerald-50 p-2.5">
                  <TrendingUp className="h-5 w-5 text-emerald-600" />
                </div>
                <div>
                  <p className="text-xs text-gray-500">Эффективность</p>
                  <p className="text-xl font-bold text-gray-900">{efficiency}%</p>
                </div>
              </div>
              {metrics ? (
                (() => {
                  const load = metrics.assigned_tasks > 0
                    ? Math.min(100, Math.round((metrics.in_progress_tasks / metrics.assigned_tasks) * 100))
                    : 0;
                  const reliability = metrics.assigned_tasks > 0
                    ? Math.min(100, Math.round(((metrics.completed_tasks) / metrics.assigned_tasks) * 100))
                    : 0;
                  const radarData = [
                    { axis: 'Продуктивность', value: efficiency },
                    { axis: 'Пунктуальность', value: onTimeRate },
                    { axis: 'Загрузка', value: load },
                    { axis: 'Надёжность', value: reliability },
                  ];

                  // eslint-disable-next-line @typescript-eslint/no-explicit-any
                  const renderAxisTick = (props: any) => {
                    const { x, y, payload, cx: chartCx } = props;
                    const anchor = x < chartCx - 5 ? 'end' : x > chartCx + 5 ? 'start' : 'middle';
                    const dx = anchor === 'end' ? -6 : anchor === 'start' ? 6 : 0;
                    return (
                      <text x={x + dx} y={y} textAnchor={anchor} fontSize={10} fill="#6b7280" dominantBaseline="central">
                        {payload.value}
                      </text>
                    );
                  };

                  return (
                    <div style={{ height: 230 }}>
                      <ResponsiveContainer width="100%" height="100%">
                        <RadarChart data={radarData} cx="50%" cy="50%" outerRadius="55%">
                          <PolarGrid stroke="#e5e7eb" />
                          <PolarAngleAxis dataKey="axis" tick={renderAxisTick} />
                          <PolarRadiusAxis angle={90} domain={[0, 100]} tick={false} axisLine={false} />
                          <Radar
                            dataKey="value"
                            stroke="#10b981"
                            fill="#10b981"
                            fillOpacity={0.25}
                            strokeWidth={2}
                          />
                          <Tooltip
                            formatter={(value: number) => [`${value}%`]}
                            contentStyle={{ borderRadius: '12px', border: '1px solid #e5e7eb', fontSize: '12px' }}
                          />
                        </RadarChart>
                      </ResponsiveContainer>
                    </div>
                  );
                })()
              ) : (
                <div className="space-y-3">
                  <div>
                    <div className="flex justify-between text-xs mb-1">
                      <span className="text-gray-500">Выполнено задач</span>
                      <span className="font-medium">{taskCompletionPercent}%</span>
                    </div>
                    <div className="h-2 rounded-full bg-gray-100">
                      <div
                        className="h-2 rounded-full bg-gradient-to-r from-emerald-400 to-lime-400 transition-all"
                        style={{ width: `${taskCompletionPercent}%` }}
                      />
                    </div>
                  </div>
                  <div>
                    <div className="flex justify-between text-xs mb-1">
                      <span className="text-gray-500">В срок</span>
                      <span className="font-medium">{onTimeRate}%</span>
                    </div>
                    <div className="h-2 rounded-full bg-gray-100">
                      <div
                        className="h-2 rounded-full bg-gradient-to-r from-blue-400 to-cyan-400 transition-all"
                        style={{ width: `${onTimeRate}%` }}
                      />
                    </div>
                  </div>
                </div>
              )}
            </div>

            {/* Horizontal stacked bar — task breakdown */}
            <div className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
              <h3 className="text-sm font-medium text-gray-700 mb-4">Статистика задач</h3>
              {metrics ? (
                (() => {
                  const barData = [
                    { name: 'В срок', value: metrics.completed_on_time || 0, color: '#10b981' },
                    { name: 'С опозданием', value: metrics.completed_late || 0, color: '#f59e0b' },
                    { name: 'В работе', value: metrics.in_progress_tasks || 0, color: '#3b82f6' },
                    { name: 'Просрочено', value: metrics.overdue_tasks || 0, color: '#ef4444' },
                  ];
                  const total = metrics.assigned_tasks || 0;
                  return (
                    <div>
                      {/* Stacked horizontal bar */}
                      {total > 0 && (
                        <div className="mb-4">
                          <div className="flex h-4 rounded-full overflow-hidden bg-gray-100">
                            {barData.map((d) =>
                              d.value > 0 ? (
                                <div
                                  key={d.name}
                                  className="h-full transition-all duration-500"
                                  style={{
                                    width: `${(d.value / total) * 100}%`,
                                    backgroundColor: d.color,
                                  }}
                                  title={`${d.name}: ${d.value}`}
                                />
                              ) : null,
                            )}
                          </div>
                        </div>
                      )}

                      <div className="mb-4 space-y-3">
                        {barData.map((d) => (
                          <div key={d.name}>
                            <div className="flex justify-between text-xs mb-1">
                              <span className="text-gray-500">{d.name}</span>
                              <span className="font-medium text-gray-700">{d.value}</span>
                            </div>
                            <div className="h-3 rounded-full bg-gray-100 overflow-hidden">
                              {total > 0 && d.value > 0 && (
                                <div
                                  className="h-full rounded-full transition-all duration-500"
                                  style={{
                                    width: `${(d.value / total) * 100}%`,
                                    backgroundColor: d.color,
                                    minWidth: d.value > 0 ? '8px' : undefined,
                                  }}
                                />
                              )}
                            </div>
                          </div>
                        ))}
                      </div>

                      <div className="border-t border-gray-100 pt-3 flex items-center justify-between">
                        <span className="text-sm text-gray-600">Всего назначено</span>
                        <span className="font-semibold text-gray-900">{total}</span>
                      </div>
                    </div>
                  );
                })()
              ) : (
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <CheckCircle className="h-4 w-4 text-emerald-500" />
                      <span className="text-sm text-gray-600">Выполнено</span>
                    </div>
                    <span className="font-semibold text-gray-900">0</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Clock className="h-4 w-4 text-blue-500" />
                      <span className="text-sm text-gray-600">В работе</span>
                    </div>
                    <span className="font-semibold text-gray-900">0</span>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}


