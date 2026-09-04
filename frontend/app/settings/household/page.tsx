'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { apiCall, ApiError } from '@/lib/api';
import { Loader2, Copy, Trash2 } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { FormField } from '@/components/ui/FormField';
import { Alert } from '@/components/ui/Alert';
import { TopBar } from '@/components/ui/TopBar';
import { AppShell } from '@/components/ui/AppShell';

interface HouseholdDetail {
  id: string;
  name: string;
  role: 'owner' | 'member';
  member_count: number;
}

interface Member {
  user_id: string;
  name: string;
  email: string;
  avatar_url: string | null;
  role: 'owner' | 'member';
  joined_at: string;
}

interface Invitation {
  code: string;
  expires_at: string;
}

export default function HouseholdSettingsPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [household, setHousehold] = useState<HouseholdDetail | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [invitation, setInvitation] = useState<Invitation | null>(null);

  const [editingName, setEditingName] = useState(false);
  const [newName, setNewName] = useState('');
  const [removingMemberId, setRemovingMemberId] = useState<string | null>(null);
  const [copiedCode, setCopiedCode] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [householdRes, membersRes, invitationRes] = await Promise.all([
          apiCall<HouseholdDetail>('/households/me'),
          apiCall<Member[]>('/households/members'),
          apiCall<Invitation>('/households/invitations/active'),
        ]);

        setHousehold(householdRes);
        setMembers(membersRes || []);
        setInvitation(invitationRes);
        setNewName(householdRes.name);
      } catch (err) {
        if (err instanceof ApiError && err.statusCode === 401) {
          router.push('/login');
        } else {
          setError('Gagal memuat data anggota');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [router]);

  const handleUpdateName = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!household || !newName.trim()) {
      setError('Nama anggota tidak boleh kosong');
      return;
    }

    setError('');
    setSuccess('');
    setUpdating(true);

    try {
      await apiCall(`/households/${household.id}`, {
        method: 'PATCH',
        body: JSON.stringify({ name: newName }),
      });

      setSuccess('Nama anggota berhasil diupdate');
      setHousehold({ ...household, name: newName });
      setEditingName(false);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal update nama anggota');
      }
    } finally {
      setUpdating(false);
    }
  };

  const handleGenerateCode = async () => {
    if (!household) return;

    setError('');
    setSuccess('');
    setUpdating(true);

    try {
      const newInv = await apiCall<Invitation>('/households/invitations', {
        method: 'POST',
      });

      setInvitation(newInv);
      setSuccess('Kode undangan baru berhasil dibuat');
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal membuat kode undangan');
      }
    } finally {
      setUpdating(false);
    }
  };

  const handleCopyCode = () => {
    if (invitation) {
      navigator.clipboard.writeText(invitation.code);
      setCopiedCode(true);
      setTimeout(() => setCopiedCode(false), 2000);
    }
  };

  const handleRemoveMember = async (userId: string) => {
    if (!window.confirm('Yakin ingin keluarkan anggota ini?')) {
      return;
    }

    setError('');
    setSuccess('');
    setUpdating(true);
    setRemovingMemberId(userId);

    try {
      await apiCall(`/households/members?targetUserID=${userId}`, {
        method: 'DELETE',
      });

      setMembers(members.filter((m) => m.user_id !== userId));
      setSuccess('Anggota berhasil dikeluarkan');
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal keluarkan anggota');
      }
    } finally {
      setUpdating(false);
      setRemovingMemberId(null);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-base-100">
        <div className="flex flex-col items-center gap-4">
          <Loader2 className="w-8 h-8 animate-spin text-primary" />
          <p className="text-base-content/60">Memuat anggota...</p>
        </div>
      </div>
    );
  }

  if (!household) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4 bg-base-100">
        <Alert type="error" message="Gagal memuat anggota" />
      </div>
    );
  }

  return (
    <AppShell active="/settings/household">
      <TopBar title="Anggota" subtitle="Kelola anggota & undangan" backHref="/settings/profile" />

      <div className="max-w-4xl mx-auto p-4 sm:p-6">
        {error && <Alert type="error" message={error} />}
        {success && <Alert type="success" message={success} />}

        <div className="flex flex-col gap-6 sm:gap-8">
          {/* Household Details */}
          <Card>
            <h2 className="text-xl sm:text-2xl font-bold mb-6 text-base-content">Detail Anggota</h2>

            {editingName ? (
              <form onSubmit={handleUpdateName} className="flex flex-col gap-4">
                <FormField label="Nama Anggota">
                  <Input
                    type="text"
                    required
                    value={newName}
                    onChange={(e) => setNewName(e.target.value)}
                    disabled={updating}
                    autoFocus
                  />
                </FormField>
                <div className="flex gap-3">
                  <Button
                    type="button"
                    variant="ghost"
                    fullWidth
                    onClick={() => {
                      setEditingName(false);
                      setNewName(household.name);
                    }}
                    disabled={updating}
                  >
                    Batal
                  </Button>
                  <Button type="submit" fullWidth disabled={updating}>
                    {updating ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        Menyimpan...
                      </>
                    ) : (
                      'Simpan'
                    )}
                  </Button>
                </div>
              </form>
            ) : (
              <div className="flex flex-col gap-4">
                <div className="p-4 bg-base-100 rounded-2xl border border-base-300">
                  <p className="text-xs sm:text-sm text-base-content/60 mb-1">Nama Anggota</p>
                  <p className="font-medium text-base sm:text-lg text-base-content">{household.name}</p>
                </div>
                <div className="p-4 bg-base-100 rounded-2xl border border-base-300">
                  <p className="text-xs sm:text-sm text-base-content/60 mb-1">Peran Anda</p>
                  <p className="font-medium text-base sm:text-lg text-base-content capitalize">
                    {household.role === 'owner' ? 'Pemilik' : 'Anggota'}
                  </p>
                </div>
                <div className="p-4 bg-base-100 rounded-2xl border border-base-300">
                  <p className="text-xs sm:text-sm text-base-content/60 mb-1">Jumlah Anggota</p>
                  <p className="font-medium text-base sm:text-lg text-base-content">{household.member_count} orang</p>
                </div>

                {household.role === 'owner' && (
                  <Button variant="outline" fullWidth onClick={() => setEditingName(true)} disabled={updating}>
                    Ubah Nama
                  </Button>
                )}
              </div>
            )}
          </Card>

          {/* Invitation Code */}
          {household.role === 'owner' && (
            <Card>
              <h2 className="text-xl sm:text-2xl font-bold mb-6 text-base-content">Kode Undangan</h2>

              {invitation ? (
                <div className="flex flex-col gap-4">
                  <div className="p-4 bg-base-100 rounded-2xl border border-base-300">
                    <p className="text-xs sm:text-sm text-base-content/60 mb-2">Kode Aktif</p>
                    <div className="flex items-center gap-2">
                      <code className="font-mono text-lg sm:text-xl font-bold text-primary tracking-widest">
                        {invitation.code}
                      </code>
                      <button
                        onClick={handleCopyCode}
                        className="inline-flex items-center justify-center w-9 h-9 rounded-full hover:bg-base-200 transition-colors"
                        title="Copy kode"
                      >
                        <Copy className="w-4 h-4" />
                      </button>
                    </div>
                    {copiedCode && <p className="text-xs text-success mt-2">Kode disalin!</p>}
                    <p className="text-xs sm:text-sm text-base-content/60 mt-2">
                      Berlaku hingga: {new Date(invitation.expires_at).toLocaleDateString('id-ID')}
                    </p>
                  </div>

                  <Button fullWidth onClick={handleGenerateCode} disabled={updating}>
                    {updating ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        Membuat...
                      </>
                    ) : (
                      'Generate Kode Baru'
                    )}
                  </Button>
                </div>
              ) : (
                <Button fullWidth onClick={handleGenerateCode} disabled={updating}>
                  {updating ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Membuat...
                    </>
                  ) : (
                    'Buat Kode Undangan'
                  )}
                </Button>
              )}
            </Card>
          )}

          {/* Members */}
          <Card>
            <h2 className="text-xl sm:text-2xl font-bold mb-6 text-base-content">Daftar Anggota</h2>

            <div className="flex flex-col gap-3">
              {members.map((member) => (
                <div key={member.user_id} className="p-4 bg-base-100 border border-base-300 rounded-2xl">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1">
                      <p className="font-semibold text-base-content">{member.name}</p>
                      <p className="text-sm text-base-content/60">{member.email}</p>
                      <div className="flex items-center gap-2 mt-2">
                        <span className="text-xs px-2 py-1 bg-primary/10 text-primary rounded-full capitalize font-medium">
                          {member.role === 'owner' ? 'Pemilik' : 'Anggota'}
                        </span>
                        <span className="text-xs text-base-content/60">
                          Bergabung {new Date(member.joined_at).toLocaleDateString('id-ID')}
                        </span>
                      </div>
                    </div>

                    {household.role === 'owner' && member.role !== 'owner' && (
                      <button
                        onClick={() => handleRemoveMember(member.user_id)}
                        className="inline-flex items-center justify-center w-9 h-9 rounded-full text-error hover:bg-error/10 transition-colors flex-shrink-0"
                        disabled={updating || removingMemberId === member.user_id}
                      >
                        {removingMemberId === member.user_id ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <Trash2 className="w-4 h-4" />
                        )}
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </div>
      </div>

      </AppShell>
  );
}
