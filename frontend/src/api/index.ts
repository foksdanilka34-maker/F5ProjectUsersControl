// Re-export all services for convenient imports
export { authService } from './services/auth.service';
export { employeeService } from './services/employee.service';
export { projectService } from './services/project.service';
export { analyticsService } from './services/analytics.service';

// Re-export types
export * from './types';

// Re-export client utilities
export { apiClient, setAccessToken, getAccessToken } from './client';
