// Client-side token management via API route handler
export async function setAuthToken(token: string) {
  const response = await fetch('/api/auth', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token }),
  });
  return response.ok;
}

export async function clearAuthToken() {
  await fetch('/api/auth', { method: 'DELETE' });
}
