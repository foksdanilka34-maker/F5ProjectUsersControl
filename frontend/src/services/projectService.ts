import { apiClient } from '../lib/apiClient';

export interface Project {
  id: number;
  name: string;
  description: string;
  status: string;
  manager_id: number;
  due_date?: string;
  created_at: string;
  updated_at: string;
}

export interface Task {
  id: number;
  project_id: number;
  title: string;
  description: string;
  status: string;
  priority: string;
  assignee_id: number;
  due_date?: string;
  created_at: string;
  updated_at: string;
}

export interface ProjectMember {
  id: number;
  project_id: number;
  user_id: number;
  role: string;
  joined_at: string;
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
  manager_id: number;
}

export interface CreateTaskRequest {
  project_id: number;
  title: string;
  description?: string;
  priority?: number;
  assignee_id?: number;
  due_date?: string;
}

export interface GetProjectsParams {
  status?: string;
  member_id?: number;
  manager_id?: number;
}

export const getProjects = async (params?: GetProjectsParams): Promise<Project[]> => {
  const searchParams = new URLSearchParams();
  if (params?.status) searchParams.set('status', params.status);
  if (params?.member_id) searchParams.set('member_id', String(params.member_id));
  if (params?.manager_id) searchParams.set('manager_id', String(params.manager_id));
  
  const queryString = searchParams.toString();
  const url = queryString ? `/projects?${queryString}` : '/projects';
  
  const response = await apiClient.request<{ projects: Project[] }>(url, {
    method: 'GET',
  });
  return response.projects || [];
};

export const getProject = async (id: number): Promise<Project> => {
  return apiClient.request<Project>(`/projects/${id}`, {
    method: 'GET',
  });
};

export const createProject = async (data: CreateProjectRequest): Promise<Project> => {
  return apiClient.request<Project>('/projects', {
    method: 'POST',
    body: JSON.stringify(data),
  });
};

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  manager_id?: number;
  status?: string;
  due_date?: string;
}

export const updateProject = async (id: number, data: UpdateProjectRequest): Promise<Project> => {
  return apiClient.request<Project>(`/projects/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
};

export const deleteProject = async (id: number): Promise<void> => {
  await apiClient.request(`/projects/${id}`, {
    method: 'DELETE',
  });
};

export const getTasks = async (projectId?: number): Promise<Task[]> => {
  const url = projectId ? `/tasks?project_id=${projectId}` : '/tasks';
  const response = await apiClient.request<{ tasks: Task[] }>(url, {
    method: 'GET',
  });
  return response.tasks || [];
};

export const getTask = async (id: number): Promise<Task> => {
  return apiClient.request<Task>(`/tasks/${id}`, {
    method: 'GET',
  });
};

export const createTask = async (data: CreateTaskRequest): Promise<Task> => {
  return apiClient.request<Task>('/tasks', {
    method: 'POST',
    body: JSON.stringify(data),
  });
};

export const updateTask = async (id: number, data: Partial<CreateTaskRequest & { status: string }>): Promise<Task> => {
  return apiClient.request<Task>(`/tasks/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
};

export const deleteTask = async (id: number): Promise<void> => {
  await apiClient.request(`/tasks/${id}`, {
    method: 'DELETE',
  });
};

export const moveTask = async (id: number, newStatus: string): Promise<Task> => {
  return apiClient.request<Task>(`/tasks/${id}/move`, {
    method: 'POST',
    body: JSON.stringify({ new_status: newStatus }),
  });
};

export const assignTask = async (id: number, assigneeId: number): Promise<Task> => {
  return apiClient.request<Task>(`/tasks/${id}/assign`, {
    method: 'POST',
    body: JSON.stringify({ assignee_id: assigneeId }),
  });
};

export const getProjectMembers = async (projectId: number): Promise<ProjectMember[]> => {
  const response = await apiClient.request<{ members: ProjectMember[] }>(`/projects/${projectId}/members`, {
    method: 'GET',
  });
  return response.members || [];
};

export const addProjectMember = async (projectId: number, userId: number, role: string): Promise<ProjectMember> => {
  return apiClient.request<ProjectMember>(`/projects/${projectId}/members`, {
    method: 'POST',
    body: JSON.stringify({ user_id: userId, role }),
  });
};

export const removeProjectMember = async (projectId: number, userId: number): Promise<void> => {
  await apiClient.request(`/projects/${projectId}/members/${userId}`, {
    method: 'DELETE',
  });
};

// Review API

export interface ReviewerInfo {
  user_id: number;
  approved: boolean;
}

export interface ReviewStatus {
  reviewers: ReviewerInfo[];
  is_active: boolean;
}

export const submitForReview = async (taskId: number): Promise<Task> => {
  return apiClient.request<Task>(`/tasks/${taskId}/review/submit`, {
    method: 'POST',
  });
};

export const approveTask = async (taskId: number): Promise<Task> => {
  return apiClient.request<Task>(`/tasks/${taskId}/review/approve`, {
    method: 'POST',
  });
};

export const getReviewStatus = async (taskId: number): Promise<ReviewStatus> => {
  return apiClient.request<ReviewStatus>(`/tasks/${taskId}/review`, {
    method: 'GET',
  });
};


