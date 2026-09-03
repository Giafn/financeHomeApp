'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { apiCall, ApiError } from '@/lib/api';
import { Loader2, Paperclip, X } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { CurrencyInput } from '@/components/ui/CurrencyInput';
import { FormField } from '@/components/ui/FormField';
import { Alert } from '@/components/ui/Alert';
import { TopBar } from '@/components/ui/TopBar';
import { AppShell } from '@/components/ui/AppShell';

interface Account {
  id: string;
  name: string;
  type: string;
  is_active: boolean;
}

interface Category {
  id: string;
  name: string;
  type: 'income' | 'expense';
  is_archived: boolean;
}

type TxType = 'income' | 'expense' | 'transfer';

function todayStr() {
  const d = new Date();
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${d.getFullYear()}-${month}-${day}`;
}

export default function NewTransactionPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState('');

  const [accounts, setAccounts] = useState<Account[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);

  const [type, setType] = useState<TxType>('expense');
  const [accountId, setAccountId] = useState('');
  const [destinationAccountId, setDestinationAccountId] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [amount, setAmount] = useState('');
  const [adminFee, setAdminFee] = useState('0');
  const [description, setDescription] = useState('');
  const [transactionDate, setTransactionDate] = useState(todayStr());
  const [attachmentUrl, setAttachmentUrl] = useState<string | null>(null);
  const [attachmentName, setAttachmentName] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [accountsRes, categoriesRes, quickSelect] = await Promise.all([
          apiCall<Account[]>('/accounts'),
          apiCall<Category[]>('/categories?include_archived=false'),
          apiCall<{ last_account_id?: string; last_category_id?: string }>('/transactions/quick-select'),
        ]);

        const activeAccounts = accountsRes || [];
        setAccounts(activeAccounts);
        setCategories(categoriesRes || []);

        if (quickSelect?.last_account_id && activeAccounts.some((a) => a.id === quickSelect.last_account_id)) {
          setAccountId(quickSelect.last_account_id);
        } else if (activeAccounts.length > 0) {
          setAccountId(activeAccounts[0].id);
        }
        if (quickSelect?.last_category_id) {
          setCategoryId(quickSelect.last_category_id);
        }
      } catch (err) {
        if (err instanceof ApiError && err.statusCode === 401) {
          router.push('/login');
        } else {
          setError('Gagal memuat data akun/kategori');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [router]);

  const filteredCategories = categories.filter((c) => c.type === type);

  useEffect(() => {
    if (type === 'transfer') return;
    if (!filteredCategories.some((c) => c.id === categoryId)) {
      setCategoryId(filteredCategories[0]?.id || '');
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [type]);

  const handleAttachmentChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setError('');
    setUploading(true);

    try {
      const presign = await apiCall<{ upload_url: string; file_url: string }>('/uploads/presign', {
        method: 'POST',
        body: JSON.stringify({ filename: file.name, content_type: file.type }),
      });

      await fetch(presign.upload_url, {
        method: 'PUT',
        headers: { 'Content-Type': file.type },
        body: file,
      });

      setAttachmentUrl(presign.file_url);
      setAttachmentName(file.name);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.statusCode === 503 ? 'Upload lampiran belum tersedia di server ini' : err.message);
      } else {
        setError('Gagal upload lampiran');
      }
    } finally {
      setUploading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!accountId) {
      setError('Akun wajib dipilih');
      return;
    }
    if (!amount || parseFloat(amount) <= 0) {
      setError('Nominal harus lebih dari 0');
      return;
    }
    if (type === 'transfer') {
      if (!destinationAccountId) {
        setError('Akun tujuan wajib dipilih');
        return;
      }
      if (destinationAccountId === accountId) {
        setError('Akun tujuan tidak boleh sama dengan akun sumber');
        return;
      }
    } else if (!categoryId) {
      setError('Kategori wajib dipilih');
      return;
    }

    setError('');
    setSubmitting(true);

    try {
      const body: Record<string, unknown> = {
        type,
        account_id: accountId,
        amount: parseFloat(amount),
        transaction_date: transactionDate,
        description: description || null,
        attachment_url: attachmentUrl,
      };
      if (type === 'transfer') {
        body.destination_account_id = destinationAccountId;
        body.admin_fee = parseFloat(adminFee) || 0;
      } else {
        body.category_id = categoryId;
      }

      await apiCall('/transactions', {
        method: 'POST',
        body: JSON.stringify(body),
      });

      router.push('/transactions');
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal membuat transaksi');
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-base-100">
        <div className="flex flex-col items-center gap-4">
          <Loader2 className="w-8 h-8 animate-spin text-primary" />
          <p className="text-base-content/60">Memuat form...</p>
        </div>
      </div>
    );
  }

  return (
    <AppShell active="/transactions">
      <TopBar title="Transaksi Baru" subtitle="Catat pemasukan, pengeluaran, atau transfer" backHref="/transactions" />

      <div className="max-w-4xl mx-auto p-4 sm:p-6">
        {error && <Alert type="error" message={error} />}

        <Card>
          {/* Type selector */}
          <div className="grid grid-cols-3 gap-2 mb-6">
            {(['expense', 'income', 'transfer'] as TxType[]).map((t) => (
              <button
                key={t}
                type="button"
                onClick={() => setType(t)}
                className={`py-3 rounded-2xl font-medium text-sm transition-colors ${
                  type === t
                    ? 'bg-primary text-primary-content'
                    : 'bg-base-100 text-base-content border border-base-300 hover:bg-base-300'
                }`}
              >
                {t === 'expense' ? 'Pengeluaran' : t === 'income' ? 'Pemasukan' : 'Transfer'}
              </button>
            ))}
          </div>

          <form onSubmit={handleSubmit} className="flex flex-col gap-5">
            <FormField label={type === 'transfer' ? 'Akun Sumber' : type === 'income' ? 'Akun Tujuan' : 'Akun Sumber'}>
              <Select value={accountId} onChange={(e) => setAccountId(e.target.value)} disabled={submitting}>
                <option value="">Pilih akun</option>
                {accounts.map((acc) => (
                  <option key={acc.id} value={acc.id}>
                    {acc.name}
                  </option>
                ))}
              </Select>
            </FormField>

            {type === 'transfer' ? (
              <FormField label="Akun Tujuan">
                <Select value={destinationAccountId} onChange={(e) => setDestinationAccountId(e.target.value)} disabled={submitting}>
                  <option value="">Pilih akun tujuan</option>
                  {accounts.filter((a) => a.id !== accountId).map((acc) => (
                    <option key={acc.id} value={acc.id}>
                      {acc.name}
                    </option>
                  ))}
                </Select>
              </FormField>
            ) : (
              <FormField label="Kategori">
                <Select value={categoryId} onChange={(e) => setCategoryId(e.target.value)} disabled={submitting}>
                  <option value="">Pilih kategori</option>
                  {filteredCategories.map((cat) => (
                    <option key={cat.id} value={cat.id}>
                      {cat.name}
                    </option>
                  ))}
                </Select>
              </FormField>
            )}

            <FormField label="Nominal">
              <CurrencyInput
                placeholder="50.000"
                value={amount}
                onChange={setAmount}
                disabled={submitting}
                autoFocus
              />
            </FormField>

            {type === 'transfer' && (
              <FormField label="Biaya Admin" hint="Opsional — default 0">
                <CurrencyInput
                  placeholder="0"
                  value={adminFee}
                  onChange={setAdminFee}
                  disabled={submitting}
                />
              </FormField>
            )}

            <FormField label="Tanggal">
              <Input
                type="date"
                required
                value={transactionDate}
                onChange={(e) => setTransactionDate(e.target.value)}
                disabled={submitting}
              />
            </FormField>

            <FormField label="Deskripsi" hint="Opsional">
              <Input
                type="text"
                placeholder="Contoh: Makan siang"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                disabled={submitting}
              />
            </FormField>

            <FormField label="Lampiran" hint="Opsional — foto struk">
              {attachmentName ? (
                <div className="flex items-center justify-between px-4 h-12 rounded-2xl bg-base-100 border border-base-300">
                  <span className="text-sm text-base-content truncate">{attachmentName}</span>
                  <button
                    type="button"
                    onClick={() => {
                      setAttachmentUrl(null);
                      setAttachmentName(null);
                    }}
                    className="text-base-content/60 hover:text-error"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              ) : (
                <label className="flex items-center justify-center gap-2 h-12 rounded-2xl bg-base-100 border border-dashed border-base-300 text-base-content/60 cursor-pointer hover:bg-base-300 transition-colors">
                  {uploading ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Mengupload...
                    </>
                  ) : (
                    <>
                      <Paperclip className="w-4 h-4" />
                      Pilih foto struk
                    </>
                  )}
                  <input
                    type="file"
                    accept="image/*"
                    className="hidden"
                    onChange={handleAttachmentChange}
                    disabled={submitting || uploading}
                  />
                </label>
              )}
            </FormField>

            <Button type="submit" fullWidth disabled={submitting || uploading} className="mt-2">
              {submitting ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Menyimpan...
                </>
              ) : (
                'Simpan Transaksi'
              )}
            </Button>
          </form>
        </Card>
      </div>
    </AppShell>
  );
}
