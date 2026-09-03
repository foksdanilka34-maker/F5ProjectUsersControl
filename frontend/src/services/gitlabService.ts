import { apiClient } from '../lib/apiClient';

export type GitLinkKind = 'BRANCH' | 'MERGE_REQUEST' | 'COMMIT';

export interface GitLabIntegration {
  project_id: number;
  base_url: string;
  gitlab_project_id: number;
  default_branch: string;
  task_key_prefix: string;
  auto_move_in_progress: boolean;
  auto_move_review: boolean;
  auto_close_on_merge: boolean;
  is_active: boolean;
  token_set: boolean;
  webhook_url: string;
  webhook_secret: string;
  created_at: string;
  updated_at: string;
}

export interface GitLabIntegrationState {
  connected: boolean;
  integration: GitLabIntegration | null;
}

export interface SaveIntegrationRequest {
  base_url: string;
  gitlab_project_id: number;
  access_token?: string;
  default_branch: string;
  task_key_prefix: string;
  auto_move_in_progress?: boolean;
  auto_move_review?: boolean;
  auto_close_on_merge?: boolean;
  is_active?: boolean;
}

export interface TaskGitLink {
  id: number;
  task_id: number;
  kind: GitLinkKind;
  external_id: string;
  title?: string;
  state?: string;
  web_url?: string;
  author_name?: string;
  created_at: string;
  updated_at: string;
}

export interface GitLabPipeline {
  id: number;
  project_id: number;
  task_id?: number;
  pipeline_id: number;
  ref: string;
  sha: string;
  status: string;
  duration_sec?: number;
  web_url?: string;
  started_at?: string;
  finished_at?: string;
  updated_at: string;
}

export interface TaskGitOverview {
  connected: boolean;
  task_key: string;
  suggested_branch: string;
  repository_url: string;
  links: TaskGitLink[] | null;
  pipelines: GitLabPipeline[] | null;
}

export interface ProjectGitSummaryItem {
  task_id: number;
  branches: number;
  merge_requests: number;
  pipeline_status?: string;
  pipeline_url?: string;
}

export interface GitLabProjectInfo {
  id: number;
  name: string;
  path_with_namespace: string;
  web_url: string;
  default_branch: string;
}

export const getIntegration = async (projectId: number): Promise<GitLabIntegrationState> => {
  return apiClient.request<GitLabIntegrationState>(`/gitlab/projects/${projectId}`, {
    method: 'GET',
  });
};

export const saveIntegration = async (
  projectId: number,
  data: SaveIntegrationRequest
): Promise<GitLabIntegrationState> => {
  return apiClient.request<GitLabIntegrationState>(`/gitlab/projects/${projectId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
};

export const deleteIntegration = async (projectId: number): Promise<void> => {
  await apiClient.request(`/gitlab/projects/${projectId}`, {
    method: 'DELETE',
  });
};

export const testConnection = async (projectId: number): Promise<GitLabProjectInfo> => {
  const response = await apiClient.request<{ project: GitLabProjectInfo }>(
    `/gitlab/projects/${projectId}/test`,
    { method: 'POST' }
  );
  return response.project;
};

export const getProjectGitSummary = async (projectId: number): Promise<ProjectGitSummaryItem[]> => {
  const response = await apiClient.request<{ items: ProjectGitSummaryItem[] | null }>(
    `/gitlab/projects/${projectId}/summary`,
    { method: 'GET' }
  );
  return response.items || [];
};

export const getTaskGit = async (taskId: number): Promise<TaskGitOverview> => {
  return apiClient.request<TaskGitOverview>(`/gitlab/tasks/${taskId}`, {
    method: 'GET',
  });
};

export const createTaskBranch = async (taskId: number): Promise<TaskGitLink> => {
  return apiClient.request<TaskGitLink>(`/gitlab/tasks/${taskId}/branch`, {
    method: 'POST',
  });
};

export const retryPipeline = async (taskId: number, pipelineId: number): Promise<GitLabPipeline> => {
  return apiClient.request<GitLabPipeline>(`/gitlab/tasks/${taskId}/pipelines/${pipelineId}/retry`, {
    method: 'POST',
  });
};

export const PIPELINE_STATUS_STYLES: Record<string, { label: string; class: string; dot: string }> = {
  success: { label: 'Успешно', class: 'bg-emerald-100 text-emerald-700', dot: 'bg-emerald-500' },
  failed: { label: 'Упал', class: 'bg-red-100 text-red-700', dot: 'bg-red-500' },
  running: { label: 'Выполняется', class: 'bg-blue-100 text-blue-700', dot: 'bg-blue-500' },
  pending: { label: 'В очереди', class: 'bg-amber-100 text-amber-700', dot: 'bg-amber-500' },
  canceled: { label: 'Отменён', class: 'bg-gray-100 text-gray-600', dot: 'bg-gray-400' },
  skipped: { label: 'Пропущен', class: 'bg-gray-100 text-gray-600', dot: 'bg-gray-400' },
};

export const pipelineStyle = (status?: string) =>
  PIPELINE_STATUS_STYLES[status ?? ''] ?? {
    label: status || 'Неизвестно',
    class: 'bg-gray-100 text-gray-600',
    dot: 'bg-gray-400',
  };
