'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiCall, ApiError } from '@/lib/api';
import {
  Menu, X, Settings, LogOut, Loader2, Wallet, Target,
  ArrowDownCircle, ArrowUpCircle, ArrowLeftRight,
} from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { FormField } from '@/components/ui/FormField';
import { Alert } from '@/components/ui/Alert';
import { BottomNav } from '@/components/ui/BottomNav';

interface UserProfile {
  id: string;
  name: string;
  email: string;
  household: { id: string; name: string; role: string } | null;
}

interface DashboardAccount {
  id: string;
  name: string;
  current_balance: number;
}

interface BudgetSummary {
  total_budget: number;
  total_spent: number;
  percentage: number;
}

interface UpcomingBill {
  bill_period_id: string;
  bill_name: string;
  amount: number;
  due_date: string;
}

interface GoalSummary {
  id: string;
  name: string;
  percentage: number;
}

interface MonthlyTrendPoint {
  month: string;
  income: number;
  expense: number;
}

interface MemberBreakdownItem {
  user_id: string;
  name: string;
  total_expense: number;
  total_income: number;
}

interface RecentTransaction {
  id: string;
  type: 'income' | 'expense' | 'transfer';
  amount: number;
  category_name?: string | null;
  account_name: string;
  transaction_date: string;
  description?: string | null;
}

