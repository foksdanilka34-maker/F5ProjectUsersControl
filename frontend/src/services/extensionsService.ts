import { apiClient } from '../lib/apiClient';

export interface Extension {
  id: number;
  key: string;
  name: string;
  description: string;
  base_url: string;
  task_panel_url?: string;
  project_tab_url?: string;
  project_tab_label?: string;
  events: string[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ProjectExtension extends Extension {
  installed: boolean;
  enabled: boolean;
}

export interface RegisterExtensionRequest {
  key: string;
  name: string;
  description?: string;
  base_url: string;
  shared_secret: string;
  task_panel_url?: string;
  project_tab_url?: string;
  project_tab_label?: string;
  events: string[];
}

export const EXTENSION_EVENTS = [
  { value: 'task_created', label: 'Создание задачи' },
  { value: 'task_status_changed', label: 'Смена статуса задачи' },
  { value: 'task_comment_added', label: 'Новый комментарий' },
];

export const listExtensions = async (): Promise<Extension[]> => {
  const response = await apiClient.request<{ extensions: Extension[] }>('/extensions', {
    method: 'GET',
  });
  return response.extensions || [];
};

export const registerExtension = async (data: RegisterExtensionRequest): Promise<Extension> => {
  return apiClient.request<Extension>('/extensions', {
    method: 'POST',
    body: JSON.stringify(data),
  });
};

export const deleteExtension = async (key: string): Promise<void> => {
  await apiClient.request(`/extensions/${key}`, {
    method: 'DELETE',
  });
};

export const listProjectExtensions = async (projectId: number): Promise<ProjectExtension[]> => {
  const response = await apiClient.request<{ extensions: ProjectExtension[] }>(
    `/extensions/projects/${projectId}`,
    { method: 'GET' }
  );
  return response.extensions || [];
};

export const toggleProjectExtension = async (
  projectId: number,
  key: string,
  enabled: boolean
): Promise<void> => {
  await apiClient.request(`/extensions/projects/${projectId}/${key}/toggle`, {
    method: 'POST',
    body: JSON.stringify({ enabled }),
  });
};

export const uninstallProjectExtension = async (projectId: number, key: string): Promise<void> => {
  await apiClient.request(`/extensions/projects/${projectId}/${key}`, {
    method: 'DELETE',
  });
};
