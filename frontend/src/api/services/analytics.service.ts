import { apiClient, unwrapResponse } from '../client';
import type {
  DashboardStats,
  EmployeeMetrics,
  ProjectMetrics,
  TopPerformer,
  ProductivityTrend,
  CompletionRateTrend,
  DashboardStatsParams,
  ListEmployeeMetricsParams,
  TopPerformersParams,
  ListProjectMetricsParams,
  ProductivityTrendsParams,
  CompletionRateTrendsParams,
  ApiResponse,
  PaginationMeta,
} from '../types';

// Helper to map backend EmployeeMetrics to frontend EmployeeMetrics
const mapEmployeeMetrics = (backendMetrics: any): EmployeeMetrics => {
  if (!backendMetrics) return {} as EmployeeMetrics;
  
  const completed = backendMetrics.completed_tasks || 0;
  const totalDurationSeconds = backendMetrics.total_task_duration_seconds || 0;
  const avgHours = completed > 0 ? (totalDurationSeconds / completed) / 3600 : 0;

  return {
    employee_id: backendMetrics.employee_id,
    employee_name: backendMetrics.employee_name || '', // Backend might not send name, need to handle or fetch separately?
    department_name: backendMetrics.department_name,
    total_tasks: backendMetrics.assigned_tasks || 0,
    completed_tasks: backendMetrics.completed_tasks || 0,
    in_progress_tasks: backendMetrics.in_progress_tasks || 0,
    overdue_tasks: backendMetrics.overdue_tasks || 0,
    completion_rate: backendMetrics.TaskCompletionRate || 0,
    average_completion_time_hours: avgHours,
    productivity_score: backendMetrics.EfficiencyScore || 0,
  };
};

// Helper to map backend ProjectMetrics to frontend ProjectMetrics
const mapProjectMetrics = (backendMetrics: any): ProjectMetrics => {
  if (!backendMetrics) return {} as ProjectMetrics;

  const completed = backendMetrics.completed_tasks || 0;
  const totalDurationSeconds = backendMetrics.total_task_duration_seconds_completed || 0;
  const avgHours = completed > 0 ? (totalDurationSeconds / completed) / 3600 : 0;

  return {
    project_id: backendMetrics.project_id,
    project_name: backendMetrics.project_name || '',
    manager_name: backendMetrics.manager_name || '',
    total_tasks: backendMetrics.total_tasks || 0,
    completed_tasks: backendMetrics.completed_tasks || 0,
    in_progress_tasks: backendMetrics.in_progress_tasks || 0,
    overdue_tasks: backendMetrics.overdue_tasks || 0,
    completion_rate: backendMetrics.delivery_performance || 0, // Assuming delivery_performance maps to completion_rate
    on_time_completion_rate: backendMetrics.schedule_performance || 0,
    average_task_duration_hours: avgHours,
    health_score: backendMetrics.health_index || 0,
  };
};

// Helper to map backend DashboardStats to frontend DashboardStats
const mapDashboardStats = (backendStats: any): DashboardStats => {
  if (!backendStats) return {} as DashboardStats;
  return {
    total_employees: backendStats.total_employees || 0,
    active_employees: backendStats.active_employees || 0,
    total_projects: backendStats.total_projects || 0,
    active_projects: backendStats.active_projects || 0,
    total_tasks: backendStats.total_tasks || 0,
    completed_tasks: backendStats.completed_tasks || 0,
    overdue_tasks: backendStats.overdue_tasks || 0,
    completion_rate: backendStats.avg_on_time_rate || 0,
    average_productivity: backendStats.avg_company_efficiency || 0,
  };
};

