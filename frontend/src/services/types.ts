export type ApiResponse<T> = {
  success: boolean;
  message?: string;
  data: T;
  error?: string;
  meta?: PaginationMeta;
};

export type PaginatedMeta = {
  page_size: number;
  page_number: number;
  total_count?: number;
};

export type PaginationMeta = PaginatedMeta;

export type ProfileDTO = {
  id: string;
  first_name: string;
  last_name: string;
  position_id: string;
  email: string;
  avatar_url?: string;
  hire_date: string;
  department?: DepartmentDTO;
  skills?: SkillDTO[];
  login: string;
  role: string;
  created_at: string;
  updated_at: string;
  is_active: boolean;
};

export type DepartmentDTO = {
  id: string;
  name: string;
};

export type SkillDTO = {
  id: string;
  name: string;
};

export type PositionDTO = {
  id: string;
  name: string;
};

export type ListProfilesResponse = ApiResponse<{ profiles: ProfileDTO[]; total_count: number; } | ProfileDTO[]>;
export type ListDepartmentsResponse = ApiResponse<{ departments: DepartmentDTO[] }>;
export type ListPositionsResponse = ApiResponse<{ positions: PositionDTO[] }>;
export type ListSkillsResponse = ApiResponse<{ skills: SkillDTO[] }>;

export type CreateProfileRequest = {
  first_name: string;
  last_name: string;
  position_id: string;
  email: string;
  department_id?: string;
  hire_date: string;
  login: string;
  password: string;
  role: string;
};

export type CreateProfileResponse = ApiResponse<ProfileDTO>;
