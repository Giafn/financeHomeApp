'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { apiCall, ApiError } from '@/lib/api';
import { Loader2, Plus, Banknote, Smartphone, Wallet, MoreVertical } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { CurrencyInput } from '@/components/ui/CurrencyInput';
import { FormField } from '@/components/ui/FormField';
import { Alert } from '@/components/ui/Alert';
import { BottomNav } from '@/components/ui/BottomNav';

interface Account {
  id: string;
  name: string;
  type: 'bank' | 'ewallet' | 'cash' | 'other';
  initial_balance: number;
  current_balance: number;
  is_active: boolean;
}

export default function AccountsPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [accounts, setAccounts] = useState<Account[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [togglingId, setTogglingId] = useState<string | null>(null);

  const [formData, setFormData] = useState({
    name: '',
    type: 'bank' as 'bank' | 'ewallet' | 'cash' | 'other',
    initial_balance: '',
  });

  useEffect(() => {
    const fetchAccounts = async () => {
      try {
        const token = localStorage.getItem('token');
        if (!token) {
          router.push('/login');
          return;
        }

        const data = await apiCall<Account[]>('/accounts?include_inactive=true');
        setAccounts(data || []);
      } catch (err) {
        if (err instanceof ApiError && err.statusCode === 401) {
          router.push('/login');
        } else {
          setError('Gagal memuat akun');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchAccounts();
  }, [router]);

  const handleCreateAccount = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name.trim() || !formData.initial_balance) {
      setError('Nama dan saldo awal diperlukan');
      return;
    }

    setError('');
    setSuccess('');
    setUpdating(true);

    try {
      const newAccount = await apiCall<Account>('/accounts', {
        method: 'POST',
        body: JSON.stringify({
          name: formData.name,
          type: formData.type,
          initial_balance: parseFloat(formData.initial_balance),
        }),
      });

      setAccounts([...accounts, newAccount]);
      setSuccess('Akun berhasil dibuat');
      setFormData({ name: '', type: 'bank', initial_balance: '' });
      setShowModal(false);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal membuat akun');
      }
    } finally {
      setUpdating(false);
    }
  };

  const handleToggleActive = async (accountId: string, currentActive: boolean) => {
    setError('');
    setSuccess('');
    setUpdating(true);
    setTogglingId(accountId);

    try {
      await apiCall(`/accounts/${accountId}`, {
        method: 'PATCH',
        body: JSON.stringify({ is_active: !currentActive }),
      });

      setAccounts(
        accounts.map((acc) =>
          acc.id === accountId ? { ...acc, is_active: !currentActive } : acc
        )
      );
      setSuccess(currentActive ? 'Akun dinonaktifkan' : 'Akun diaktifkan');
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal update akun');
      }
    } finally {
      setUpdating(false);
      setTogglingId(null);
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'bank':
        return <Banknote className="w-5 h-5" />;
      case 'ewallet':
        return <Smartphone className="w-5 h-5" />;
      case 'cash':
        return <Wallet className="w-5 h-5" />;
      default:
        return <MoreVertical className="w-5 h-5" />;
    }
  };

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(value);
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-base-100">
        <div className="flex flex-col items-center gap-4">
          <Loader2 className="w-8 h-8 animate-spin text-primary" />
          <p className="text-base-content/60">Memuat akun...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-base-100 pb-24">
      {/* Topbar */}
      <div className="border-b border-base-300 bg-base-200 sticky top-0 z-50">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
          <h1 className="text-2xl sm:text-3xl font-bold text-base-content">Akun</h1>
          <Button onClick={() => setShowModal(true)}>
            <Plus className="w-4 h-4" />
            Tambah
          </Button>
        </div>
      </div>

      <div className="max-w-6xl mx-auto p-4 sm:p-6">
        {error && <Alert type="error" message={error} />}
        {success && <Alert type="success" message={success} />}

        {accounts.length === 0 ? (
          <div className="text-center py-12">
            <Wallet className="w-12 h-12 mx-auto mb-4 text-base-content/40" />
            <p className="text-base-content/60 mb-4">Belum ada akun</p>
            <Button onClick={() => setShowModal(true)}>
              <Plus className="w-4 h-4" />
              Buat Akun Pertama
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
            {accounts.map((account) => (
              <Card key={account.id} className={!account.is_active ? 'opacity-50' : ''}>
                <div className="flex items-start justify-between gap-2 mb-4">
                  <div className="flex items-center gap-3 flex-1">
                    <div className="p-2 bg-base-100 rounded-full text-primary">
                      {getTypeIcon(account.type)}
                    </div>
                    <div className="flex-1">
                      <h3 className="font-semibold text-base-content">{account.name}</h3>
                      <p className="text-xs text-base-content/60 capitalize">{account.type}</p>
                    </div>
                  </div>
                  {!account.is_active && (
                    <span className="text-xs px-2 py-1 bg-base-300 text-base-content/60 rounded-full">
                      Nonaktif
                    </span>
                  )}
                </div>

                <div className="mb-6">
                  <p className="text-xs text-base-content/60 mb-1">Saldo</p>
                  <p
                    className={`text-2xl font-bold ${
                      account.current_balance < 0 ? 'text-error' : 'text-base-content'
                    }`}
                  >
                    {formatCurrency(account.current_balance)}
                  </p>
                </div>

                <div className="pt-4 border-t border-base-300">
                  <Button
                    variant="ghost"
                    fullWidth
                    onClick={() => handleToggleActive(account.id, account.is_active)}
                    disabled={updating || togglingId === account.id}
                  >
                    {togglingId === account.id ? (
                      <Loader2 className="w-4 h-4 animate-spin" />
                    ) : account.is_active ? (
                      'Nonaktifkan'
                    ) : (
                      'Aktifkan'
                    )}
                  </Button>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* Create Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-end sm:items-center justify-center p-4">
          <Card className="w-full sm:max-w-md rounded-t-3xl sm:rounded-3xl">
            <h2 className="text-xl sm:text-2xl font-bold text-base-content mb-6">Tambah Akun</h2>

            <form onSubmit={handleCreateAccount} className="flex flex-col gap-4">
              <FormField label="Nama Akun">
                <Input
                  type="text"
                  placeholder="BCA Suami"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  disabled={updating}
                />
              </FormField>

              <FormField label="Tipe Akun">
                <Select
                  value={formData.type}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      type: e.target.value as 'bank' | 'ewallet' | 'cash' | 'other',
                    })
                  }
                  disabled={updating}
                >
                  <option value="bank">Bank</option>
                  <option value="ewallet">E-Wallet</option>
                  <option value="cash">Tunai</option>
                  <option value="other">Lainnya</option>
                </Select>
              </FormField>

              <FormField label="Saldo Awal">
                <CurrencyInput
                  placeholder="5.000.000"
                  value={formData.initial_balance}
                  onChange={(v) => setFormData({ ...formData, initial_balance: v })}
                  disabled={updating}
                />
              </FormField>

              <div className="flex gap-3 pt-4">
                <Button
                  type="button"
                  variant="ghost"
                  fullWidth
                  onClick={() => {
                    setShowModal(false);
                    setFormData({ name: '', type: 'bank', initial_balance: '' });
                    setError('');
                  }}
                  disabled={updating}
                >
                  Batal
                </Button>
                <Button type="submit" fullWidth disabled={updating}>
                  {updating ? (
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
          </Card>
        </div>
      )}

      <BottomNav active="/accounts" />
    </div>
  );
}