export const analyticsService = {
  // ===== Dashboard =====
  async getDashboardStats(params?: DashboardStatsParams): Promise<DashboardStats> {
    const response = await apiClient.get<ApiResponse<any>>(
      '/api/v1/analytics/dashboard',
      { params }
    );
    const data = unwrapResponse(response.data);
    return mapDashboardStats(data);
  },

  /**
   * Stream dashboard stats via Server-Sent Events
   * Returns EventSource for real-time updates
   */
  streamDashboardStats(
    params?: DashboardStatsParams & { interval_seconds?: number },
    onMessage?: (data: DashboardStats) => void,
    onError?: (error: Event) => void
  ): EventSource {
    const searchParams = new URLSearchParams();
    if (params?.start_date) searchParams.append('start_date', params.start_date);
    if (params?.end_date) searchParams.append('end_date', params.end_date);
    if (params?.interval_seconds) searchParams.append('interval_seconds', params.interval_seconds.toString());

    const url = `${apiClient.defaults.baseURL}/api/v1/analytics/dashboard/stream?${searchParams.toString()}`;
    const eventSource = new EventSource(url);

    if (onMessage) {
      eventSource.onmessage = (event) => {
        try {
          const parsed = JSON.parse(event.data);
          // SSE usually wraps data. Check structure.
          // Assuming parsed.data is the stats object if using standard response wrapper
          const statsData = parsed.data || parsed; 
          onMessage(mapDashboardStats(statsData));
        } catch (error) {
          console.error('Failed to parse SSE message:', error);
        }
      };
    }

    if (onError) {
      eventSource.onerror = onError;
    }

    return eventSource;
  },

  // ===== Employee Metrics =====
  async listEmployeeMetrics(
    params?: ListEmployeeMetricsParams
  ): Promise<{ metrics: EmployeeMetrics[]; meta?: PaginationMeta }> {
    const response = await apiClient.get<ApiResponse<any>>(
      '/api/v1/analytics/employees/metrics',
      { params }
    );
    const data = unwrapResponse(response.data);
    // data.metrics is the array
    const metrics = (data.metrics || []).map(mapEmployeeMetrics);
    return {
      metrics,
      meta: response.data.meta,
    };
  },

  async getEmployeeMetrics(employeeId: string): Promise<EmployeeMetrics> {
    const response = await apiClient.get<ApiResponse<any>>(
      `/api/v1/analytics/employees/${employeeId}/metrics`
    );
    const data = unwrapResponse(response.data);
    // data.metrics is the object inside the response
    return mapEmployeeMetrics(data.metrics);
  },

  async getTopPerformers(params?: TopPerformersParams): Promise<TopPerformer[]> {
    const response = await apiClient.get<ApiResponse<any>>(
      '/api/v1/analytics/employees/top-performers',
      { params }
    );
    const data = unwrapResponse(response.data);
    // TopPerformersResponse has TopEmployees field?
    // Let's check proto: TopEmployees []*TopEmployee
    // JSON: top_employees
    
    // Wait, GetTopPerformers returns TopPerformersResponse?
    // No, proto says GetTopPerformers returns ListEmployeeMetricsResponse?
    // No, proto says GetTopPerformers returns DashboardStatsResponse?
    // Let's check proto again.
    
    // It seems I need to check the proto for GetTopPerformers response type.
    // Assuming it returns a list of TopPerformer objects for now based on previous code.
    // But previous code was just casting.
    
    // Actually, let's look at the proto file again.
    // GetTopPerformers returns ListEmployeeMetricsResponse? No.
    
    // Found it in proto:
    // rpc GetTopPerformers (GetTopPerformersRequest) returns (ListEmployeeMetricsResponse);
    
    // So it returns ListEmployeeMetricsResponse, which has `metrics` field (array of EmployeeMetrics).
    // So we should map it using mapEmployeeMetrics.
    
    // But the frontend expects TopPerformer interface:
    // interface TopPerformer { employee_id, employee_name, productivity_score, rank, ... }
    
    // The backend returns EmployeeMetrics.
    // We need to map EmployeeMetrics to TopPerformer.
    
    const metrics = (data.metrics || []).map(mapEmployeeMetrics);
    return metrics.map((m: EmployeeMetrics, index: number) => ({
        employee_id: m.employee_id,
        employee_name: m.employee_name || 'Unknown',
        department_name: m.department_name,
        completed_tasks: m.completed_tasks,
        productivity_score: m.productivity_score,
        rank: index + 1
    }));
  },

  // ===== Project Metrics =====
  async listProjectMetrics(
    params?: ListProjectMetricsParams
  ): Promise<{ metrics: ProjectMetrics[]; meta?: PaginationMeta }> {
    const response = await apiClient.get<ApiResponse<any>>(
      '/api/v1/analytics/projects/metrics',
      { params }
    );
    const data = unwrapResponse(response.data);
    const metrics = (data.metrics || []).map(mapProjectMetrics);
    return {
      metrics,
      meta: response.data.meta,
    };
  },

  async getProjectMetrics(
    projectId: string,
    params?: { start_date?: string; end_date?: string }
  ): Promise<ProjectMetrics> {
    const response = await apiClient.get<ApiResponse<any>>(
      `/api/v1/analytics/projects/${projectId}/metrics`,
      { params }
    );
    const data = unwrapResponse(response.data);
    return mapProjectMetrics(data.metrics);
  },

  // ===== Trends =====
  async getProductivityTrends(params?: ProductivityTrendsParams): Promise<ProductivityTrend[]> {
    const response = await apiClient.get<ApiResponse<any>>(
      '/api/v1/analytics/trends/productivity',
      { params }
    );
    const data = unwrapResponse(response.data);
    // Proto: ProductivityTrendsResponse { entries: [...] }
    return (data.entries || []).map((entry: any) => ({
        period: 'DAILY', // Default or from params
        date: entry.date,
        productivity_score: entry.avg_efficiency || 0,
        completed_tasks: entry.total_tasks_completed || 0,
        total_tasks: 0 // Not in trend entry
    }));
  },

  async getCompletionRateTrends(params?: CompletionRateTrendsParams): Promise<CompletionRateTrend[]> {
    const response = await apiClient.get<ApiResponse<any>>(
      '/api/v1/analytics/trends/completion-rate',
      { params }
    );
    const data = unwrapResponse(response.data);
    // Proto: CompletionRateTrendsResponse { entries: [...] }
    return (data.entries || []).map((entry: any) => ({
        period: 'WEEKLY',
        date: entry.date,
        completion_rate: entry.overall_rate || 0,
        on_time_rate: entry.on_time_rate || 0,
        total_completed: entry.completed_count || 0
    }));
  },
};
