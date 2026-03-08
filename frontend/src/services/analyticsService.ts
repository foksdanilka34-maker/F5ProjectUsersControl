import { apiClient } from '../lib/apiClient';

export interface DashboardStats {
  total_employees: number;
  active_employees: number;
  total_projects: number;
  active_projects: number;
  total_tasks: number;
  completed_tasks: number;
  overdue_tasks: number;           // Текущие просроченные (не завершены и дедлайн прошёл)
  completed_on_time: number;      // Завершено вовремя
  completed_late: number;         // Завершено с опозданием
  avg_completion_rate: number;
  avg_on_time_rate: number;       // CompletedOnTime / CompletedTasks * 100
  top_employees: TopEmployee[];
  problematic_projects: ProblematicProject[];
}

export interface TopEmployee {
  employee_id: number;
  completion_rate: number;
  tasks_completed: number;
}

export interface ProblematicProject {
  project_id: number;
  on_time_rate: number;
  health_status: string;
}

export interface ProjectMetrics {
  project_id: number;
  manager_id: number;
  total_tasks: number;
  completed_tasks: number;
  in_progress_tasks: number;
  overdue_tasks: number;           // Текущие просроченные (не завершены)
  completed_on_time: number;      // Завершено вовремя
  completed_late: number;         // Завершено с опозданием
  team_size: number;
  progress_percent: number;
  on_time_rate: number;           // CompletedOnTime / CompletedTasks * 100
  health_status: string;
}

export interface EmployeeMetrics {
  employee_id: number;
  assigned_tasks: number;
  completed_tasks: number;
  in_progress_tasks: number;
  overdue_tasks: number;           // Текущие просроченные (не завершены)
  completed_on_time: number;      // Завершено вовремя
  completed_late: number;         // Завершено с опозданием
  completion_rate: number;
  on_time_rate: number;           // CompletedOnTime / CompletedTasks * 100
}

export const getDashboardStats = async (): Promise<DashboardStats> => {
  return apiClient.request<DashboardStats>('/analytics/dashboard', {
    method: 'GET',
  });
};

export const getSummary = async (): Promise<DashboardStats> => {
  return apiClient.request<DashboardStats>('/analytics/summary', {
    method: 'GET',
  });
};

export const getProjectMetrics = async (projectId: number): Promise<ProjectMetrics> => {
  return apiClient.request<ProjectMetrics>(`/analytics/project/${projectId}`, {
    method: 'GET',
  });
};

export const getEmployeeMetrics = async (employeeId: number): Promise<EmployeeMetrics> => {
  return apiClient.request<EmployeeMetrics>(`/analytics/employee/${employeeId}`, {
    method: 'GET',
  });
};


