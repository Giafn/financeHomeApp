'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { apiCall, ApiError } from '@/lib/api';
import { Loader2, Plus, Palette, Trash2, RotateCcw } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { FormField } from '@/components/ui/FormField';
import { Alert } from '@/components/ui/Alert';
import { AppShell } from '@/components/ui/AppShell';

interface Category {
  id: string;
  name: string;
  type: 'income' | 'expense';
  icon: string | null;
  color: string | null;
  is_archived: boolean;
}

export default function CategoriesPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [categories, setCategories] = useState<Category[]>([]);
  const [activeTab, setActiveTab] = useState<'income' | 'expense'>('income');
  const [showModal, setShowModal] = useState(false);
  const [archivingId, setArchivingId] = useState<string | null>(null);
  const [unarchivingId, setUnarchivingId] = useState<string | null>(null);

  const [formData, setFormData] = useState({
    name: '',
    type: 'income' as 'income' | 'expense',
    icon: '',
    color: '#4CAF50',
  });

  useEffect(() => {
    const fetchCategories = async () => {
      try {
        const token = localStorage.getItem('token');
        if (!token) {
          router.push('/login');
          return;
        }

        const data = await apiCall<Category[]>('/categories?include_archived=true');
        setCategories(data || []);
      } catch (err) {
        if (err instanceof ApiError && err.statusCode === 401) {
          router.push('/login');
        } else {
          setError('Gagal memuat kategori');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchCategories();
  }, [router]);

  const handleCreateCategory = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name.trim()) {
      setError('Nama kategori diperlukan');
      return;
    }

    setError('');
    setSuccess('');
    setUpdating(true);

    try {
      const newCategory = await apiCall<Category>('/categories', {
        method: 'POST',
        body: JSON.stringify({
          name: formData.name,
          type: formData.type,
          icon: formData.icon || null,
          color: formData.color || null,
        }),
      });

      setCategories([...categories, newCategory]);
      setSuccess('Kategori berhasil dibuat');
      setFormData({ name: '', type: activeTab, icon: '', color: '#4CAF50' });
      setShowModal(false);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal membuat kategori');
      }
    } finally {
      setUpdating(false);
    }
  };

  const handleArchiveCategory = async (categoryId: string) => {
    if (!window.confirm('Yakin ingin arsipkan kategori ini?')) {
      return;
    }

    setError('');
    setSuccess('');
    setUpdating(true);
    setArchivingId(categoryId);

    try {
      await apiCall(`/categories/${categoryId}`, {
        method: 'DELETE',
      });

      setCategories(
        categories.map((cat) =>
          cat.id === categoryId ? { ...cat, is_archived: true } : cat
        )
      );
      setSuccess('Kategori berhasil diarsipkan');
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal arsipkan kategori');
      }
    } finally {
      setUpdating(false);
      setArchivingId(null);
    }
  };

  const handleUnarchiveCategory = async (categoryId: string) => {
    setError('');
    setSuccess('');
    setUpdating(true);
    setUnarchivingId(categoryId);

    try {
      await apiCall(`/categories/${categoryId}/unarchive`, {
        method: 'POST',
      });

      setCategories(
        categories.map((cat) =>
          cat.id === categoryId ? { ...cat, is_archived: false } : cat
        )
      );
      setSuccess('Kategori berhasil diaktifkan');
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal aktifkan kategori');
      }
    } finally {
      setUpdating(false);
      setUnarchivingId(null);
    }
  };

  const filteredCategories = categories.filter((cat) => cat.type === activeTab);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-base-100">
        <div className="flex flex-col items-center gap-4">
          <Loader2 className="w-8 h-8 animate-spin text-primary" />
          <p className="text-base-content/60">Memuat kategori...</p>
        </div>
      </div>
    );
  }

  return (
    <AppShell active="/categories">
      {/* Topbar */}
      <div className="border-b border-base-300 bg-base-200 sticky top-0 z-50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
          <h1 className="text-2xl sm:text-3xl font-bold text-base-content">Kategori</h1>
          <Button onClick={() => {
            setFormData({ name: '', type: activeTab, icon: '', color: '#4CAF50' });
            setShowModal(true);
          }}>
            <Plus className="w-4 h-4" />
            Tambah
          </Button>
        </div>
      </div>

      <div className="max-w-4xl mx-auto p-4 sm:p-6">
        {error && <Alert type="error" message={error} />}
        {success && <Alert type="success" message={success} />}

        {/* Tab Navigation */}
        <div className="flex gap-2 mb-6 sm:mb-8">
          <button
            onClick={() => setActiveTab('income')}
            className={`px-4 sm:px-6 py-2 sm:py-3 rounded-2xl font-medium transition-colors ${
              activeTab === 'income'
                ? 'bg-primary text-primary-content'
                : 'bg-base-200 text-base-content hover:bg-base-300'
            }`}
          >
            Pemasukan
          </button>
          <button
            onClick={() => setActiveTab('expense')}
            className={`px-4 sm:px-6 py-2 sm:py-3 rounded-2xl font-medium transition-colors ${
              activeTab === 'expense'
                ? 'bg-primary text-primary-content'
                : 'bg-base-200 text-base-content hover:bg-base-300'
            }`}
          >
            Pengeluaran
          </button>
        </div>

        {/* Categories Grid */}
        {filteredCategories.length === 0 ? (
          <div className="text-center py-12">
            <Palette className="w-12 h-12 mx-auto mb-4 text-base-content/40" />
            <p className="text-base-content/60 mb-4">Belum ada kategori</p>
            <Button onClick={() => {
              setFormData({ name: '', type: activeTab, icon: '', color: '#4CAF50' });
              setShowModal(true);
            }}>
              <Plus className="w-4 h-4" />
              Buat Kategori Pertama
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
            {filteredCategories.map((category) => (
              <Card key={category.id} className={category.is_archived ? 'opacity-50' : ''}>
                <div className="flex items-start justify-between gap-4 mb-4">
                  <div className="flex items-center gap-3 flex-1">
                    {category.color && (
                      <div
                        className="w-10 h-10 sm:w-12 sm:h-12 rounded-full flex items-center justify-center"
                        style={{ backgroundColor: category.color }}
                      >
                        {category.icon && (
                          <Palette className="w-5 h-5 sm:w-6 sm:h-6 text-white" />
                        )}
                      </div>
                    )}
                    <div className="flex-1">
                      <h3 className="font-semibold text-base-content text-sm sm:text-base">{category.name}</h3>
                      {category.is_archived && (
                        <span className="text-xs px-2 py-1 bg-base-300 text-base-content/60 rounded-full">
                          Diarsipkan
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                <div className="pt-4 border-t border-base-300 flex gap-2">
                  {category.is_archived ? (
                    <Button
                      variant="ghost"
                      fullWidth
                      size="sm"
                      onClick={() => handleUnarchiveCategory(category.id)}
                      disabled={updating || unarchivingId === category.id}
                    >
                      {unarchivingId === category.id ? (
                        <Loader2 className="w-4 h-4 animate-spin" />
                      ) : (
                        <>
                          <RotateCcw className="w-4 h-4" />
                          Aktifkan
                        </>
                      )}
                    </Button>
                  ) : (
                    <Button
                      variant="ghost"
                      fullWidth
                      size="sm"
                      onClick={() => handleArchiveCategory(category.id)}
                      disabled={updating || archivingId === category.id}
                    >
                      {archivingId === category.id ? (
                        <Loader2 className="w-4 h-4 animate-spin" />
                      ) : (
                        <>
                          <Trash2 className="w-4 h-4" />
                          Arsipkan
                        </>
                      )}
                    </Button>
                  )}
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
            <h2 className="text-xl sm:text-2xl font-bold text-base-content mb-6">Tambah Kategori</h2>

            <form onSubmit={handleCreateCategory} className="flex flex-col gap-4">
              <FormField label="Nama Kategori">
                <Input
                  type="text"
                  placeholder="Contoh: Gaji, Makan & Minum"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  disabled={updating}
                />
              </FormField>

              <FormField label="Tipe">
                <Select
                  value={formData.type}
                  onChange={(e) =>
                    setFormData({ ...formData, type: e.target.value as 'income' | 'expense' })
                  }
                  disabled={updating}
                >
                  <option value="income">Pemasukan</option>
                  <option value="expense">Pengeluaran</option>
                </Select>
              </FormField>

              <FormField label="Warna (Opsional)">
                <div className="flex gap-2">
                  <Input
                    type="color"
                    value={formData.color}
                    onChange={(e) => setFormData({ ...formData, color: e.target.value })}
                    disabled={updating}
                    className="flex-shrink-0 w-16"
                  />
                  <div
                    className="flex-1 rounded-2xl h-12"
                    style={{ backgroundColor: formData.color }}
                  />
                </div>
              </FormField>

              <div className="flex gap-3 pt-4">
                <Button
                  type="button"
                  variant="ghost"
                  fullWidth
                  onClick={() => {
                    setShowModal(false);
                    setFormData({ name: '', type: 'income', icon: '', color: '#4CAF50' });
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

      </AppShell>
  );
}
