'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { apiCall, ApiError } from '@/lib/api';
import Link from 'next/link';
import { Loader2 } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { FormField } from '@/components/ui/FormField';
import { Alert } from '@/components/ui/Alert';
import { TopBar } from '@/components/ui/TopBar';
import { AppShell } from '@/components/ui/AppShell';

interface UserProfile {
  id: string;
  name: string;
  email: string;
  avatar_url: string | null;
  household: { id: string; name: string; role: string } | null;
}

export default function ProfileSettingsPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [profile, setProfile] = useState<UserProfile | null>(null);

  const [formData, setFormData] = useState({
    name: '',
    avatar_url: '',
  });

  const [passwordForm, setPasswordForm] = useState({
    old_password: '',
    new_password: '',
    confirm_password: '',
  });

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const data = await apiCall<UserProfile>('/users/me');
        setProfile(data);
        setFormData({ name: data.name, avatar_url: data.avatar_url || '' });
      } catch (err) {
        if (err instanceof ApiError && err.statusCode === 401) {
          router.push('/login');
        } else {
          setError('Gagal memuat profil');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, [router]);

  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    setUpdating(true);

    try {
      await apiCall('/users/me', {
        method: 'PATCH',
        body: JSON.stringify({
          name: formData.name,
          avatar_url: formData.avatar_url || null,
        }),
      });

      setSuccess('Profil berhasil diperbarui');
      if (profile) {
        setProfile({ ...profile, name: formData.name, avatar_url: formData.avatar_url || null });
      }
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal memperbarui profil');
      }
    } finally {
      setUpdating(false);
    }
  };

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (passwordForm.new_password !== passwordForm.confirm_password) {
      setError('Password baru tidak cocok');
      return;
    }

    setUpdating(true);

    try {
      await apiCall('/users/me/change-password', {
        method: 'POST',
        body: JSON.stringify({
          old_password: passwordForm.old_password,
          new_password: passwordForm.new_password,
        }),
      });

      setSuccess('Password berhasil diubah');
      setPasswordForm({ old_password: '', new_password: '', confirm_password: '' });
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.statusCode === 400) {
          setError('Password lama tidak sesuai');
        } else {
          setError(err.message);
        }
      } else {
        setError('Gagal mengubah password');
      }
    } finally {
      setUpdating(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-base-100">
        <div className="flex flex-col items-center gap-4">
          <Loader2 className="w-8 h-8 animate-spin text-primary" />
          <p className="text-base-content/60">Memuat profil...</p>
        </div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4 bg-base-100">
        <Alert type="error" message="Gagal memuat profil" />
      </div>
    );
  }

  return (
    <AppShell active="/settings/profile">
      <TopBar title="Pengaturan" subtitle="Kelola profil & keamanan" backHref="/dashboard" />

      <div className="max-w-4xl mx-auto p-4 sm:p-6">
        {/* Household link */}
        <div className="mb-6 sm:mb-8">
          <Link href="/settings/household">
            <Button variant="outline" fullWidth>
              Pengaturan Rumah Tangga
            </Button>
          </Link>
        </div>

        {error && <Alert type="error" message={error} />}
        {success && <Alert type="success" message={success} />}

        <div className="flex flex-col gap-6 sm:gap-8">
          {/* Profil Dasar */}
          <Card>
            <h2 className="text-xl sm:text-2xl font-bold mb-6 text-base-content">Profil Dasar</h2>

            <div className="mb-6 p-4 bg-base-100 rounded-2xl border border-base-300">
              <p className="text-xs sm:text-sm text-base-content/60 mb-1">Email (tidak bisa diubah)</p>
              <p className="font-medium text-sm sm:text-base text-base-content">{profile.email}</p>
            </div>

            {profile.household && (
              <div className="mb-6 p-4 bg-base-100 rounded-2xl border border-base-300">
                <p className="text-xs sm:text-sm text-base-content/60 mb-2">Rumah Tangga</p>
                <div>
                  <p className="font-medium text-sm sm:text-base text-base-content">{profile.household.name}</p>
                  <p className="text-xs sm:text-sm text-base-content/60 mt-1">
                    Role:{' '}
                    <span className="font-medium">
                      {profile.household.role === 'owner' ? 'Pemilik' : 'Anggota'}
                    </span>
                  </p>
                </div>
              </div>
            )}

            <form onSubmit={handleUpdateProfile} className="flex flex-col gap-5 sm:gap-6">
              <FormField label="Nama">
                <Input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  disabled={updating}
                />
              </FormField>

              <FormField label="URL Avatar" hint="Opsional">
                <Input
                  type="url"
                  placeholder="https://example.com/avatar.jpg"
                  value={formData.avatar_url}
                  onChange={(e) => setFormData({ ...formData, avatar_url: e.target.value })}
                  disabled={updating}
                />
              </FormField>

              <Button type="submit" fullWidth disabled={updating} className="mt-2">
                {updating ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Menyimpan...
                  </>
                ) : (
                  'Simpan Perubahan'
                )}
              </Button>
            </form>
          </Card>

          {/* Ubah Password */}
          <Card>
            <h2 className="text-xl sm:text-2xl font-bold mb-6 text-base-content">Keamanan</h2>

            <form onSubmit={handleChangePassword} className="flex flex-col gap-5 sm:gap-6">
              <FormField label="Password Saat Ini">
                <Input
                  type="password"
                  placeholder="••••••••"
                  required
                  value={passwordForm.old_password}
                  onChange={(e) => setPasswordForm({ ...passwordForm, old_password: e.target.value })}
                  disabled={updating}
                  autoComplete="current-password"
                />
              </FormField>

              <FormField label="Password Baru">
                <Input
                  type="password"
                  placeholder="Minimal 8 karakter"
                  required
                  minLength={8}
                  value={passwordForm.new_password}
                  onChange={(e) => setPasswordForm({ ...passwordForm, new_password: e.target.value })}
                  disabled={updating}
                  autoComplete="new-password"
                />
              </FormField>

              <FormField label="Konfirmasi Password Baru">
                <Input
                  type="password"
                  placeholder="••••••••"
                  required
                  minLength={8}
                  value={passwordForm.confirm_password}
                  onChange={(e) => setPasswordForm({ ...passwordForm, confirm_password: e.target.value })}
                  disabled={updating}
                  autoComplete="new-password"
                />
              </FormField>

              <Button type="submit" fullWidth disabled={updating} className="mt-2">
                {updating ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Mengubah...
                  </>
                ) : (
                  'Ubah Password'
                )}
              </Button>
            </form>
          </Card>
        </div>
      </div>

      </AppShell>
  );
}
