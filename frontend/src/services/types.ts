
export type PaginatedMeta = {
  page_size: number;
  page_number: number;
  total_count?: number;
};

export type PaginationMeta = PaginatedMeta;

export type ProfileDTO = {
  id: number;
  first_name: string;
  last_name: string;
  position_id: number;
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
  id: number;
  name: string;
};

export type SkillDTO = {
  id: number;
  name: string;
};

export type PositionDTO = {
  id: number;
  name: string;
};

export type ListProfilesResponse = {
  profiles: ProfileDTO[];
  total_count: number;
};

export type ListDepartmentsResponse = {
  departments: DepartmentDTO[];
};

export type ListPositionsResponse = {
  positions: PositionDTO[];
};

export type ListSkillsResponse = {
  skills: SkillDTO[];
};

export type CreateProfileRequest = {
  first_name: string;
  last_name: string;
  position_id: number;
  email: string;
  department_id?: number;
  hire_date: string;
  login: string;
  password: string;
  role: string;
};

export type CreateProfileResponse = ProfileDTO;

export type UpdateProfileRequest = Partial<CreateProfileRequest> & {
  avatar_url?: string | null;
};


