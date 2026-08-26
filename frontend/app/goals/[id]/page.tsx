'use client';

import { useState, useEffect, use } from 'react';
import { useRouter } from 'next/navigation';
import { apiCall, ApiError } from '@/lib/api';
import { Loader2, PartyPopper, Trash2, PiggyBank } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Select } from '@/components/ui/Input';
import { CurrencyInput } from '@/components/ui/CurrencyInput';
import { FormField } from '@/components/ui/FormField';
import { Alert } from '@/components/ui/Alert';
import { TopBar } from '@/components/ui/TopBar';

interface Contribution {
  id: string;
  amount: number;
  transaction_date: string;
  created_by_name: string;
  account_name: string;
  destination_account_id?: string | null;
  account_id: string;
}

interface GoalDetail {
  id: string;
  name: string;
  target_amount: number;
  linked_account_id: string;
  target_date?: string | null;
  status: 'active' | 'achieved' | 'cancelled';
  current_amount: number;
  percentage: number;
  contributions: Contribution[];
}

interface Account {
  id: string;
  name: string;
  is_active: boolean;
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(value);
}

export default function GoalDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [goal, setGoal] = useState<GoalDetail | null>(null);
  const [accounts, setAccounts] = useState<Account[]>([]);

  const [showSaveModal, setShowSaveModal] = useState(false);
  const [saveSourceAccountId, setSaveSourceAccountId] = useState('');
  const [saveAmount, setSaveAmount] = useState('');

  const fetchData = async () => {
    setLoading(true);
    setError('');
    try {
      const [goalRes, accountsRes] = await Promise.all([
        apiCall<GoalDetail>(`/goals/${id}`),
        apiCall<Account[]>('/accounts'),
      ]);
      setGoal(goalRes);
      setAccounts((accountsRes || []).filter((a) => a.id !== goalRes.linked_account_id));
    } catch (err) {
      if (err instanceof ApiError && err.statusCode === 401) {
        router.push('/login');
      } else {
        setError('Gagal memuat detail goal');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  const openSaveModal = () => {
    setSaveSourceAccountId(accounts[0]?.id || '');
    setSaveAmount('');
    setShowSaveModal(true);
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!saveSourceAccountId || !saveAmount || parseFloat(saveAmount) <= 0 || !goal) {
      setError('Akun sumber dan nominal wajib diisi dengan benar');
      return;
    }

    setError('');
    setSuccess('');
    setUpdating(true);

    try {
      const today = new Date();
      const dateStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;

      await apiCall('/transactions', {
        method: 'POST',
        body: JSON.stringify({
          type: 'transfer',
          account_id: saveSourceAccountId,
          destination_account_id: goal.linked_account_id,
          goal_id: goal.id,
          amount: parseFloat(saveAmount),
          transaction_date: dateStr,
        }),
      });

      setSuccess('Berhasil menabung!');
      setShowSaveModal(false);
      fetchData();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal menabung');
      }
    } finally {
      setUpdating(false);
    }
  };

  const handleDeleteGoal = async () => {
    if (!window.confirm('Hapus goal ini?')) return;

    setError('');
    setUpdating(true);

    try {
      await apiCall(`/goals/${id}`, { method: 'DELETE' });
      router.push('/goals');
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal hapus goal');
      }
      setUpdating(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-base-100">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!goal) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4 bg-base-100">
        <Alert type="error" message="Goal tidak ditemukan" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-base-100 pb-6">
      <TopBar title={goal.name} subtitle="Detail target tabungan" backHref="/goals" />

      <div className="max-w-2xl mx-auto p-4 sm:p-6">
        {error && <Alert type="error" message={error} />}
        {success && <Alert type="success" message={success} />}

        <Card className="mb-6">
          {goal.status === 'achieved' && (
            <div className="flex items-center gap-2 mb-4 px-4 py-3 rounded-2xl bg-primary/10 text-primary">
              <PartyPopper className="w-5 h-5" />
              <span className="font-semibold text-sm">Selamat! Goal ini sudah tercapai 🎉</span>
            </div>
          )}

          <div className="mb-4">
            <p className="text-2xl sm:text-3xl font-bold text-base-content">{formatCurrency(goal.current_amount)}</p>
            <p className="text-sm text-base-content/60">dari target {formatCurrency(goal.target_amount)}</p>
          </div>

          <div className="w-full h-3 rounded-full bg-base-300 overflow-hidden mb-2">
            <div
              className={`h-full rounded-full transition-all ${goal.status === 'achieved' ? 'bg-primary' : 'bg-success'}`}
              style={{ width: `${Math.min(100, goal.percentage)}%` }}
            />
          </div>
          <p className="text-xs text-base-content/60 mb-6">{goal.percentage.toFixed(0)}% tercapai</p>

          {goal.target_date && (
            <p className="text-xs text-base-content/60 mb-4">
              Target: {new Date(goal.target_date).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })}
            </p>
          )}

          <div className="flex gap-3">
            <Button fullWidth onClick={openSaveModal} disabled={updating || accounts.length === 0}>
              <PiggyBank className="w-4 h-4" />
              Nabung
            </Button>
            <Button variant="ghost" onClick={handleDeleteGoal} disabled={updating}>
              <Trash2 className="w-4 h-4" />
            </Button>
          </div>
        </Card>

        <Card>
          <h2 className="text-lg sm:text-xl font-bold text-base-content mb-4">Histori Kontribusi</h2>

          {goal.contributions.length === 0 ? (
            <p className="text-sm text-base-content/60 text-center py-8">Belum ada kontribusi</p>
          ) : (
            <div className="flex flex-col gap-3">
              {goal.contributions.map((c) => {
                const isWithdrawal = c.account_id === goal.linked_account_id;
                return (
                  <div key={c.id} className="flex items-center justify-between px-4 py-3 bg-base-100 border border-base-300 rounded-2xl">
                    <div>
                      <p className="text-sm font-medium text-base-content">
                        {isWithdrawal ? 'Penarikan' : 'Setoran'} oleh {c.created_by_name}
                      </p>
                      <p className="text-xs text-base-content/60">
                        {new Date(c.transaction_date).toLocaleDateString('id-ID')}
                      </p>
                    </div>
                    <p className={`font-semibold ${isWithdrawal ? 'text-error' : 'text-success'}`}>
                      {isWithdrawal ? '-' : '+'}
                      {formatCurrency(c.amount)}
                    </p>
                  </div>
                );
              })}
            </div>
          )}
        </Card>
      </div>

      {showSaveModal && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-end sm:items-center justify-center p-4">
          <Card className="w-full sm:max-w-md rounded-t-3xl sm:rounded-3xl">
            <h2 className="text-xl sm:text-2xl font-bold text-base-content mb-6">Nabung ke {goal.name}</h2>

            <form onSubmit={handleSave} className="flex flex-col gap-4">
              <FormField label="Dari Akun">
                <Select value={saveSourceAccountId} onChange={(e) => setSaveSourceAccountId(e.target.value)} disabled={updating}>
                  <option value="">Pilih akun sumber</option>
                  {accounts.map((acc) => (
                    <option key={acc.id} value={acc.id}>
                      {acc.name}
                    </option>
                  ))}
                </Select>
              </FormField>

              <FormField label="Nominal">
                <CurrencyInput
                  placeholder="500.000"
                  value={saveAmount}
                  onChange={setSaveAmount}
                  disabled={updating}
                  autoFocus
                />
              </FormField>

              <div className="flex gap-3 pt-2">
                <Button type="button" variant="ghost" fullWidth onClick={() => setShowSaveModal(false)} disabled={updating}>
                  Batal
                </Button>
                <Button type="submit" fullWidth disabled={updating}>
                  {updating ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Nabung'}
                </Button>
              </div>
            </form>
          </Card>
        </div>
      )}
    </div>
  );
}
