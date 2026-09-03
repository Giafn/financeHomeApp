'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { apiCall, ApiError } from '@/lib/api';
import { Loader2, Plus, Receipt, CheckCircle2, Clock, AlertTriangle, X, Pencil, Trash2, Square } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { CurrencyInput } from '@/components/ui/CurrencyInput';
import { FormField } from '@/components/ui/FormField';
import { Alert } from '@/components/ui/Alert';
import { AppShell } from '@/components/ui/AppShell';

interface BillPeriod {
  id: string;
  bill_id: string;
  period: string;
  due_date: string;
  status: 'upcoming' | 'paid' | 'overdue';
  transaction_id?: string | null;
  paid_at?: string | null;
}

interface Bill {
  id: string;
  name: string;
  category_id: string;
  amount: number;
  due_day: number;
  start_period: string;
  end_period?: string | null;
  reminder_days_before: number;
  is_active: boolean;
  next_period?: BillPeriod | null;
}

interface Category {
  id: string;
  name: string;
}

interface Account {
  id: string;
  name: string;
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(value);
}

function statusBadge(status: string) {
  if (status === 'paid') {
    return (
      <span className="inline-flex items-center gap-1 text-xs px-2 py-1 bg-success/20 text-success rounded-full font-medium">
        <CheckCircle2 className="w-3 h-3" />
        Lunas
      </span>
    );
  }
  if (status === 'overdue') {
    return (
      <span className="inline-flex items-center gap-1 text-xs px-2 py-1 bg-error/20 text-error rounded-full font-medium">
        <AlertTriangle className="w-3 h-3" />
        Terlambat
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1 text-xs px-2 py-1 bg-primary/20 text-primary rounded-full font-medium">
      <Clock className="w-3 h-3" />
      Akan Datang
    </span>
  );
}

export default function BillsPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [bills, setBills] = useState<Bill[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);

  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showPayModal, setShowPayModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [editingBill, setEditingBill] = useState<Bill | null>(null);
  const [payingPeriod, setPayingPeriod] = useState<{ bill: Bill; period: BillPeriod } | null>(null);

  const [formName, setFormName] = useState('');
  const [formCategoryId, setFormCategoryId] = useState('');
  const [formAmount, setFormAmount] = useState('');
  const [formDueDay, setFormDueDay] = useState('');
  const [formStartPeriod, setFormStartPeriod] = useState('');
  const [formEndPeriod, setFormEndPeriod] = useState('');

  const [editName, setEditName] = useState('');
  const [editCategoryId, setEditCategoryId] = useState('');
  const [editAmount, setEditAmount] = useState('');
  const [editDueDay, setEditDueDay] = useState('');

  const [payAccountId, setPayAccountId] = useState('');
  const [payAmount, setPayAmount] = useState('');
  const [payDate, setPayDate] = useState('');

  const fetchData = async () => {
    setLoading(true);
    setError('');
    try {
      const [billsRes, categoriesRes, accountsRes] = await Promise.all([
        apiCall<Bill[]>('/bills?is_active=true'),
        apiCall<Category[]>('/categories?type=expense&include_archived=false'),
        apiCall<Account[]>('/accounts'),
      ]);
      setBills(billsRes || []);
      setCategories(categoriesRes || []);
      setAccounts(accountsRes || []);
    } catch (err) {
      if (err instanceof ApiError && err.statusCode === 401) {
        router.push('/login');
      } else {
        setError('Gagal memuat tagihan');
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
    setFormCategoryId(categories[0]?.id || '');
    setFormAmount('');
    setFormDueDay('');
    setFormStartPeriod('');
    setFormEndPeriod('');
    setShowCreateModal(true);
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formName.trim() || !formCategoryId || !formAmount || !formDueDay || !formStartPeriod) {
      setError('Semua field wajib diisi (kecuali periode akhir)');
      return;
    }

    setError('');
    setSuccess('');
    setUpdating(true);

    try {
      await apiCall('/bills', {
        method: 'POST',
        body: JSON.stringify({
          name: formName,
          category_id: formCategoryId,
          amount: parseFloat(formAmount),
          due_day: parseInt(formDueDay, 10),
          start_period: formStartPeriod,
          end_period: formEndPeriod || null,
        }),
      });
      setSuccess('Tagihan berhasil dibuat');
      setShowCreateModal(false);
      fetchData();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal membuat tagihan');
      }
    } finally {
      setUpdating(false);
    }
  };

  const openPayModal = (bill: Bill, period: BillPeriod) => {
    setPayingPeriod({ bill, period });
    setPayAccountId(accounts[0]?.id || '');
    setPayAmount(String(bill.amount));
    const today = new Date();
    setPayDate(`${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`);
    setShowPayModal(true);
  };

  const handlePay = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!payingPeriod || !payAccountId || !payAmount || parseFloat(payAmount) <= 0) {
      setError('Akun dan nominal wajib diisi dengan benar');
      return;
    }

    setError('');
    setSuccess('');
    setUpdating(true);

    try {
      await apiCall(`/bill-periods/${payingPeriod.period.id}/pay`, {
        method: 'POST',
        body: JSON.stringify({
          account_id: payAccountId,
          amount: parseFloat(payAmount),
          transaction_date: payDate,
        }),
      });
      setSuccess('Tagihan berhasil ditandai dibayar');
      setShowPayModal(false);
      fetchData();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal menandai dibayar');
      }
    } finally {
      setUpdating(false);
    }
  };

  const openEditModal = (bill: Bill) => {
    setEditingBill(bill);
    setEditName(bill.name);
    setEditCategoryId(bill.category_id);
    setEditAmount(String(bill.amount));
    setEditDueDay(String(bill.due_day));
    setShowEditModal(true);
  };

  const handleEdit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingBill || !editName.trim() || !editCategoryId || !editAmount || !editDueDay) {
      setError('Semua field wajib diisi');
      return;
    }

    setError('');
    setSuccess('');
    setUpdating(true);

    try {
      await apiCall(`/bills/${editingBill.id}`, {
        method: 'PATCH',
        body: JSON.stringify({
          name: editName,
          category_id: editCategoryId,
          amount: parseFloat(editAmount),
          due_day: parseInt(editDueDay, 10),
        }),
      });
      setSuccess('Tagihan berhasil diperbarui');
      setShowEditModal(false);
      setEditingBill(null);
      fetchData();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal memperbarui tagihan');
      }
    } finally {
      setUpdating(false);
    }
  };

  const handleDelete = async (bill: Bill) => {
    if (!window.confirm(`Hapus tagihan "${bill.name}"? Semua periodenya akan ikut terhapus.`)) {
      return;
    }
    setError('');
    setSuccess('');
    setUpdating(true);
    try {
      await apiCall(`/bills/${bill.id}`, { method: 'DELETE' });
      setSuccess('Tagihan berhasil dihapus');
      fetchData();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal menghapus tagihan');
      }
    } finally {
      setUpdating(false);
    }
  };

  const handleStop = async (bill: Bill) => {
    if (!window.confirm(`Hentikan tagihan "${bill.name}"? Periode masa depan yang belum dibayar akan dihapus.`)) {
      return;
    }
    setError('');
    setSuccess('');
    setUpdating(true);
    try {
      await apiCall(`/bills/${bill.id}/stop`, { method: 'POST' });
      setSuccess('Tagihan berhasil dihentikan');
      fetchData();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal menghentikan tagihan');
      }
    } finally {
      setUpdating(false);
    }
  };

  return (
    <AppShell active="/bills">
      <div className="border-b border-base-300 bg-base-200 sticky top-0 z-50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
          <h1 className="text-2xl sm:text-3xl font-bold text-base-content">Tagihan</h1>
          <Button onClick={openCreateModal} disabled={categories.length === 0}>
            <Plus className="w-4 h-4" />
            Tagihan Baru
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
        ) : bills.length === 0 ? (
          <div className="text-center py-16">
            <Receipt className="w-12 h-12 mx-auto mb-4 text-base-content/40" />
            <p className="text-base-content/60 mb-4">Belum ada tagihan berulang</p>
            <Button onClick={openCreateModal} disabled={categories.length === 0}>
              <Plus className="w-4 h-4" />
              Buat Tagihan Pertama
            </Button>
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            {bills.map((bill) => (
              <Card key={bill.id} className="!p-5">
                <div className="flex items-start justify-between gap-3 mb-3">
                  <div>
                    <p className="font-semibold text-base-content">{bill.name}</p>
                    <p className="text-xs text-base-content/60">
                      {formatCurrency(bill.amount)} · tiap tanggal {bill.due_day}
                    </p>
                  </div>
                  {bill.next_period && statusBadge(bill.next_period.status)}
                </div>

                <div className="flex items-center gap-2 pt-3 border-t border-base-300">
                  <Button size="sm" variant="outline" onClick={() => openEditModal(bill)} disabled={updating}>
                    <Pencil className="w-3.5 h-3.5" />
                    Edit
                  </Button>
                  <Button size="sm" variant="outline" onClick={() => handleStop(bill)} disabled={updating}>
                    <Square className="w-3.5 h-3.5" />
                    Stop
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => handleDelete(bill)} disabled={updating} className="!text-error hover:!bg-error/10 ml-auto">
                    <Trash2 className="w-3.5 h-3.5" />
                    Hapus
                  </Button>
                </div>

                {bill.next_period ? (
                  <div className="flex items-center justify-between pt-3 border-t border-base-300 mt-3">
                    <div>
                      <p className="text-xs text-base-content/60">
                        Periode {bill.next_period.period} · jatuh tempo{' '}
                        {new Date(bill.next_period.due_date).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })}
                      </p>
                    </div>
                    {bill.next_period.status !== 'paid' && (
                      <Button size="sm" onClick={() => openPayModal(bill, bill.next_period!)} disabled={updating || accounts.length === 0}>
                        Tandai Dibayar
                      </Button>
                    )}
                  </div>
                ) : (
                  <p className="text-xs text-base-content/60 pt-3 border-t border-base-300 mt-3">Semua periode sudah dibayar</p>
                )}
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-end sm:items-center justify-center p-4">
          <Card className="w-full sm:max-w-md rounded-t-3xl sm:rounded-3xl max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl sm:text-2xl font-bold text-base-content">Tagihan Baru</h2>
              <button onClick={() => setShowCreateModal(false)} className="text-base-content/60 hover:text-base-content">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleCreate} className="flex flex-col gap-4">
              <FormField label="Nama Tagihan">
                <Input
                  type="text"
                  placeholder="Contoh: Cicilan Renovasi"
                  required
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  disabled={updating}
                  autoFocus
                />
              </FormField>

              <FormField label="Kategori">
                <Select value={formCategoryId} onChange={(e) => setFormCategoryId(e.target.value)} disabled={updating}>
                  <option value="">Pilih kategori</option>
                  {categories.map((cat) => (
                    <option key={cat.id} value={cat.id}>
                      {cat.name}
                    </option>
                  ))}
                </Select>
              </FormField>

              <FormField label="Nominal per Periode">
                <CurrencyInput
                  placeholder="500.000"
                  value={formAmount}
                  onChange={setFormAmount}
                  disabled={updating}
                />
              </FormField>

              <FormField label="Tanggal Jatuh Tempo" hint="1-31, otomatis disesuaikan kalau bulan lebih pendek">
                <Input
                  type="number"
                  placeholder="10"
                  required
                  min="1"
                  max="31"
                  value={formDueDay}
                  onChange={(e) => setFormDueDay(e.target.value)}
                  disabled={updating}
                />
              </FormField>

              <FormField label="Mulai Periode" hint="Format YYYY-MM">
                <Input
                  type="month"
                  required
                  value={formStartPeriod}
                  onChange={(e) => setFormStartPeriod(e.target.value)}
                  disabled={updating}
                />
              </FormField>

              <FormField label="Sampai Periode" hint="Opsional — kosongkan untuk cicilan tanpa batas akhir jelas">
                <Input
                  type="month"
                  value={formEndPeriod}
                  onChange={(e) => setFormEndPeriod(e.target.value)}
                  disabled={updating}
                />
              </FormField>

              <div className="flex gap-3 pt-2">
                <Button type="button" variant="ghost" fullWidth onClick={() => setShowCreateModal(false)} disabled={updating}>
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

      {/* Pay Modal */}
      {showPayModal && payingPeriod && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-end sm:items-center justify-center p-4">
          <Card className="w-full sm:max-w-md rounded-t-3xl sm:rounded-3xl">
            <h2 className="text-xl sm:text-2xl font-bold text-base-content mb-2">Tandai Dibayar</h2>
            <p className="text-sm text-base-content/60 mb-6">
              {payingPeriod.bill.name} · periode {payingPeriod.period.period}
            </p>

            <form onSubmit={handlePay} className="flex flex-col gap-4">
              <FormField label="Dari Akun">
                <Select value={payAccountId} onChange={(e) => setPayAccountId(e.target.value)} disabled={updating}>
                  <option value="">Pilih akun</option>
                  {accounts.map((acc) => (
                    <option key={acc.id} value={acc.id}>
                      {acc.name}
                    </option>
                  ))}
                </Select>
              </FormField>

              <FormField label="Nominal">
                <CurrencyInput
                  value={payAmount}
                  onChange={setPayAmount}
                  disabled={updating}
                />
              </FormField>

              <FormField label="Tanggal Bayar">
                <Input
                  type="date"
                  required
                  value={payDate}
                  onChange={(e) => setPayDate(e.target.value)}
                  disabled={updating}
                />
              </FormField>

              <div className="flex gap-3 pt-2">
                <Button type="button" variant="ghost" fullWidth onClick={() => setShowPayModal(false)} disabled={updating}>
                  Batal
                </Button>
                <Button type="submit" fullWidth disabled={updating}>
                  {updating ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Konfirmasi Bayar'}
                </Button>
              </div>
            </form>
          </Card>
        </div>
      )}

      {/* Edit Modal */}
      {showEditModal && editingBill && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-end sm:items-center justify-center p-4">
          <Card className="w-full sm:max-w-md rounded-t-3xl sm:rounded-3xl max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl sm:text-2xl font-bold text-base-content">Edit Tagihan</h2>
              <button onClick={() => setShowEditModal(false)} className="text-base-content/60 hover:text-base-content">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleEdit} className="flex flex-col gap-4">
              <FormField label="Nama Tagihan">
                <Input
                  type="text"
                  placeholder="Contoh: Cicilan Renovasi"
                  required
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  disabled={updating}
                />
              </FormField>

              <FormField label="Kategori">
                <Select value={editCategoryId} onChange={(e) => setEditCategoryId(e.target.value)} disabled={updating}>
                  {categories.map((cat) => (
                    <option key={cat.id} value={cat.id}>
                      {cat.name}
                    </option>
                  ))}
                </Select>
              </FormField>

              <FormField label="Nominal per Periode" hint="Berlaku untuk periode yang belum dibayar; yang sudah dibayar tidak berubah">
                <CurrencyInput value={editAmount} onChange={setEditAmount} disabled={updating} />
              </FormField>

              <FormField label="Tanggal Jatuh Tempo" hint="1-31, otomatis disesuaikan kalau bulan lebih pendek">
                <Input
                  type="number"
                  placeholder="10"
                  required
                  min="1"
                  max="31"
                  value={editDueDay}
                  onChange={(e) => setEditDueDay(e.target.value)}
                  disabled={updating}
                />
              </FormField>

              <div className="flex gap-3 pt-2">
                <Button type="button" variant="ghost" fullWidth onClick={() => setShowEditModal(false)} disabled={updating}>
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
