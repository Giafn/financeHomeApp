'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiCall, ApiError } from '@/lib/api';
import { Loader2, Plus, PartyPopper, Target } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { CurrencyInput } from '@/components/ui/CurrencyInput';
import { FormField } from '@/components/ui/FormField';
import { Alert } from '@/components/ui/Alert';
import { AppShell } from '@/components/ui/AppShell';

interface Goal {
  id: string;
  name: string;
  icon?: string | null;
  target_amount: number;
  linked_account_id: string;
  target_date?: string | null;
  status: 'active' | 'achieved' | 'cancelled';
  current_amount: number;
  percentage: number;
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

export default function GoalsPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [goals, setGoals] = useState<Goal[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);

  const [showModal, setShowModal] = useState(false);
  const [formName, setFormName] = useState('');
  const [formTargetAmount, setFormTargetAmount] = useState('');
  const [formLinkedAccountId, setFormLinkedAccountId] = useState('');
  const [formTargetDate, setFormTargetDate] = useState('');

  const fetchData = async () => {
    setLoading(true);
    setError('');
    try {
      const [goalsRes, accountsRes] = await Promise.all([
        apiCall<Goal[]>('/goals'),
        apiCall<Account[]>('/accounts'),
      ]);
      setGoals(goalsRes || []);
      setAccounts(accountsRes || []);
    } catch (err) {
      if (err instanceof ApiError && err.statusCode === 401) {
        router.push('/login');
      } else {
        setError('Gagal memuat goal');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const openCreateModal = () => {
    setFormName('');
    setFormTargetAmount('');
    setFormLinkedAccountId(accounts[0]?.id || '');
    setFormTargetDate('');
    setShowModal(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formName.trim() || !formTargetAmount || parseFloat(formTargetAmount) <= 0 || !formLinkedAccountId) {
      setError('Semua field wajib diisi dengan benar');
      return;
    }

    setError('');
    setSuccess('');
    setUpdating(true);

    try {
      await apiCall('/goals', {
        method: 'POST',
        body: JSON.stringify({
          name: formName,
          target_amount: parseFloat(formTargetAmount),
          linked_account_id: formLinkedAccountId,
          target_date: formTargetDate || null,
        }),
      });
      setSuccess('Goal berhasil dibuat');
      setShowModal(false);
      fetchData();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal membuat goal');
      }
    } finally {
      setUpdating(false);
    }
  };

  const activeGoals = goals.filter((g) => g.status !== 'cancelled');

  return (
    <AppShell active="/goals">
      <div className="border-b border-base-300 bg-base-200 sticky top-0 z-50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
          <h1 className="text-2xl sm:text-3xl font-bold text-base-content">Goals</h1>
          <Button onClick={openCreateModal} disabled={accounts.length === 0}>
            <Plus className="w-4 h-4" />
            Goal Baru
          </Button>
        </div>
      </div>

      <div className="max-w-4xl mx-auto p-4 sm:p-6">
        {error && <Alert type="error" message={error} />}
        {success && <Alert type="success" message={success} />}

        {loading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="w-8 h-8 animate-spin text-primary" />
          </div>
        ) : activeGoals.length === 0 ? (
          <div className="text-center py-16">
            <Target className="w-12 h-12 mx-auto mb-4 text-base-content/40" />
            <p className="text-base-content/60 mb-4">Belum ada target tabungan</p>
            <Button onClick={openCreateModal} disabled={accounts.length === 0}>
              <Plus className="w-4 h-4" />
              Buat Goal Pertama
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6">
            {activeGoals.map((g) => (
              <Link key={g.id} href={`/goals/${g.id}`}>
                <Card className="!p-5 hover:border-primary transition-colors cursor-pointer h-full">
                  <div className="flex items-start justify-between gap-2 mb-3">
                    <div>
                      <p className="font-semibold text-base-content">{g.name}</p>
                      <p className="text-xs text-base-content/60">
                        {formatCurrency(g.current_amount)} dari {formatCurrency(g.target_amount)}
                      </p>
                    </div>
                    {g.status === 'achieved' && (
                      <span className="inline-flex items-center gap-1 text-xs px-2 py-1 bg-primary/20 text-primary rounded-full font-medium flex-shrink-0">
                        <PartyPopper className="w-3 h-3" />
                        Tercapai!
                      </span>
                    )}
                  </div>

                  <div className="w-full h-2.5 rounded-full bg-base-300 overflow-hidden mb-2">
                    <div
                      className={`h-full rounded-full transition-all ${
                        g.status === 'achieved' ? 'bg-primary' : 'bg-success'
                      }`}
                      style={{ width: `${Math.min(100, g.percentage)}%` }}
                    />
                  </div>
                  <p className="text-xs text-base-content/60">{g.percentage.toFixed(0)}% tercapai</p>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-end sm:items-center justify-center p-4">
          <Card className="w-full sm:max-w-md rounded-t-3xl sm:rounded-3xl">
            <h2 className="text-xl sm:text-2xl font-bold text-base-content mb-6">Goal Baru</h2>

            <form onSubmit={handleSubmit} className="flex flex-col gap-4">
              <FormField label="Nama Goal">
                <Input
                  type="text"
                  placeholder="Contoh: Renovasi Rumah"
                  required
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  disabled={updating}
                  autoFocus
                />
              </FormField>

              <FormField label="Target Nominal">
                <CurrencyInput
                  placeholder="50.000.000"
                  value={formTargetAmount}
                  onChange={setFormTargetAmount}
                  disabled={updating}
                />
              </FormField>

              <FormField label="Akun Tabungan Tujuan">
                <Select value={formLinkedAccountId} onChange={(e) => setFormLinkedAccountId(e.target.value)} disabled={updating}>
                  <option value="">Pilih akun</option>
                  {accounts.map((acc) => (
                    <option key={acc.id} value={acc.id}>
                      {acc.name}
                    </option>
                  ))}
                </Select>
              </FormField>

              <FormField label="Target Tanggal" hint="Opsional">
                <Input
                  type="date"
                  value={formTargetDate}
                  onChange={(e) => setFormTargetDate(e.target.value)}
                  disabled={updating}
                />
              </FormField>

              <div className="flex gap-3 pt-2">
                <Button type="button" variant="ghost" fullWidth onClick={() => setShowModal(false)} disabled={updating}>
                  Batal
                </Button>
                <Button type="submit" fullWidth disabled={updating}>
                  {updating ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Buat'}
                </Button>
              </div>
            </form>
          </Card>
        </div>
      )}

      </AppShell>
  );
}
