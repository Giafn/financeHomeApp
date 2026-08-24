'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Loader2 } from 'lucide-react';
import { apiCall, ApiError } from '@/lib/api';
import { setAuthToken } from '@/lib/auth';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { FormField } from '@/components/ui/FormField';
import { Alert } from '@/components/ui/Alert';

export default function LoginPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [formData, setFormData] = useState({ email: '', password: '' });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const data = await apiCall<{ token: string; user: { id: string; name: string; email: string } }>(
        '/auth/login',
        {
          method: 'POST',
          body: JSON.stringify(formData),
        }
      );

      await setAuthToken(data.token);
      localStorage.setItem('token', data.token);

      const profile = await apiCall<{
        household: { id: string; name: string; role: string } | null;
      }>('/users/me');

      if (profile.household) {
        router.push('/dashboard');
      } else {
        router.push('/onboarding');
      }
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.statusCode === 401 ? 'Email atau password salah' : err.message);
      } else {
        setError('Login gagal');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 sm:p-6 bg-base-100">
      <div className="w-full max-w-md">
        <div className="mb-6 sm:mb-8">
          <h1 className="text-3xl sm:text-4xl font-bold mb-2 text-base-content">Welcome back</h1>
          <p className="text-sm sm:text-base text-base-content/60">Kelola keuangan keluarga Anda dengan mudah</p>
        </div>

        <Card>
          {error && <Alert type="error" message={error} />}

          <form onSubmit={handleSubmit} className="flex flex-col gap-5 sm:gap-6">
            <FormField label="Email">
              <Input
                type="email"
                placeholder="nama@example.com"
                required
                value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                disabled={loading}
                autoComplete="email"
              />
            </FormField>

            <FormField label="Password">
              <Input
                type="password"
                placeholder="••••••••"
                required
                minLength={8}
                value={formData.password}
                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                disabled={loading}
                autoComplete="current-password"
              />
            </FormField>

            <Button type="submit" fullWidth disabled={loading} className="mt-2">
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Memproses...
                </>
              ) : (
                'Masuk'
              )}
            </Button>
          </form>

          <div className="flex items-center gap-4 my-6">
            <div className="h-px flex-1 bg-base-300" />
            <span className="text-xs text-base-content/40">atau</span>
            <div className="h-px flex-1 bg-base-300" />
          </div>

          <div className="text-center">
            <p className="text-base-content/60 text-xs sm:text-sm">
              Belum punya akun?{' '}
              <Link href="/register" className="text-primary font-medium hover:underline">
                Daftar sekarang
              </Link>
            </p>
          </div>
        </Card>
      </div>
    </div>
  );
}
