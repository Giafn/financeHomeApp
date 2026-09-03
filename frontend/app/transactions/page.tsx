'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiCall, ApiError } from '@/lib/api';
import { Loader2, Plus, ArrowDownCircle, ArrowUpCircle, ArrowLeftRight, ChevronLeft, ChevronRight, Trash2, X, Image as ImageIcon } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Select } from '@/components/ui/Input';
import { Alert } from '@/components/ui/Alert';
import { AppShell } from '@/components/ui/AppShell';

interface TransactionItem {
  id: string;
  type: 'income' | 'expense' | 'transfer';
  account_name: string;
  category_name?: string | null;
  amount: number;
  admin_fee?: number;
  description?: string | null;
  transaction_date: string;
  created_by_name: string;
  attachment_url?: string | null;
}

interface Account {
  id: string;
  name: string;
}

interface Category {
  id: string;
  name: string;
}

const LIMIT = 20;

function formatCurrency(value: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(value);
}

export default function TransactionsPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [selected, setSelected] = useState<TransactionItem | null>(null);

  const [items, setItems] = useState<TransactionItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);

  const [accounts, setAccounts] = useState<Account[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);

  const [typeFilter, setTypeFilter] = useState('');
  const [accountFilter, setAccountFilter] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');

  useEffect(() => {
    const fetchFilters = async () => {
      try {
        const [accountsRes, categoriesRes] = await Promise.all([
          apiCall<Account[]>('/accounts?include_inactive=true'),
          apiCall<Category[]>('/categories?include_archived=true'),
        ]);
        setAccounts(accountsRes || []);
        setCategories(categoriesRes || []);
      } catch {
        // filters are non-critical, ignore failure
      }
    };
    fetchFilters();
  }, []);

  const fetchTransactions = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const params = new URLSearchParams();
      if (typeFilter) params.set('type', typeFilter);
      if (accountFilter) params.set('account_id', accountFilter);
      if (categoryFilter) params.set('category_id', categoryFilter);
      params.set('page', String(page));
      params.set('limit', String(LIMIT));

      const data = await apiCall<{ items: TransactionItem[]; pagination: { total: number } }>(
        `/transactions?${params.toString()}`
      );
      setItems(data.items || []);
      setTotal(data.pagination?.total || 0);
    } catch (err) {
      if (err instanceof ApiError && err.statusCode === 401) {
        router.push('/login');
      } else {
        setError('Gagal memuat transaksi');
      }
    } finally {
      setLoading(false);
    }
  }, [typeFilter, accountFilter, categoryFilter, page, router]);

  useEffect(() => {
    fetchTransactions();
  }, [fetchTransactions]);

  useEffect(() => {
    setPage(1);
  }, [typeFilter, accountFilter, categoryFilter]);

  const handleDelete = async (id: string) => {
    if (!window.confirm('Yakin ingin hapus transaksi ini?')) return;

    setDeletingId(id);
    setError('');

    try {
      await apiCall(`/transactions/${id}`, { method: 'DELETE' });
      fetchTransactions();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal hapus transaksi');
      }
    } finally {
      setDeletingId(null);
    }
  };

  const typeIcon = (type: string) => {
    if (type === 'income') return <ArrowDownCircle className="w-5 h-5 text-success" />;
    if (type === 'expense') return <ArrowUpCircle className="w-5 h-5 text-error" />;
    return <ArrowLeftRight className="w-5 h-5 text-primary" />;
  };

  const amountClass = (type: string) => {
    if (type === 'income') return 'text-success';
    if (type === 'expense') return 'text-error';
    return 'text-base-content';
  };

  const amountPrefix = (type: string) => (type === 'income' ? '+' : type === 'expense' ? '-' : '');

  const totalPages = Math.max(1, Math.ceil(total / LIMIT));

  return (
    <AppShell active="/transactions">
      <div className="border-b border-base-300 bg-base-200 sticky top-0 z-50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
          <h1 className="text-2xl sm:text-3xl font-bold text-base-content">Transaksi</h1>
          <Link href="/transactions/new">
            <Button>
              <Plus className="w-4 h-4" />
              Tambah
            </Button>
          </Link>
        </div>
      </div>

      <div className="max-w-4xl mx-auto p-4 sm:p-6">
        {error && <Alert type="error" message={error} />}

        {/* Filters */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 mb-6">
          <Select value={typeFilter} onChange={(e) => setTypeFilter(e.target.value)}>
            <option value="">Semua Tipe</option>
            <option value="income">Pemasukan</option>
            <option value="expense">Pengeluaran</option>
            <option value="transfer">Transfer</option>
          </Select>
          <Select value={accountFilter} onChange={(e) => setAccountFilter(e.target.value)}>
            <option value="">Semua Akun</option>
            {accounts.map((acc) => (
              <option key={acc.id} value={acc.id}>
                {acc.name}
              </option>
            ))}
          </Select>
          <Select value={categoryFilter} onChange={(e) => setCategoryFilter(e.target.value)}>
            <option value="">Semua Kategori</option>
            {categories.map((cat) => (
              <option key={cat.id} value={cat.id}>
                {cat.name}
              </option>
            ))}
          </Select>
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="w-8 h-8 animate-spin text-primary" />
          </div>
        ) : items.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-base-content/60 mb-4">Belum ada transaksi</p>
            <Link href="/transactions/new">
              <Button>
                <Plus className="w-4 h-4" />
                Catat Transaksi Pertama
              </Button>
            </Link>
          </div>
        ) : (
          <>
            <div className="flex flex-col gap-3 mb-6">
              {items.map((tx) => (
                <Card key={tx.id} className="!p-4">
                  <div className="flex items-center gap-3">
                    <button
                      onClick={() => setSelected(tx)}
                      className="flex items-center gap-3 flex-1 min-w-0 text-left"
                    >
                      {typeIcon(tx.type)}
                      <div className="flex-1 min-w-0">
                        <p className="font-medium text-base-content truncate">
                          {tx.description || tx.category_name || (tx.type === 'transfer' ? 'Transfer' : 'Transaksi')}
                        </p>
                        <p className="text-xs text-base-content/60">
                          {tx.account_name} · {new Date(tx.transaction_date).toLocaleDateString('id-ID')} · {tx.created_by_name}
                        </p>
                      </div>
                      <p className={`font-semibold whitespace-nowrap ${amountClass(tx.type)}`}>
                        {amountPrefix(tx.type)}
                        {formatCurrency(tx.amount)}
                      </p>
                    </button>
                    <button
                      onClick={() => handleDelete(tx.id)}
                      disabled={deletingId === tx.id}
                      className="inline-flex items-center justify-center w-8 h-8 rounded-full text-base-content/40 hover:text-error hover:bg-error/10 transition-colors flex-shrink-0"
                    >
                      {deletingId === tx.id ? (
                        <Loader2 className="w-4 h-4 animate-spin" />
                      ) : (
                        <Trash2 className="w-4 h-4" />
                      )}
                    </button>
                  </div>
                </Card>
              ))}
            </div>

            {/* Pagination */}
            <div className="flex items-center justify-between">
              <p className="text-sm text-base-content/60">
                Halaman {page} dari {totalPages} · {total} transaksi
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page <= 1}
                >
                  <ChevronLeft className="w-4 h-4" />
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page >= totalPages}
                >
                  <ChevronRight className="w-4 h-4" />
                </Button>
              </div>
            </div>
          </>
        )}
      </div>

      {/* Detail Modal */}
      {selected && (
        <div
          className="fixed inset-0 bg-black/60 z-50 flex items-end sm:items-center justify-center p-4"
          onClick={() => setSelected(null)}
        >
          <Card
            className="w-full sm:max-w-md rounded-t-3xl sm:rounded-3xl"
            onClick={(e: React.MouseEvent) => e.stopPropagation()}
          >
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-2">
                {typeIcon(selected.type)}
                <h2 className="text-xl sm:text-2xl font-bold text-base-content">
                  Detail Transaksi
                </h2>
              </div>
              <button
                onClick={() => setSelected(null)}
                className="inline-flex items-center justify-center w-8 h-8 rounded-full text-base-content/50 hover:text-base-content hover:bg-base-300 transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="flex flex-col gap-3">
              <div>
                <p className="text-sm text-base-content/60">Tipe</p>
                <p className="font-medium text-base-content capitalize">
                  {selected.type === 'income'
                    ? 'Pemasukan'
                    : selected.type === 'expense'
                      ? 'Pengeluaran'
                      : 'Transfer'}
                </p>
              </div>
              <div>
                <p className="text-sm text-base-content/60">Akun</p>
                <p className="font-medium text-base-content">{selected.account_name}</p>
              </div>
              {selected.category_name && (
                <div>
                  <p className="text-sm text-base-content/60">Kategori</p>
                  <p className="font-medium text-base-content">{selected.category_name}</p>
                </div>
              )}
              <div>
                <p className="text-sm text-base-content/60">Jumlah</p>
                <p className={`text-2xl font-bold ${amountClass(selected.type)}`}>
                  {amountPrefix(selected.type)}
                  {formatCurrency(selected.amount)}
                </p>
              </div>
              {selected.type === 'transfer' && (selected.admin_fee ?? 0) > 0 && (
                <div>
                  <p className="text-sm text-base-content/60">Biaya Admin</p>
                  <p className="font-medium text-base-content">{formatCurrency(selected.admin_fee ?? 0)}</p>
                </div>
              )}
              {selected.description && (
                <div>
                  <p className="text-sm text-base-content/60">Deskripsi</p>
                  <p className="font-medium text-base-content">{selected.description}</p>
                </div>
              )}
              <div>
                <p className="text-sm text-base-content/60">Tanggal</p>
                <p className="font-medium text-base-content">
                  {new Date(selected.transaction_date).toLocaleDateString('id-ID', {
                    weekday: 'long',
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric',
                  })}
                </p>
              </div>
              <div>
                <p className="text-sm text-base-content/60">Dicatat oleh</p>
                <p className="font-medium text-base-content">{selected.created_by_name}</p>
              </div>

              {selected.attachment_url ? (
                <div>
                  <p className="text-sm text-base-content/60 mb-2">Bukti Pembayaran</p>
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={selected.attachment_url}
                    alt="Bukti transaksi"
                    className="w-full rounded-xl object-contain bg-base-200 max-h-80"
                  />
                </div>
              ) : (
                <div className="flex items-center gap-2 text-sm text-base-content/40 py-2">
                  <ImageIcon className="w-4 h-4" />
                  Tidak ada bukti
                </div>
              )}
            </div>
          </Card>
        </div>
      )}

      </AppShell>
  );
}
