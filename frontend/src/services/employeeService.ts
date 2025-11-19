import { apiClient } from '../lib/apiClient';
import type {
  CreateProfileRequest,
  CreateProfileResponse,
  ListDepartmentsResponse,
  ListPositionsResponse,
  ListProfilesResponse,
  ListSkillsResponse,
} from './types';

const EMPLOYEES_PREFIX = '/employees';

export type ListProfilesQuery = {
  pageSize?: number;
  pageNumber?: number;
  departmentId?: string;
  positionId?: string;
};

const toQuery = (query: ListProfilesQuery = {}) => {
  const params = new URLSearchParams();
  if (query.pageSize) params.set('page_size', String(query.pageSize));
  if (query.pageNumber) params.set('page_number', String(query.pageNumber));
  if (query.departmentId) params.set('department_id', query.departmentId);
  if (query.positionId) params.set('position_id', query.positionId);
  return params.toString();
};

export function listProfiles(query?: ListProfilesQuery) {
  const qs = toQuery(query);
  const path = qs ? `${EMPLOYEES_PREFIX}/profiles?${qs}` : `${EMPLOYEES_PREFIX}/profiles`;
  return apiClient.request<ListProfilesResponse>(path);
}

export function listDepartments() {
  return apiClient.request<ListDepartmentsResponse>(`${EMPLOYEES_PREFIX}/departments`);
}

export function listPositions() {
  return apiClient.request<ListPositionsResponse>(`${EMPLOYEES_PREFIX}/positions`);
}

export function listSkills() {
  return apiClient.request<ListSkillsResponse>(`${EMPLOYEES_PREFIX}/skills`);
}

export function createProfile(payload: CreateProfileRequest) {
  return apiClient.request<CreateProfileResponse>(`${EMPLOYEES_PREFIX}/profiles`, {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}
