'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { apiCall, ApiError } from '@/lib/api';
import Link from 'next/link';
import { Loader2, Upload, X } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { FormField } from '@/components/ui/FormField';
import { Alert } from '@/components/ui/Alert';
import { TopBar } from '@/components/ui/TopBar';
import { AppShell } from '@/components/ui/AppShell';

const MAX_AVATAR_MB = 5;

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
  const [uploading, setUploading] = useState(false);
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

  const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;

    if (!file.type.startsWith('image/')) {
      setError('Hanya file gambar yang diperbolehkan');
      return;
    }
    if (file.size > MAX_AVATAR_MB * 1024 * 1024) {
      setError(`Ukuran avatar maksimal ${MAX_AVATAR_MB} MB`);
      return;
    }

    setError('');
    setSuccess('');
    setUploading(true);

    try {
      const presign = await apiCall<{ upload_url: string; file_url: string }>('/uploads/presign', {
        method: 'POST',
        body: JSON.stringify({ filename: file.name, content_type: file.type }),
      });

      const res = await fetch(presign.upload_url, {
        method: 'PUT',
        headers: { 'Content-Type': file.type },
        body: file,
      });
      if (!res.ok) {
        setError('Gagal mengunggah avatar');
        setUploading(false);
        return;
      }

      setFormData((prev) => ({ ...prev, avatar_url: presign.file_url }));
      setSuccess('Avatar berhasil diunggah. Klik "Simpan Perubahan" untuk menyimpan.');
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.statusCode === 400 ? err.message : 'Gagal mengunggah avatar');
      } else {
        setError('Gagal mengunggah avatar');
      }
    } finally {
      setUploading(false);
    }
  };

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
              Pengaturan Anggota
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
                <p className="text-xs sm:text-sm text-base-content/60 mb-2">Anggota</p>
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

              <FormField label="Avatar" hint={`Hanya gambar, maksimal ${MAX_AVATAR_MB} MB`}>
                <div className="flex items-center gap-4">
                  <div className="w-16 h-16 rounded-full overflow-hidden bg-base-200 border border-base-300 flex items-center justify-center flex-shrink-0">
                    {formData.avatar_url ? (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img src={formData.avatar_url} alt="Avatar" className="w-full h-full object-cover" />
                    ) : (
                      <Loader2 className="w-6 h-6 text-base-content/40" />
                    )}
                  </div>

                  <div className="flex flex-col gap-2">
                    <label
                      className={`inline-flex items-center justify-center gap-2 rounded-full border border-base-300 text-base-content px-4 h-10 text-sm font-semibold transition-all active:scale-[0.97] hover:bg-base-200 ${
                        uploading ? 'opacity-40 cursor-not-allowed' : 'cursor-pointer'
                      }`}
                    >
                      {uploading ? (
                        <>
                          <Loader2 className="w-4 h-4 animate-spin" />
                          Mengunggah...
                        </>
                      ) : (
                        <>
                          <Upload className="w-4 h-4" />
                          Pilih Gambar
                        </>
                      )}
                      <input
                        type="file"
                        accept="image/*"
                        className="hidden"
                        disabled={uploading}
                        onChange={handleAvatarUpload}
                      />
                    </label>
                    {formData.avatar_url && (
                      <button
                        type="button"
                        onClick={() => setFormData({ ...formData, avatar_url: '' })}
                        className="inline-flex items-center gap-1 text-xs text-error font-medium"
                      >
                        <X className="w-3 h-3" />
                        Hapus avatar
                      </button>
                    )}
                  </div>
                </div>
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
