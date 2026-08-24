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

export default function RegisterPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    password: '',
    confirmPassword: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (formData.password !== formData.confirmPassword) {
      setError('Password tidak cocok');
      return;
    }

    setLoading(true);

    try {
      const data = await apiCall<{ token: string; user: { id: string; name: string; email: string } }>(
        '/auth/register',
        {
          method: 'POST',
          body: JSON.stringify({
            name: formData.name,
            email: formData.email,
            password: formData.password,
          }),
        }
      );

      await setAuthToken(data.token);
      localStorage.setItem('token', data.token);

      router.push('/onboarding');
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.statusCode === 409) {
          setError('Email sudah terdaftar');
        } else {
          setError(err.message);
        }
      } else {
        setError('Registrasi gagal');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 sm:p-6 bg-base-100">
      <div className="w-full max-w-md">
        <div className="mb-6 sm:mb-8">
          <h1 className="text-3xl sm:text-4xl font-bold mb-2 text-base-content">Mulai sekarang</h1>
          <p className="text-sm sm:text-base text-base-content/60">Buat akun untuk kelola keuangan keluarga</p>
        </div>

        <Card>
          {error && <Alert type="error" message={error} />}

          <form onSubmit={handleSubmit} className="flex flex-col gap-5 sm:gap-6">
            <FormField label="Nama lengkap">
              <Input
                type="text"
                placeholder="Budi Santoso"
                required
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                disabled={loading}
                autoComplete="name"
              />
            </FormField>

            <FormField label="Email">
              <Input
                type="email"
                placeholder="budi@example.com"
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
                placeholder="Minimal 8 karakter"
                required
                minLength={8}
                value={formData.password}
                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                disabled={loading}
                autoComplete="new-password"
              />
            </FormField>

            <FormField label="Konfirmasi Password">
              <Input
                type="password"
                placeholder="••••••••"
                required
                minLength={8}
                value={formData.confirmPassword}
                onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
                disabled={loading}
                autoComplete="new-password"
              />
            </FormField>

            <Button type="submit" fullWidth disabled={loading} className="mt-2">
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Memproses...
                </>
              ) : (
                'Daftar'
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
              Sudah punya akun?{' '}
              <Link href="/login" className="text-primary font-medium hover:underline">
                Masuk di sini
              </Link>
            </p>
          </div>
        </Card>
      </div>
    </div>
  );
}
