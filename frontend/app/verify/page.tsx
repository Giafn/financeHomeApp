'use client';

import { useEffect, useState, Suspense } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { Loader2, CheckCircle2, XCircle, MailX, MailCheck } from 'lucide-react';
import { apiCall, ApiError } from '@/lib/api';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Alert } from '@/components/ui/Alert';

function VerifyContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get('token') || '';

  const [status, setStatus] = useState<'loading' | 'success' | 'error' | 'invalid'>(
    token ? 'loading' : 'invalid'
  );
  const [message, setMessage] = useState('');

  useEffect(() => {
    if (!token) return;

    apiCall('/auth/verify', {
      method: 'POST',
      body: JSON.stringify({ token }),
    })
      .then(() => {
        setStatus('success');
      })
      .catch((err) => {
        setStatus('error');
        if (err instanceof ApiError) {
          setMessage(err.message);
        } else {
          setMessage('Verifikasi gagal. Silakan coba lagi.');
        }
      });
  }, [token]);

  return (
    <div className="min-h-screen flex items-center justify-center p-4 sm:p-6 bg-base-100">
      <div className="w-full max-w-md text-center">
        {status === 'loading' && (
          <div className="flex flex-col items-center gap-4">
            <Loader2 className="w-10 h-10 animate-spin text-primary" />
            <p className="text-base-content/60">Memverifikasi email...</p>
          </div>
        )}

        {status === 'success' && (
          <>
            <div className="flex justify-center mb-6">
              <div className="w-16 h-16 rounded-full bg-success/10 flex items-center justify-center">
                <CheckCircle2 className="w-8 h-8 text-success" />
              </div>
            </div>
            <h1 className="text-3xl font-bold mb-3 text-base-content">Email berhasil diverifikasi!</h1>
            <p className="text-sm sm:text-base text-base-content/60 mb-8">
              Akun kamu sudah aktif. Sekarang kamu bisa masuk ke Family Finance.
            </p>
            <Card className="p-4">
              <Link href="/login">
                <Button fullWidth>Masuk sekarang</Button>
              </Link>
            </Card>
          </>
        )}

        {status === 'invalid' && (
          <>
            <div className="flex justify-center mb-6">
              <div className="w-16 h-16 rounded-full bg-error/10 flex items-center justify-center">
                <XCircle className="w-8 h-8 text-error" />
              </div>
            </div>
            <h1 className="text-3xl font-bold mb-3 text-base-content">Link tidak valid</h1>
            <Card className="p-4 mb-6">
              <Alert type="error" message="Link verifikasi tidak lengkap. Pastikan kamu membuka link dari email yang utuh." />
              <p className="text-xs sm:text-sm text-base-content/60 mt-4 flex items-start justify-center gap-2">
                <MailX className="w-4 h-4 mt-0.5" />
                Jika link sudah kadaluarsa, kamu bisa kirim ulang email verifikasi dari halaman login.
              </p>
            </Card>
            <Link href="/login" className="inline-flex items-center gap-2 text-primary font-medium hover:underline">
              <MailCheck className="w-4 h-4" />
              Ke halaman Login
            </Link>
          </>
        )}

        {status === 'error' && (
          <>
            <div className="flex justify-center mb-6">
              <div className="w-16 h-16 rounded-full bg-error/10 flex items-center justify-center">
                <XCircle className="w-8 h-8 text-error" />
              </div>
            </div>
            <h1 className="text-3xl font-bold mb-3 text-base-content">Verifikasi gagal</h1>
            <Card className="p-4 mb-6">
              <Alert type="error" message={message} />
              <p className="text-xs sm:text-sm text-base-content/60 mt-4 flex items-start justify-center gap-2">
                <MailX className="w-4 h-4 mt-0.5" />
                Jika link sudah kadaluarsa, kamu bisa kirim ulang email verifikasi dari halaman login.
              </p>
            </Card>
            <Link href="/login" className="inline-flex items-center gap-2 text-primary font-medium hover:underline">
              <MailCheck className="w-4 h-4" />
              Ke halaman Login
            </Link>
          </>
        )}
      </div>
    </div>
  );
}

export default function VerifyPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen flex items-center justify-center p-4 sm:p-6 bg-base-100">
          <Loader2 className="w-10 h-10 animate-spin text-primary" />
        </div>
      }
    >
      <VerifyContent />
    </Suspense>
  );
}
