'use client';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
}

class ApiError extends Error {
  constructor(public statusCode: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

export async function apiCall<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_URL}${endpoint}`;
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  if (options.headers && typeof options.headers === 'object') {
    Object.assign(headers, options.headers);
  }

  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(url, {
    ...options,
    headers,
    credentials: 'include',
  });

  const json = (await response.json()) as ApiResponse<T>;

  if (!response.ok) {
    throw new ApiError(response.status, json.error || 'Request failed');
  }

  if (!json.success) {
    throw new ApiError(response.status, json.error || 'Request failed');
  }

  return json.data as T;
}

export { ApiError };
