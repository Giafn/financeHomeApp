'use client';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'https://api-finance.giafn.my.id/api/v1';

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

// downloadFile fetches a binary response (PDF/Excel export, dst) and triggers a browser
// download — apiCall() assumes JSON envelope, tidak cocok untuk endpoint file.
export async function downloadFile(endpoint: string): Promise<void> {
  const url = `${API_URL}${endpoint}`;
  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;

  const response = await fetch(url, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    credentials: 'include',
  });

  if (!response.ok) {
    const json = await response.json().catch(() => ({}));
    throw new ApiError(response.status, json.error || 'Gagal mengunduh file');
  }

  const disposition = response.headers.get('Content-Disposition') || '';
  const match = disposition.match(/filename="(.+)"/);
  const filename = match ? match[1] : 'download';

  const blob = await response.blob();
  const blobUrl = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = blobUrl;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  window.URL.revokeObjectURL(blobUrl);
}

export { ApiError };
