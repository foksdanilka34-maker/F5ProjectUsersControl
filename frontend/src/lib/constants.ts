
export const roleLabels: Record<string, string> = {
  admin: 'Администратор',
  manager: 'Менеджер',
  developer: 'Разработчик',
  analyst: 'Аналитик',
  designer: 'Дизайнер',
  tester: 'Тестировщик',
  hr: 'HR-специалист',
  viewer: 'Наблюдатель',
  user: 'Пользователь',
};

export const projectStatusLabels: Record<string, string> = {
  ACTIVE: 'Активный',
  PLANNING: 'Планирование',
  ON_HOLD: 'На паузе',
  COMPLETED: 'Завершён',
  CANCELLED: 'Отменён',
  PROJECT_STATUS_UNSPECIFIED: 'Без статуса',
};

export const projectStatusColors: Record<string, string> = {
  ACTIVE: 'bg-green-100 text-green-800',
  PLANNING: 'bg-blue-100 text-blue-800',
  ON_HOLD: 'bg-yellow-100 text-yellow-800',
  COMPLETED: 'bg-gray-100 text-gray-800',
  CANCELLED: 'bg-red-100 text-red-800',
  PROJECT_STATUS_UNSPECIFIED: 'bg-gray-100 text-gray-500',
};

export const taskStatusLabels: Record<string, string> = {
  TODO: 'К выполнению',
  IN_PROGRESS: 'В работе',
  REVIEW: 'На проверке',
  DONE: 'Выполнено',
  BLOCKED: 'Заблокировано',
};

export const taskStatusColors: Record<string, string> = {
  TODO: 'bg-gray-100 text-gray-800',
  IN_PROGRESS: 'bg-blue-100 text-blue-800',
  REVIEW: 'bg-purple-100 text-purple-800',
  DONE: 'bg-green-100 text-green-800',
  BLOCKED: 'bg-red-100 text-red-800',
};

export const taskPriorityLabels: Record<string, string> = {
  LOW: 'Низкий',
  MEDIUM: 'Средний',
  HIGH: 'Высокий',
  CRITICAL: 'Критический',
};

export const taskPriorityColors: Record<string, string> = {
  LOW: 'bg-gray-100 text-gray-600',
  MEDIUM: 'bg-blue-100 text-blue-700',
  HIGH: 'bg-orange-100 text-orange-700',
  CRITICAL: 'bg-red-100 text-red-700',
};

export const getLabel = (
  value: string | undefined | null,
  labels: Record<string, string>,
  fallback = 'Не указано'
): string => {
  if (!value) return fallback;
  return labels[value] || value;
};


