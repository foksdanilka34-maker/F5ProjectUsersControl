import { useEffect, useState } from 'react';
import { listDepartments, listPositions } from '../services/employeeService';
import type { DepartmentDTO, PositionDTO } from '../services/types';
import { ApiError } from '../lib/apiClient';

export type ReferenceState<T> = {
  items: T[];
  loading: boolean;
  error: string | null;
};

export function useEmployeeReferences() {
  const [departments, setDepartments] = useState<ReferenceState<DepartmentDTO>>({
    items: [],
    loading: true,
    error: null,
  });
  const [positions, setPositions] = useState<ReferenceState<PositionDTO>>({
    items: [],
    loading: true,
    error: null,
  });

  useEffect(() => {
    let aborted = false;

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

    fetchDepartments();
    fetchPositions();

    return () => {
      aborted = true;
    };
  }, []);

  return { departments, positions };
}

function getErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    const payload = error.payload;
    if (payload && typeof payload === 'object') {
      if ('error' in payload && payload.error) {
        return String(payload.error);
      }
      if ('message' in payload && payload.message) {
        return String(payload.message);
      }
    }
    return `Ошибка ${error.status}`;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return 'Не удалось загрузить справочник';
}
