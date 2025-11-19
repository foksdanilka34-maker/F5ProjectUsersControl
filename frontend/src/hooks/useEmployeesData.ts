import { useEffect, useMemo, useState } from 'react';
import type { DepartmentDTO, PositionDTO, ProfileDTO, SkillDTO } from '../services/types';
import { listDepartments, listPositions, listProfiles, listSkills } from '../services/employeeService';
import { ApiError } from '../lib/apiClient';

export type ProfilesState = {
  items: ProfileDTO[];
  total: number;
  loading: boolean;
  error: string | null;
};

export type ReferenceState<T> = {
  items: T[];
  loading: boolean;
  error: string | null;
};

export function useEmployeesData() {
  const [profiles, setProfiles] = useState<ProfilesState>({ items: [], total: 0, loading: true, error: null });
  const [departments, setDepartments] = useState<ReferenceState<DepartmentDTO>>({ items: [], loading: true, error: null });
  const [positions, setPositions] = useState<ReferenceState<PositionDTO>>({ items: [], loading: true, error: null });
  const [skills, setSkills] = useState<ReferenceState<SkillDTO>>({ items: [], loading: true, error: null });

  useEffect(() => {
    let aborted = false;

    const fetchProfiles = async () => {
      setProfiles((prev) => ({ ...prev, loading: true, error: null }));
      try {
        const response = await listProfiles();
        const payload = Array.isArray(response.data)
          ? { profiles: response.data, total_count: response.data.length }
          : response.data;
        if (!payload) throw new Error('Пустой ответ ListProfiles');
        if (!aborted) {
          setProfiles({ items: payload.profiles ?? [], total: payload.total_count ?? 0, loading: false, error: null });
        }
      } catch (error) {
        if (!aborted) {
          setProfiles((prev) => ({ ...prev, loading: false, error: getErrorMessage(error) }));
        }
      }
    };

    const fetchDepartments = async () => {
      setDepartments((prev) => ({ ...prev, loading: true, error: null }));
      try {
        const response = await listDepartments();
        const items = response.data?.departments ?? [];
        if (!aborted) {
          setDepartments({ items, loading: false, error: null });
        }
      } catch (error) {
        if (!aborted) {
          setDepartments((prev) => ({ ...prev, loading: false, error: getErrorMessage(error) }));
        }
      }
    };

    const fetchPositions = async () => {
      setPositions((prev) => ({ ...prev, loading: true, error: null }));
      try {
        const response = await listPositions();
        const items = response.data?.positions ?? [];
        if (!aborted) {
          setPositions({ items, loading: false, error: null });
        }
      } catch (error) {
        if (!aborted) {
          setPositions((prev) => ({ ...prev, loading: false, error: getErrorMessage(error) }));
        }
      }
    };

    const fetchSkills = async () => {
      setSkills((prev) => ({ ...prev, loading: true, error: null }));
      try {
        const response = await listSkills();
        const items = response.data?.skills ?? [];
        if (!aborted) {
          setSkills({ items, loading: false, error: null });
        }
      } catch (error) {
        if (!aborted) {
          setSkills((prev) => ({ ...prev, loading: false, error: getErrorMessage(error) }));
        }
      }
    };

    fetchProfiles();
    fetchDepartments();
    fetchPositions();
    fetchSkills();

    return () => {
      aborted = true;
    };
  }, []);

  const stats = useMemo(
    () => [
      { label: 'Профилей в базе', value: profiles.total },
      { label: 'Активные', value: profiles.items.filter((profile) => profile.is_active).length },
      { label: 'Отделов', value: departments.items.length },
      { label: 'Навыков', value: skills.items.length },
    ],
    [profiles.items, profiles.total, departments.items.length, skills.items.length],
  );

  return { profiles, departments, positions, skills, stats };
}

function getErrorMessage(error: unknown) {
  if (error instanceof ApiError) {
    const payloadMessage =
      error.payload && typeof error.payload === 'object' && 'error' in error.payload
        ? String((error.payload as Record<string, unknown>).error)
        : undefined;
    return payloadMessage ?? `Ошибка ${error.status}`;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return 'Не удалось загрузить данные';
}
