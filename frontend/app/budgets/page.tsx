'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { apiCall, ApiError } from '@/lib/api';
import { Loader2, ChevronLeft, ChevronRight, Plus, Pencil, Trash2 } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { CurrencyInput } from '@/components/ui/CurrencyInput';
import { FormField } from '@/components/ui/FormField';
import { Alert } from '@/components/ui/Alert';
import { AppShell } from '@/components/ui/AppShell';

interface Budget {
  id: string;
  category_id: string;
  category_name: string;
  period: string;
  amount: number;
  spent: number;
  percentage: number;
}

interface Category {
  id: string;
  name: string;
  type: 'income' | 'expense';
  is_archived: boolean;
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(value);
}

function currentPeriod() {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
}

function shiftPeriod(period: string, delta: number) {
  const [y, m] = period.split('-').map(Number);
  const d = new Date(y, m - 1 + delta, 1);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
}

function periodLabel(period: string) {
  const [y, m] = period.split('-').map(Number);
  const d = new Date(y, m - 1, 1);
  return d.toLocaleDateString('id-ID', { month: 'long', year: 'numeric' });
}

function barColor(pct: number) {
  if (pct > 100) return 'bg-error';
  if (pct >= 80) return 'bg-warning';
  return 'bg-success';
}

export default function BudgetsPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [period, setPeriod] = useState(currentPeriod());
  const [budgets, setBudgets] = useState<Budget[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);

  const [showModal, setShowModal] = useState(false);
  const [editingBudget, setEditingBudget] = useState<Budget | null>(null);
  const [formCategoryId, setFormCategoryId] = useState('');
  const [formAmount, setFormAmount] = useState('');

  const isPastPeriod = period < currentPeriod();

  const fetchBudgets = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const [budgetsRes, categoriesRes] = await Promise.all([
        apiCall<Budget[]>(`/budgets?period=${period}`),
        apiCall<Category[]>('/categories?type=expense&include_archived=false'),
      ]);
      setBudgets(budgetsRes || []);
      setCategories(categoriesRes || []);
    } catch (err) {
      if (err instanceof ApiError && err.statusCode === 401) {
        router.push('/login');
      } else {
        setError('Gagal memuat budget');
      }
    } finally {
      setLoading(false);
    }
  }, [period, router]);

  useEffect(() => {
    fetchBudgets();
  }, [fetchBudgets]);

  const budgetedCategoryIds = new Set(budgets.map((b) => b.category_id));
  const availableCategories = categories.filter((c) => !budgetedCategoryIds.has(c.id));

  const openCreateModal = () => {
    setEditingBudget(null);
    setFormCategoryId(availableCategories[0]?.id || '');
    setFormAmount('');
    setShowModal(true);
  };

  const openEditModal = (budget: Budget) => {
    setEditingBudget(budget);
    setFormCategoryId(budget.category_id);
    setFormAmount(String(budget.amount));
    setShowModal(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formAmount || parseFloat(formAmount) <= 0) {
      setError('Nominal harus lebih dari 0');
      return;
    }

    setError('');
    setSuccess('');
    setUpdating(true);

    try {
      if (editingBudget) {
        await apiCall(`/budgets/${editingBudget.id}`, {
          method: 'PATCH',
          body: JSON.stringify({ amount: parseFloat(formAmount) }),
        });
        setSuccess('Budget berhasil diupdate');
      } else {
        if (!formCategoryId) {
          setError('Kategori wajib dipilih');
          setUpdating(false);
          return;
        }
        await apiCall('/budgets', {
          method: 'POST',
          body: JSON.stringify({ category_id: formCategoryId, period, amount: parseFloat(formAmount) }),
        });
        setSuccess('Budget berhasil dibuat');
      }
      setShowModal(false);
      fetchBudgets();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal menyimpan budget');
      }
    } finally {
      setUpdating(false);
    }
  };

  const handleDelete = async (budget: Budget) => {
    if (!window.confirm(`Hapus budget "${budget.category_name}"?`)) return;

    setError('');
    setSuccess('');
    setUpdating(true);

    try {
      await apiCall(`/budgets/${budget.id}`, { method: 'DELETE' });
      setSuccess('Budget berhasil dihapus');
      fetchBudgets();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal hapus budget');
      }
    } finally {
      setUpdating(false);
    }
  };

  return (
    <AppShell active="/budgets">
      <div className="border-b border-base-300 bg-base-200 sticky top-0 z-50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
          <h1 className="text-2xl sm:text-3xl font-bold text-base-content">Budget</h1>
          {!isPastPeriod && (
            <Button onClick={openCreateModal} disabled={availableCategories.length === 0}>
              <Plus className="w-4 h-4" />
              Tambah
            </Button>
          )}
        </div>
      </div>

      <div className="max-w-4xl mx-auto p-4 sm:p-6">
        {error && <Alert type="error" message={error} />}
        {success && <Alert type="success" message={success} />}

        {/* Month navigator */}
        <div className="flex items-center justify-between mb-6 bg-base-200 rounded-2xl p-2">
          <button
            onClick={() => setPeriod((p) => shiftPeriod(p, -1))}
            className="inline-flex items-center justify-center w-10 h-10 rounded-full hover:bg-base-300 transition-colors"
          >
            <ChevronLeft className="w-5 h-5" />
          </button>
          <p className="font-semibold text-base-content capitalize">{periodLabel(period)}</p>
          <button
            onClick={() => setPeriod((p) => shiftPeriod(p, 1))}
            className="inline-flex items-center justify-center w-10 h-10 rounded-full hover:bg-base-300 transition-colors"
          >
            <ChevronRight className="w-5 h-5" />
          </button>
        </div>

        {isPastPeriod && (
          <p className="text-xs text-base-content/60 mb-4 text-center">
            Bulan lampau bersifat read-only
          </p>
        )}

        {loading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="w-8 h-8 animate-spin text-primary" />
          </div>
        ) : budgets.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-base-content/60 mb-4">Belum ada budget untuk bulan ini</p>
            {!isPastPeriod && (
              <Button onClick={openCreateModal} disabled={availableCategories.length === 0}>
                <Plus className="w-4 h-4" />
                Buat Budget Pertama
              </Button>
            )}
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            {budgets.map((b) => (
              <Card key={b.id} className="!p-5">
                <div className="flex items-start justify-between gap-3 mb-3">
                  <div>
                    <p className="font-semibold text-base-content">{b.category_name}</p>
                    <p className="text-xs text-base-content/60">
                      {formatCurrency(b.spent)} dari {formatCurrency(b.amount)}
                    </p>
                  </div>
                  <div className="flex items-center gap-1">
                    <p
                      className={`text-sm font-bold ${
                        b.percentage > 100 ? 'text-error' : b.percentage >= 80 ? 'text-warning' : 'text-success'
                      }`}
                    >
                      {b.percentage.toFixed(0)}%
                    </p>
                    {!isPastPeriod && (
                      <>
                        <button
                          onClick={() => openEditModal(b)}
                          className="inline-flex items-center justify-center w-8 h-8 rounded-full hover:bg-base-300 transition-colors ml-2"
                        >
                          <Pencil className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => handleDelete(b)}
                          disabled={updating}
                          className="inline-flex items-center justify-center w-8 h-8 rounded-full text-error hover:bg-error/10 transition-colors"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </>
                    )}
                  </div>
                </div>

                <div className="w-full h-2.5 rounded-full bg-base-300 overflow-hidden">
                  <div
                    className={`h-full rounded-full transition-all ${barColor(b.percentage)}`}
                    style={{ width: `${Math.min(100, b.percentage)}%` }}
                  />
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* Create/Edit Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-end sm:items-center justify-center p-4">
          <Card className="w-full sm:max-w-md rounded-t-3xl sm:rounded-3xl">
            <h2 className="text-xl sm:text-2xl font-bold text-base-content mb-6">
              {editingBudget ? 'Edit Budget' : 'Tambah Budget'}
            </h2>

            <form onSubmit={handleSubmit} className="flex flex-col gap-4">
              <FormField label="Kategori">
                {editingBudget ? (
                  <Input type="text" value={editingBudget.category_name} disabled />
                ) : (
                  <Select value={formCategoryId} onChange={(e) => setFormCategoryId(e.target.value)} disabled={updating}>
                    <option value="">Pilih kategori</option>
                    {availableCategories.map((cat) => (
                      <option key={cat.id} value={cat.id}>
                        {cat.name}
                      </option>
                    ))}
                  </Select>
                )}
              </FormField>

              <FormField label="Nominal Budget">
                <CurrencyInput
                  placeholder="1.000.000"
                  value={formAmount}
                  onChange={setFormAmount}
                  disabled={updating}
                  autoFocus
                />
              </FormField>

              <div className="flex gap-3 pt-2">
                <Button type="button" variant="ghost" fullWidth onClick={() => setShowModal(false)} disabled={updating}>
                  Batal
                </Button>
                <Button type="submit" fullWidth disabled={updating}>
                  {updating ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Simpan'}
                </Button>
              </div>
            </form>
          </Card>
        </div>
      )}

      </AppShell>
  );
}