interface DashboardData {
  total_balance: number;
  accounts: DashboardAccount[];
  budget_summary: BudgetSummary;
  upcoming_bills: UpcomingBill[];
  active_goals: GoalSummary[];
  monthly_trend: MonthlyTrendPoint[];
  member_breakdown: MemberBreakdownItem[];
  recent_transactions: RecentTransaction[];
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

function formatCurrencyShort(value: number) {
  if (Math.abs(value) >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}jt`;
  if (Math.abs(value) >= 1_000) return `${(value / 1_000).toFixed(0)}rb`;
  return String(value);
}

function monthLabel(month: string) {
  const [y, m] = month.split('-').map(Number);
  return new Date(y, m - 1, 1).toLocaleDateString('id-ID', { month: 'short' });
}

export default function DashboardPage() {
  const router = useRouter();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [dashboard, setDashboard] = useState<DashboardData | null>(null);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [menuOpen, setMenuOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [updating, setUpdating] = useState(false);

  const [showPayModal, setShowPayModal] = useState(false);
  const [payingBill, setPayingBill] = useState<UpcomingBill | null>(null);
  const [payAccountId, setPayAccountId] = useState('');
  const [payAmount, setPayAmount] = useState('');
  const [payDate, setPayDate] = useState('');

  const fetchAll = async () => {
    setLoading(true);
    setError('');
    try {
      const [profileRes, dashboardRes, accountsRes] = await Promise.all([
        apiCall<UserProfile>('/users/me'),
        apiCall<DashboardData>('/dashboard'),
        apiCall<Account[]>('/accounts'),
      ]);
      setProfile(profileRes);
      setDashboard(dashboardRes);
      setAccounts(accountsRes || []);
    } catch (err) {
      if (err instanceof ApiError && err.statusCode === 401) {
        router.push('/login');
      } else {
        setError('Gagal memuat dashboard');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/login');
      return;
    }
    fetchAll();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('token');
    router.push('/login');
  };

  const openPayModal = (bill: UpcomingBill) => {
    setPayingBill(bill);
    setPayAccountId(accounts[0]?.id || '');
    setPayAmount(String(bill.amount));
    const today = new Date();
    setPayDate(`${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`);
    setShowPayModal(true);
  };

  const handlePay = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!payingBill || !payAccountId || !payAmount || parseFloat(payAmount) <= 0) {
      setError('Akun dan nominal wajib diisi dengan benar');
      return;
    }

    setError('');
    setSuccess('');
    setUpdating(true);

    try {
      await apiCall(`/bill-periods/${payingBill.bill_period_id}/pay`, {
        method: 'POST',
        body: JSON.stringify({
          account_id: payAccountId,
          amount: parseFloat(payAmount),
          transaction_date: payDate,
        }),
      });
      setSuccess('Tagihan berhasil ditandai dibayar');
      setShowPayModal(false);
      fetchAll();
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

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-base-100">
        <div className="flex flex-col items-center gap-4">
          <Loader2 className="w-8 h-8 animate-spin text-primary" />
          <p className="text-base-content/60">Memuat dashboard...</p>
        </div>
      </div>
    );
  }

  const maxTrendValue = dashboard
    ? Math.max(1, ...dashboard.monthly_trend.flatMap((t) => [t.income, t.expense]))
    : 1;

  const typeIcon = (type: string) => {
    if (type === 'income') return <ArrowDownCircle className="w-4 h-4 text-success" />;
    if (type === 'expense') return <ArrowUpCircle className="w-4 h-4 text-error" />;
    return <ArrowLeftRight className="w-4 h-4 text-primary" />;
  };

  return (
    <div className="min-h-screen bg-base-100">
      {/* Topbar */}
      <div className="border-b border-base-300 sticky top-0 z-50 bg-base-200">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
          <div className="flex-1">
            <h1 className="text-xl sm:text-2xl font-bold text-base-content">Family Finance</h1>
            {profile?.household && (
              <p className="text-xs sm:text-sm text-base-content/60 mt-0.5">{profile.household.name}</p>
            )}
          </div>

          <div className="hidden sm:flex items-center gap-2 sm:gap-3">
            <Link href="/settings/profile">
              <Button variant="ghost">
                <Settings className="w-4 h-4" />
                {profile?.name || 'Profil'}
              </Button>
            </Link>
            <Button variant="outline" onClick={handleLogout}>
              <LogOut className="w-4 h-4" />
              Keluar
            </Button>
          </div>

          <div className="sm:hidden">
            <button
              className="inline-flex items-center justify-center w-10 h-10 rounded-full hover:bg-base-300 transition-colors"
              onClick={() => setMenuOpen(!menuOpen)}
            >
              {menuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </button>
          </div>
        </div>

        {menuOpen && (
          <div className="sm:hidden border-t border-base-300 p-4 flex flex-col gap-2">
            <Link href="/settings/profile">
              <Button variant="ghost" fullWidth className="justify-start">
                <Settings className="w-4 h-4" />
                Profil
              </Button>
            </Link>
            <Button variant="outline" fullWidth className="justify-start" onClick={handleLogout}>
              <LogOut className="w-4 h-4" />
              Keluar
            </Button>
          </div>
        )}
      </div>

      <div className="max-w-6xl mx-auto px-4 sm:px-6 py-6 sm:py-8 pb-24">
        {error && <Alert type="error" message={error} />}
        {success && <Alert type="success" message={success} />}

        <div className="mb-6 sm:mb-8">
          <h2 className="text-2xl sm:text-3xl font-bold text-base-content mb-1">
            Selamat datang, {profile?.name}!
          </h2>
        </div>

        {dashboard && (
          <div className="flex flex-col gap-6 sm:gap-8">
            {/* 1. Total saldo + breakdown akun */}
            <Card>
              <div className="flex items-center gap-2 mb-4">
                <Wallet className="w-5 h-5 text-primary" />
                <h3 className="text-lg sm:text-xl font-bold text-base-content">Total Saldo</h3>
              </div>
              <p className="text-3xl sm:text-4xl font-bold text-base-content mb-4">
                {formatCurrency(dashboard.total_balance)}
              </p>
              <div className="flex flex-col gap-2">
                {dashboard.accounts.map((acc) => (
                  <div key={acc.id} className="flex items-center justify-between text-sm">
                    <span className="text-base-content/60">{acc.name}</span>
                    <span className="font-medium text-base-content">{formatCurrency(acc.current_balance)}</span>
                  </div>
                ))}
              </div>
            </Card>

            {/* 2. Budget bulan ini */}
            <Card>
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg sm:text-xl font-bold text-base-content">Budget Bulan Ini</h3>
                <Link href="/budgets" className="text-xs text-primary hover:underline">
                  Detail
                </Link>
              </div>
              {dashboard.budget_summary.total_budget > 0 ? (
                <>
                  <div className="flex items-baseline justify-between mb-2">
                    <p className="text-sm text-base-content/60">
                      {formatCurrency(dashboard.budget_summary.total_spent)} dari{' '}
                      {formatCurrency(dashboard.budget_summary.total_budget)}
                    </p>
                    <p
                      className={`text-sm font-bold ${
                        dashboard.budget_summary.percentage > 100
                          ? 'text-error'
                          : dashboard.budget_summary.percentage >= 80
                            ? 'text-warning'
                            : 'text-success'
                      }`}
                    >
                      {dashboard.budget_summary.percentage.toFixed(0)}%
                    </p>
                  </div>
                  <div className="w-full h-2.5 rounded-full bg-base-300 overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all ${
                        dashboard.budget_summary.percentage > 100
                          ? 'bg-error'
                          : dashboard.budget_summary.percentage >= 80
                            ? 'bg-warning'
                            : 'bg-success'
                      }`}
                      style={{ width: `${Math.min(100, dashboard.budget_summary.percentage)}%` }}
                    />
                  </div>
                </>
              ) : (
                <p className="text-sm text-base-content/60">Belum ada budget bulan ini</p>
              )}
            </Card>

            {/* 3. Tagihan mendatang */}
            <Card>
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg sm:text-xl font-bold text-base-content">Tagihan 7 Hari Ke Depan</h3>
                <Link href="/bills" className="text-xs text-primary hover:underline">
                  Lihat Semua
                </Link>
              </div>
              {dashboard.upcoming_bills.length === 0 ? (
                <p className="text-sm text-base-content/60">Tidak ada tagihan mendatang</p>
              ) : (
                <div className="flex flex-col gap-3">
                  {dashboard.upcoming_bills.map((bill) => (
                    <div
                      key={bill.bill_period_id}
                      className="flex items-center justify-between px-4 py-3 bg-base-100 border border-base-300 rounded-2xl"
                    >
                      <div>
                        <p className="text-sm font-medium text-base-content">{bill.bill_name}</p>
                        <p className="text-xs text-base-content/60">
                          {formatCurrency(bill.amount)} · jatuh tempo{' '}
                          {new Date(bill.due_date).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })}
                        </p>
                      </div>
                      <Button size="sm" onClick={() => openPayModal(bill)} disabled={updating || accounts.length === 0}>
                        Tandai Dibayar
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </Card>

            {/* 4. Progress goals */}
            <Card>
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg sm:text-xl font-bold text-base-content">Goals</h3>
                <Link href="/goals" className="text-xs text-primary hover:underline">
                  Lihat Semua
                </Link>
              </div>
              {dashboard.active_goals.length === 0 ? (
                <p className="text-sm text-base-content/60">Belum ada goal aktif</p>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  {dashboard.active_goals.map((g) => (
                    <Link key={g.id} href={`/goals/${g.id}`}>
                      <div className="px-4 py-3 bg-base-100 border border-base-300 rounded-2xl hover:border-primary transition-colors">
                        <div className="flex items-center gap-2 mb-2">
                          <Target className="w-4 h-4 text-primary" />
                          <p className="text-sm font-medium text-base-content truncate">{g.name}</p>
                        </div>
                        <div className="w-full h-2 rounded-full bg-base-300 overflow-hidden mb-1">
                          <div
                            className="h-full rounded-full bg-success transition-all"
                            style={{ width: `${Math.min(100, g.percentage)}%` }}
                          />
                        </div>
                        <p className="text-xs text-base-content/60">{g.percentage.toFixed(0)}%</p>
                      </div>
                    </Link>
                  ))}
                </div>
              )}
            </Card>

            {/* 5. Tren 6 bulan */}
            <Card>
              <h3 className="text-lg sm:text-xl font-bold text-base-content mb-4">Tren 6 Bulan Terakhir</h3>
              <div className="flex items-end justify-between gap-2 h-40 mb-2">
                {dashboard.monthly_trend.map((t) => (
                  <div key={t.month} className="flex-1 flex flex-col items-center gap-1 h-full justify-end">
                    <div className="w-full flex gap-1 items-end justify-center h-full">
                      <div
                        className="w-1/2 max-w-3 bg-success rounded-t"
                        style={{ height: `${(t.income / maxTrendValue) * 100}%` }}
                        title={`Income: ${formatCurrency(t.income)}`}
                      />
                      <div
                        className="w-1/2 max-w-3 bg-error rounded-t"
                        style={{ height: `${(t.expense / maxTrendValue) * 100}%` }}
                        title={`Expense: ${formatCurrency(t.expense)}`}
                      />
                    </div>
                    <p className="text-xs text-base-content/60">{monthLabel(t.month)}</p>
                  </div>
                ))}
              </div>
              <div className="flex items-center gap-4 justify-center text-xs text-base-content/60">
                <span className="flex items-center gap-1.5">
                  <span className="w-2.5 h-2.5 rounded-full bg-success" /> Pemasukan
                </span>
                <span className="flex items-center gap-1.5">
                  <span className="w-2.5 h-2.5 rounded-full bg-error" /> Pengeluaran
                </span>
              </div>
            </Card>

            {/* 6. Kontribusi pengeluaran per anggota */}
            <Card>
              <h3 className="text-lg sm:text-xl font-bold text-base-content mb-4">Kontribusi Pengeluaran Bulan Ini</h3>
              {dashboard.member_breakdown.length === 0 ? (
                <p className="text-sm text-base-content/60">Belum ada data</p>
              ) : (
                <div className="flex flex-col gap-3">
                  {dashboard.member_breakdown.map((m) => {
                    const totalHouseholdExpense = dashboard.member_breakdown.reduce((sum, x) => sum + x.total_expense, 0);
                    const share = totalHouseholdExpense > 0 ? (m.total_expense / totalHouseholdExpense) * 100 : 0;
                    return (
                      <div key={m.user_id}>
                        <div className="flex items-center justify-between text-sm mb-1">
                          <span className="text-base-content">{m.name}</span>
                          <span className="font-medium text-base-content">{formatCurrency(m.total_expense)}</span>
                        </div>
                        <div className="w-full h-2 rounded-full bg-base-300 overflow-hidden">
                          <div className="h-full rounded-full bg-primary transition-all" style={{ width: `${share}%` }} />
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </Card>

            {/* 7. Transaksi terbaru */}
            <Card>
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg sm:text-xl font-bold text-base-content">Transaksi Terbaru</h3>
                <Link href="/transactions" className="text-xs text-primary hover:underline">
                  Lihat Semua
                </Link>
              </div>
              {dashboard.recent_transactions.length === 0 ? (
                <p className="text-sm text-base-content/60">Belum ada transaksi</p>
              ) : (
                <div className="flex flex-col gap-3">
                  {dashboard.recent_transactions.map((tx) => (
                    <div key={tx.id} className="flex items-center gap-3">
                      {typeIcon(tx.type)}
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-base-content truncate">
                          {tx.description || tx.category_name || (tx.type === 'transfer' ? 'Transfer' : 'Transaksi')}
                        </p>
                        <p className="text-xs text-base-content/60">
                          {tx.account_name} · {new Date(tx.transaction_date).toLocaleDateString('id-ID')}
                        </p>
                      </div>
                      <p
                        className={`text-sm font-semibold whitespace-nowrap ${
                          tx.type === 'income' ? 'text-success' : tx.type === 'expense' ? 'text-error' : 'text-base-content'
                        }`}
                      >
                        {tx.type === 'income' ? '+' : tx.type === 'expense' ? '-' : ''}
                        {formatCurrencyShort(tx.amount)}
                      </p>
                    </div>
                  ))}
                </div>
              )}
            </Card>
          </div>
        )}
      </div>

      {/* Pay bill modal */}
      {showPayModal && payingBill && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-end sm:items-center justify-center p-4">
          <Card className="w-full sm:max-w-md rounded-t-3xl sm:rounded-3xl">
            <h2 className="text-xl sm:text-2xl font-bold text-base-content mb-2">Tandai Dibayar</h2>
            <p className="text-sm text-base-content/60 mb-6">{payingBill.bill_name}</p>

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
                <Input
                  type="number"
                  required
                  step="0.01"
                  min="0.01"
                  value={payAmount}
                  onChange={(e) => setPayAmount(e.target.value)}
                  disabled={updating}
                />
              </FormField>

              <FormField label="Tanggal Bayar">
                <Input type="date" required value={payDate} onChange={(e) => setPayDate(e.target.value)} disabled={updating} />
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

      <BottomNav active="/dashboard" />
    </div>
  );
}
