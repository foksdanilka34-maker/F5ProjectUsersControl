import { apiClient, unwrapResponse } from '../client';
import type {
  Profile,
  CreateProfileRequest,
  UpdateProfileRequest,
  ChangeUserStatusRequest,
  Department,
  CreateDepartmentRequest,
  UpdateDepartmentRequest,
  Position,
  CreatePositionRequest,
  UpdatePositionRequest,
  Skill,
  CreateSkillRequest,
  AddSkillToEmployeeRequest,
  ListProfilesParams,
  ApiResponse,
  PaginationMeta,
} from '../types';

export const employeeService = {
  // ===== Profiles =====
  async createProfile(data: CreateProfileRequest): Promise<Profile> {
    const response = await apiClient.post<ApiResponse<Profile>>(
      '/api/v1/employees/profiles',
      data
    );
    return unwrapResponse(response.data);
  },

  async listProfiles(params?: ListProfilesParams): Promise<{ profiles: Profile[]; meta?: PaginationMeta }> {
    const response = await apiClient.get<ApiResponse<any>>(
      '/api/v1/employees/profiles',
      { params }
    );
    const data = unwrapResponse(response.data);
    return {
      profiles: data.profiles || [],
      meta: response.data.meta,
    };
  },

  async getProfile(id: string): Promise<Profile> {
    const response = await apiClient.get<ApiResponse<Profile>>(
      `/api/v1/employees/profiles/${id}`
    );
    return unwrapResponse(response.data);
  },

  async updateProfile(id: string, data: UpdateProfileRequest): Promise<Profile> {
    const response = await apiClient.patch<ApiResponse<Profile>>(
      `/api/v1/employees/profiles/${id}`,
      data
    );
    return unwrapResponse(response.data);
  },

  async deleteProfile(id: string): Promise<void> {
    const response = await apiClient.delete<ApiResponse>(
      `/api/v1/employees/profiles/${id}`
    );
    unwrapResponse(response.data);
  },

  async changeUserStatus(id: string, data: ChangeUserStatusRequest): Promise<void> {
    const response = await apiClient.patch<ApiResponse>(
      `/api/v1/employees/profiles/${id}/status`,
      data
    );
    unwrapResponse(response.data);
  },

  // ===== Departments =====
  async createDepartment(data: CreateDepartmentRequest): Promise<Department> {
    const response = await apiClient.post<ApiResponse<Department>>(
      '/api/v1/employees/departments',
      data
    );
    return unwrapResponse(response.data);
  },

  async listDepartments(): Promise<Department[]> {
    const response = await apiClient.get<ApiResponse<any>>(
      '/api/v1/employees/departments'
    );
    const data = unwrapResponse(response.data);
    return data.departments || [];
  },

  async getDepartment(id: string): Promise<Department> {
    const response = await apiClient.get<ApiResponse<Department>>(
      `/api/v1/employees/departments/${id}`
    );
    return unwrapResponse(response.data);
  },

  async updateDepartment(id: string, data: UpdateDepartmentRequest): Promise<Department> {
    const response = await apiClient.put<ApiResponse<Department>>(
      `/api/v1/employees/departments/${id}`,
      data
    );
    return unwrapResponse(response.data);
  },

  async deleteDepartment(id: string): Promise<void> {
    const response = await apiClient.delete<ApiResponse>(
      `/api/v1/employees/departments/${id}`
    );
    unwrapResponse(response.data);
  },

  // ===== Positions =====
  async createPosition(data: CreatePositionRequest): Promise<Position> {
    const response = await apiClient.post<ApiResponse<Position>>(
      '/api/v1/employees/positions',
      data
    );
    return unwrapResponse(response.data);
  },

  async listPositions(): Promise<Position[]> {
    const response = await apiClient.get<ApiResponse<any>>(
      '/api/v1/employees/positions'
    );
    const data = unwrapResponse(response.data);
    return data.positions || [];
  },

  async getPosition(id: string): Promise<Position> {
    const response = await apiClient.get<ApiResponse<Position>>(
      `/api/v1/employees/positions/${id}`
    );
    return unwrapResponse(response.data);
  },

  async updatePosition(id: string, data: UpdatePositionRequest): Promise<Position> {
    const response = await apiClient.put<ApiResponse<Position>>(
      `/api/v1/employees/positions/${id}`,
      data
    );
    return unwrapResponse(response.data);
  },

  async deletePosition(id: string): Promise<void> {
    const response = await apiClient.delete<ApiResponse>(
      `/api/v1/employees/positions/${id}`
    );
    unwrapResponse(response.data);
  },

  // ===== Skills =====
  async createSkill(data: CreateSkillRequest): Promise<Skill> {
    const response = await apiClient.post<ApiResponse<Skill>>(
      '/api/v1/employees/skills',
      data
    );
    return unwrapResponse(response.data);
  },

  async listSkills(): Promise<Skill[]> {
    const response = await apiClient.get<ApiResponse<any>>(
      '/api/v1/employees/skills'
    );
    const data = unwrapResponse(response.data);
    return data.skills || [];
  },

  async addSkillToEmployee(employeeId: string, data: AddSkillToEmployeeRequest): Promise<void> {
    const response = await apiClient.post<ApiResponse>(
      `/api/v1/employees/profiles/${employeeId}/skills`,
      data
    );
    unwrapResponse(response.data);
  },

  async removeSkillFromEmployee(employeeId: string, skillId: string): Promise<void> {
    const response = await apiClient.delete<ApiResponse>(
      `/api/v1/employees/profiles/${employeeId}/skills/${skillId}`
    );
    unwrapResponse(response.data);
  },

  async deleteSkill(id: string): Promise<void> {
    const response = await apiClient.delete<ApiResponse>(
      `/api/v1/employees/skills/${id}`
    );
    unwrapResponse(response.data);
  },

  async updateSkill(id: string, data: { name: string }): Promise<Skill> {
    const response = await apiClient.put<ApiResponse<Skill>>(
      `/api/v1/employees/skills/${id}`,
      data
    );
    return unwrapResponse(response.data);
  },
};
