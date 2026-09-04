'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Loader2, ChevronRight } from 'lucide-react';
import { apiCall, ApiError } from '@/lib/api';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { FormField } from '@/components/ui/FormField';
import { Alert } from '@/components/ui/Alert';

export default function OnboardingPage() {
  const router = useRouter();
  const [step, setStep] = useState<'choice' | 'create' | 'join'>('choice');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [householdName, setHouseholdName] = useState('');
  const [invitationCode, setInvitationCode] = useState('');

  const handleCreateHousehold = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await apiCall('/households', {
        method: 'POST',
        body: JSON.stringify({ name: householdName }),
      });

      router.push('/dashboard');
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.statusCode === 409) {
          setError('Anda sudah tergabung dalam anggota lain');
        } else {
          setError(err.message);
        }
      } else {
        setError('Gagal membuat anggota');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleJoinHousehold = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await apiCall('/households/join', {
        method: 'POST',
        body: JSON.stringify({ code: invitationCode }),
      });

      router.push('/dashboard');
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.statusCode === 400) {
          setError('Kode undangan tidak valid atau sudah kadaluarsa');
        } else if (err.statusCode === 409) {
          setError('Anda sudah tergabung dalam anggota lain');
        } else {
          setError(err.message);
        }
      } else {
        setError('Gagal bergabung ke anggota');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 sm:p-6 bg-base-100">
      <div className="w-full max-w-md">
        <div className="mb-6 sm:mb-8">
          <h1 className="text-3xl sm:text-4xl font-bold mb-2 text-base-content">Siapkan anggota</h1>
          <p className="text-sm sm:text-base text-base-content/60">Buat atau bergabung dengan anggota untuk mulai</p>
        </div>

        <Card>
          {error && <Alert type="error" message={error} />}

          {step === 'choice' && (
            <div className="flex flex-col gap-4">
              <button
                onClick={() => setStep('create')}
                disabled={loading}
                className="bg-base-100 border border-base-300 hover:border-primary transition-colors rounded-2xl p-6 w-full text-left active:scale-[0.98]"
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1">
                    <h3 className="font-semibold text-base-content mb-1 text-sm sm:text-base">Buat Anggota</h3>
                    <p className="text-xs sm:text-sm text-base-content/60">Anda akan menjadi pemilik</p>
                  </div>
                  <ChevronRight className="w-5 h-5 text-primary flex-shrink-0 mt-1" />
                </div>
              </button>

              <button
                onClick={() => setStep('join')}
                disabled={loading}
                className="bg-base-100 border border-base-300 hover:border-primary transition-colors rounded-2xl p-6 w-full text-left active:scale-[0.98]"
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1">
                    <h3 className="font-semibold text-base-content mb-1 text-sm sm:text-base">Bergabung dengan Anggota</h3>
                    <p className="text-xs sm:text-sm text-base-content/60">Gunakan kode undangan</p>
                  </div>
                  <ChevronRight className="w-5 h-5 text-primary flex-shrink-0 mt-1" />
                </div>
              </button>
            </div>
          )}

          {step === 'create' && (
            <form onSubmit={handleCreateHousehold} className="flex flex-col gap-5 sm:gap-6">
              <FormField label="Nama Anggota">
                <Input
                  type="text"
                  placeholder="Kost Bareng / Anggota"
                  required
                  minLength={2}
                  value={householdName}
                  onChange={(e) => setHouseholdName(e.target.value)}
                  disabled={loading}
                  autoFocus
                />
              </FormField>

              <div className="flex gap-3 pt-2">
                <Button
                  type="button"
                  variant="ghost"
                  fullWidth
                  onClick={() => setStep('choice')}
                  disabled={loading}
                >
                  Kembali
                </Button>
                <Button type="submit" fullWidth disabled={loading}>
                  {loading ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Membuat...
                    </>
                  ) : (
                    'Buat'
                  )}
                </Button>
              </div>
            </form>
          )}

          {step === 'join' && (
            <form onSubmit={handleJoinHousehold} className="flex flex-col gap-5 sm:gap-6">
              <FormField label="Kode Undangan" hint="Minta kode ini dari pemilik anggota">
                <Input
                  type="text"
                  placeholder="AB12CD34"
                  required
                  className="uppercase tracking-widest font-mono"
                  value={invitationCode}
                  onChange={(e) => setInvitationCode(e.target.value.toUpperCase())}
                  disabled={loading}
                  autoFocus
                />
              </FormField>

              <div className="flex gap-3 pt-2">
                <Button
                  type="button"
                  variant="ghost"
                  fullWidth
                  onClick={() => setStep('choice')}
                  disabled={loading}
                >
                  Kembali
                </Button>
                <Button type="submit" fullWidth disabled={loading}>
                  {loading ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Bergabung...
                    </>
                  ) : (
                    'Bergabung'
                  )}
                </Button>
              </div>
            </form>
          )}
        </Card>
      </div>
    </div>
  );
}
