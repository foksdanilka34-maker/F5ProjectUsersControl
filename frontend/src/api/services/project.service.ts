import { apiClient, unwrapResponse } from '../client';
import type {
  Project,
  CreateProjectRequest,
  UpdateProjectRequest,
  Task,
  CreateTaskRequest,
  UpdateTaskRequest,
  MoveTaskRequest,
  AssignTaskRequest,
  ProjectMember,
  AddMemberToProjectRequest,
  ListProjectsParams,
  ListTasksParams,
  ApiResponse,
  PaginationMeta,
} from '../types';

export const projectService = {
  // ===== Projects =====
  async createProject(data: CreateProjectRequest): Promise<Project> {
    const response = await apiClient.post<ApiResponse<Project>>(
      '/api/v1/projects',
      data
    );
    return unwrapResponse(response.data);
  },

  async listProjects(params?: ListProjectsParams): Promise<{ projects: Project[]; meta?: PaginationMeta }> {
    const response = await apiClient.get<ApiResponse<Project[]>>(
      '/api/v1/projects',
      { params }
    );
    return {
      projects: unwrapResponse(response.data),
      meta: response.data.meta,
    };
  },

  async getProject(id: string): Promise<Project> {
    const response = await apiClient.get<ApiResponse<Project>>(
      `/api/v1/projects/${id}`
    );
    return unwrapResponse(response.data);
  },

  async updateProject(id: string, data: UpdateProjectRequest): Promise<Project> {
    const response = await apiClient.patch<ApiResponse<Project>>(
      `/api/v1/projects/${id}`,
      data
    );
    return unwrapResponse(response.data);
  },

  async deleteProject(id: string): Promise<void> {
    const response = await apiClient.delete<ApiResponse>(
      `/api/v1/projects/${id}`
    );
    unwrapResponse(response.data);
  },

  // ===== Project Members =====
  async listProjectMembers(projectId: string): Promise<ProjectMember[]> {
    const response = await apiClient.get<ApiResponse<ProjectMember[]>>(
      `/api/v1/projects/${projectId}/members`
    );
    return unwrapResponse(response.data);
  },

  async addMemberToProject(projectId: string, data: AddMemberToProjectRequest): Promise<void> {
    const response = await apiClient.post<ApiResponse>(
      `/api/v1/projects/${projectId}/members`,
      data
    );
    unwrapResponse(response.data);
  },

  async removeMemberFromProject(projectId: string, memberId: string): Promise<void> {
    const response = await apiClient.delete<ApiResponse>(
      `/api/v1/projects/${projectId}/members/${memberId}`
    );
    unwrapResponse(response.data);
  },

  // ===== Tasks =====
  async createTask(projectId: string, data: CreateTaskRequest): Promise<Task> {
    const response = await apiClient.post<ApiResponse<Task>>(
      `/api/v1/projects/${projectId}/tasks`,
      data
    );
    return unwrapResponse(response.data);
  },

  async listTasksByProject(projectId: string, params?: ListTasksParams): Promise<Task[]> {
    const response = await apiClient.get<ApiResponse<Task[]>>(
      `/api/v1/projects/${projectId}/tasks`,
      { params }
    );
    return unwrapResponse(response.data);
  },

  async getTask(projectId: string, taskId: string): Promise<Task> {
    const response = await apiClient.get<ApiResponse<Task>>(
      `/api/v1/projects/${projectId}/tasks/${taskId}`
    );
    return unwrapResponse(response.data);
  },

  async updateTask(projectId: string, taskId: string, data: UpdateTaskRequest): Promise<Task> {
    const response = await apiClient.patch<ApiResponse<Task>>(
      `/api/v1/projects/${projectId}/tasks/${taskId}`,
      data
    );
    return unwrapResponse(response.data);
  },

  async deleteTask(projectId: string, taskId: string): Promise<void> {
    const response = await apiClient.delete<ApiResponse>(
      `/api/v1/projects/${projectId}/tasks/${taskId}`
    );
    unwrapResponse(response.data);
  },

  async moveTask(projectId: string, taskId: string, data: MoveTaskRequest): Promise<Task> {
    const response = await apiClient.post<ApiResponse<Task>>(
      `/api/v1/projects/${projectId}/tasks/${taskId}/move`,
      data
    );
    return unwrapResponse(response.data);
  },

  async assignTask(projectId: string, taskId: string, data: AssignTaskRequest): Promise<Task> {
    const response = await apiClient.post<ApiResponse<Task>>(
      `/api/v1/projects/${projectId}/tasks/${taskId}/assign`,
      data
    );
    return unwrapResponse(response.data);
  },
};
