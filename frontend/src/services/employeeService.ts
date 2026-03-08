import { apiClient } from '../lib/apiClient';
import type {
  CreateProfileRequest,
  CreateProfileResponse,
  ListDepartmentsResponse,
  ListPositionsResponse,
  ListProfilesResponse,
  ListSkillsResponse,
  ProfileDTO,
  DepartmentDTO,
  PositionDTO,
  SkillDTO,
  UpdateProfileRequest,
} from './types';

const PROFILES_PATH = '/profiles';
const DEPARTMENTS_PATH = '/departments';
const POSITIONS_PATH = '/positions';
const SKILLS_PATH = '/skills';

export type ListProfilesQuery = {
  pageSize?: number;
  pageNumber?: number;
  departmentId?: number;
  positionId?: number;
};

const toQuery = (query: ListProfilesQuery = {}) => {
  const params = new URLSearchParams();
  if (query.pageSize) params.set('page_size', String(query.pageSize));
  if (query.pageNumber) params.set('page_number', String(query.pageNumber));
  if (query.departmentId) params.set('department_id', String(query.departmentId));
  if (query.positionId) params.set('position_id', String(query.positionId));
  return params.toString();
};

export function listProfiles(query?: ListProfilesQuery) {
  const qs = toQuery(query);
  const path = qs ? `${PROFILES_PATH}?${qs}` : PROFILES_PATH;
  return apiClient.request<ListProfilesResponse>(path);
}

export function getProfile(id: number) {
  return apiClient.request<ProfileDTO>(`${PROFILES_PATH}/${id}`);
}

export function createProfile(payload: CreateProfileRequest) {
  return apiClient.request<CreateProfileResponse>(PROFILES_PATH, {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateProfile(id: number, payload: Record<string, unknown>) {
  return apiClient.request<ProfileDTO>(`${PROFILES_PATH}/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export function deleteProfile(id: number) {
  return apiClient.request(`${PROFILES_PATH}/${id}`, { method: 'DELETE' });
}

export function listDepartments() {
  return apiClient.request<ListDepartmentsResponse>(DEPARTMENTS_PATH);
}

export function createDepartment(name: string) {
  return apiClient.request<DepartmentDTO>(DEPARTMENTS_PATH, {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

export function updateDepartment(id: number, name: string) {
  return apiClient.request<DepartmentDTO>(`${DEPARTMENTS_PATH}/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name }),
  });
}

export function deleteDepartment(id: number) {
  return apiClient.request(`${DEPARTMENTS_PATH}/${id}`, { method: 'DELETE' });
}

export function listPositions() {
  return apiClient.request<ListPositionsResponse>(POSITIONS_PATH);
}

export function createPosition(name: string) {
  return apiClient.request<PositionDTO>(POSITIONS_PATH, {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

export function updatePosition(id: number, name: string) {
  return apiClient.request<PositionDTO>(`${POSITIONS_PATH}/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name }),
  });
}

export function deletePosition(id: number) {
  return apiClient.request(`${POSITIONS_PATH}/${id}`, { method: 'DELETE' });
}

export function listSkills() {
  return apiClient.request<ListSkillsResponse>(SKILLS_PATH);
}

export function createSkill(name: string) {
  return apiClient.request<SkillDTO>(SKILLS_PATH, {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

export function deleteSkill(id: number) {
  return apiClient.request(`${SKILLS_PATH}/${id}`, {
    method: 'DELETE',
  });
}

export function addSkillToProfile(profileId: number, skillId: number) {
  return apiClient.request(`${PROFILES_PATH}/${profileId}/skills`, {
    method: 'POST',
    body: JSON.stringify({ skill_id: skillId }),
  });
}

export function removeSkillFromProfile(profileId: number, skillId: number) {
  return apiClient.request(`${PROFILES_PATH}/${profileId}/skills/${skillId}`, {
    method: 'DELETE',
  });
}


