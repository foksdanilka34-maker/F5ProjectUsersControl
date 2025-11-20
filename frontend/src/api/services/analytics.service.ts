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

export const analyticsService = {
  // ===== Dashboard =====
  async getDashboardStats(params?: DashboardStatsParams): Promise<DashboardStats> {
    const response = await apiClient.get<ApiResponse<DashboardStats>>(
      '/api/v1/analytics/dashboard',
      { params }
    );
    return unwrapResponse(response.data);
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
          const data = JSON.parse(event.data) as DashboardStats;
          onMessage(data);
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
    const response = await apiClient.get<ApiResponse<EmployeeMetrics[]>>(
      '/api/v1/analytics/employees/metrics',
      { params }
    );
    return {
      metrics: unwrapResponse(response.data),
      meta: response.data.meta,
    };
  },

  async getEmployeeMetrics(employeeId: string): Promise<EmployeeMetrics> {
    const response = await apiClient.get<ApiResponse<EmployeeMetrics>>(
      `/api/v1/analytics/employees/${employeeId}/metrics`
    );
    return unwrapResponse(response.data);
  },

  async getTopPerformers(params?: TopPerformersParams): Promise<TopPerformer[]> {
    const response = await apiClient.get<ApiResponse<TopPerformer[]>>(
      '/api/v1/analytics/employees/top-performers',
      { params }
    );
    return unwrapResponse(response.data);
  },

  // ===== Project Metrics =====
  async listProjectMetrics(
    params?: ListProjectMetricsParams
  ): Promise<{ metrics: ProjectMetrics[]; meta?: PaginationMeta }> {
    const response = await apiClient.get<ApiResponse<ProjectMetrics[]>>(
      '/api/v1/analytics/projects/metrics',
      { params }
    );
    return {
      metrics: unwrapResponse(response.data),
      meta: response.data.meta,
    };
  },

  async getProjectMetrics(
    projectId: string,
    params?: { start_date?: string; end_date?: string }
  ): Promise<ProjectMetrics> {
    const response = await apiClient.get<ApiResponse<ProjectMetrics>>(
      `/api/v1/analytics/projects/${projectId}/metrics`,
      { params }
    );
    return unwrapResponse(response.data);
  },

  // ===== Trends =====
  async getProductivityTrends(params?: ProductivityTrendsParams): Promise<ProductivityTrend[]> {
    const response = await apiClient.get<ApiResponse<ProductivityTrend[]>>(
      '/api/v1/analytics/trends/productivity',
      { params }
    );
    return unwrapResponse(response.data);
  },

  async getCompletionRateTrends(params?: CompletionRateTrendsParams): Promise<CompletionRateTrend[]> {
    const response = await apiClient.get<ApiResponse<CompletionRateTrend[]>>(
      '/api/v1/analytics/trends/completion-rate',
      { params }
    );
    return unwrapResponse(response.data);
  },
};
