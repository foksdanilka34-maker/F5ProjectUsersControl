// API Response wrapper
export interface ApiResponse<T = any> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
  meta?: PaginationMeta;
}

export interface PaginationMeta {
  page_size?: number;
  page_number?: number;
  total_count?: number;
}

// Auth types
export type UserRole = 'specialist' | 'manager' | 'admin' | 'director';

export interface LoginRequest {
  login: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token?: string;
  expires_in: number;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface RefreshResponse {
  access_token: string;
  refresh_token?: string;
  expires_in: number;
}

export interface User {
  user_id: string;
  role: UserRole;
  login: string;
}

// Employee types
export interface Department {
  id: string;
  name: string;
}

export interface Position {
  id: string;
  name: string;
}

export interface Skill {
  id: string;
  name: string;
}

export interface Profile {
  id: string;
  first_name: string;
  last_name: string;
  position_id?: string;
  position?: Position;
  email: string;
  department_id?: string;
  department?: Department;
  avatar_url?: string;
  hire_date: string;
  created_at: string;
  updated_at: string;
  skills?: Skill[];
  is_active?: boolean;
}

export interface CreateProfileRequest {
  first_name: string;
  last_name: string;
  email: string;
  hire_date: string; // ISO date string
  login: string;
  password: string;
  role: UserRole;
  position_id: string;
  department_id?: string;
}

export interface UpdateProfileRequest {
  first_name?: string;
  last_name?: string;
  email?: string;
  position_id?: string;
  department_id?: string;
  avatar_url?: string;
}

export interface ChangeUserStatusRequest {
  status: boolean;
}

export interface CreateDepartmentRequest {
  name: string;
}

export interface UpdateDepartmentRequest {
  name: string;
}

export interface CreatePositionRequest {
  name: string;
}

export interface UpdatePositionRequest {
  name: string;
}

export interface CreateSkillRequest {
  name: string;
}

export interface AddSkillToEmployeeRequest {
  skill_id: string;
}

export interface ListProfilesParams {
  page_size?: number;
  page_number?: number;
  department_id?: string;
  position_id?: string;
}

// Project types
export type ProjectStatus = 0 | 1 | 2 | 3; // UNSPECIFIED, ACTIVE, ON_HOLD, ARCHIVED
export type TaskStatus = 0 | 1 | 2 | 3 | 4; // UNSPECIFIED, TODO, IN_PROGRESS, REVIEW, DONE
export type TaskPriority = 0 | 1 | 2 | 3 | 4; // UNSPECIFIED, LOW, MEDIUM, HIGH, CRITICAL

export interface Project {
  id: string;
  name: string;
  description?: string;
  manager_id: string;
  manager_name?: string;
  status: ProjectStatus;
  created_at: string;
  updated_at: string;
  due_date?: string;
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
  due_date?: string; // ISO date string
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  due_date?: string;
  status?: ProjectStatus;
}

export interface Task {
  id: string;
  project_id: string;
  title: string;
  description?: string;
  priority: TaskPriority;
  status: TaskStatus;
  creator_id: string;
  assignee_id?: string;
  assignee_name?: string;
  order_index: number;
  created_at: string;
  updated_at: string;
  due_date: string;
  started_at?: string;
  completed_at?: string;
}

export interface CreateTaskRequest {
  title: string;
  description?: string;
  priority?: TaskPriority;
  assignee_id?: string;
  due_date: string;
}

export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  priority?: TaskPriority;
  status?: TaskStatus;
  assignee_id?: string;
  due_date?: string;
}

export interface MoveTaskRequest {
  new_status: TaskStatus;
  new_order_index: number;
}

export interface AssignTaskRequest {
  assignee_id: string;
}

export interface ProjectMember {
  user_id: string;
  user_name: string;
  role: string;
  added_at: string;
}

export interface AddMemberToProjectRequest {
  user_id: string;
}

export interface ListProjectsParams {
  page_size?: number;
  page_number?: number;
  manager_id?: string;
  status?: ProjectStatus;
}

export interface ListTasksParams {
  status?: TaskStatus;
  assignee_id?: string;
  priority?: TaskPriority;
}

// Analytics types
export interface DashboardStats {
  total_employees: number;
  active_employees: number;
  total_projects: number;
  active_projects: number;
  total_tasks: number;
  completed_tasks: number;
  overdue_tasks: number;
  completion_rate: number;
  average_productivity: number;
}

export interface EmployeeMetrics {
  employee_id: string;
  employee_name: string;
  department_name?: string;
  total_tasks: number;
  completed_tasks: number;
  in_progress_tasks: number;
  overdue_tasks: number;
  completion_rate: number;
  average_completion_time_hours: number;
  productivity_score: number;
}

export interface ProjectMetrics {
  project_id: string;
  project_name: string;
  manager_name: string;
  total_tasks: number;
  completed_tasks: number;
  in_progress_tasks: number;
  overdue_tasks: number;
  completion_rate: number;
  on_time_completion_rate: number;
  average_task_duration_hours: number;
  health_score: number;
}

export interface TopPerformer {
  employee_id: string;
  employee_name: string;
  department_name?: string;
  completed_tasks: number;
  productivity_score: number;
  rank: number;
}

export interface ProductivityTrend {
  period: string;
  date: string;
  productivity_score: number;
  completed_tasks: number;
  total_tasks: number;
}

export interface CompletionRateTrend {
  period: string;
  date: string;
  completion_rate: number;
  on_time_rate: number;
  total_completed: number;
}

export interface DashboardStatsParams {
  start_date?: string;
  end_date?: string;
}

export interface ListEmployeeMetricsParams {
  page_size?: number;
  page_number?: number;
  start_date?: string;
  end_date?: string;
  department_id?: string;
}

export interface TopPerformersParams {
  start_date?: string;
  end_date?: string;
  limit?: number;
  department_id?: string;
}

export interface ListProjectMetricsParams {
  page_size?: number;
  page_number?: number;
  start_date?: string;
  end_date?: string;
  manager_id?: string;
}

export interface ProductivityTrendsParams {
  period?: 'DAILY' | 'WEEKLY' | 'MONTHLY';
  limit?: number;
  department_id?: string;
  employee_id?: string;
}

export interface CompletionRateTrendsParams {
  period?: 'DAILY' | 'WEEKLY' | 'MONTHLY';
  limit?: number;
  project_id?: string;
  department_id?: string;
}
