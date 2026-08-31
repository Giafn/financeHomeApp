'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { apiCall, downloadFile, ApiError } from '@/lib/api';
import { Loader2, FileDown, FileSpreadsheet, TrendingUp, TrendingDown } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input, Select } from '@/components/ui/Input';
import { Alert } from '@/components/ui/Alert';
import { AppShell } from '@/components/ui/AppShell';

interface TrendPoint {
  month: string;
  income: number;
  expense: number;
}

interface CategoryBreakdown {
  category_id: string;
  category_name: string;
  total: number;
  percentage: number;
}

interface MemberBreakdown {
  user_id: string;
  name: string;
  total_expense: number;
  total_income: number;
}

interface ComparisonData {
  current: { period: string; total_expense: number };
  previous: { period: string; total_expense: number };
  diff_amount: number;
  diff_percentage: number;
  by_category: { category_name: string; current: number; previous: number; diff_percentage: number }[];
}

type Tab = 'trend' | 'category' | 'member' | 'comparison';

function formatCurrency(value: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(value);
}

function monthLabel(month: string) {
  const [y, m] = month.split('-').map(Number);
  return new Date(y, m - 1, 1).toLocaleDateString('id-ID', { month: 'short' });
}

function currentMonth() {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
}

const PIE_COLORS = ['#f3df1f', '#8bc34a', '#ff5252', '#4fc3f7', '#ba68c8', '#ffb74d', '#81c784', '#e57373'];

export default function ReportsPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [exporting, setExporting] = useState(false);
  const [error, setError] = useState('');

  const [tab, setTab] = useState<Tab>('trend');
  const [periodType, setPeriodType] = useState<'month' | 'year'>('month');
  const [period, setPeriod] = useState(currentMonth());

  const [trend, setTrend] = useState<TrendPoint[]>([]);
  const [categoryBreakdown, setCategoryBreakdown] = useState<CategoryBreakdown[]>([]);
  const [memberBreakdown, setMemberBreakdown] = useState<MemberBreakdown[]>([]);
  const [comparison, setComparison] = useState<ComparisonData | null>(null);

  const yearOnly = period.split('-')[0];

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const periodParam = periodType === 'year' ? yearOnly : period;
      const [trendRes, categoryRes, memberRes, comparisonRes] = await Promise.all([
        apiCall<TrendPoint[]>('/reports/trend?months=6'),
        apiCall<CategoryBreakdown[]>(`/reports/category-breakdown?period=${periodParam}&period_type=${periodType}`),
        apiCall<MemberBreakdown[]>(`/reports/member-breakdown?period=${periodParam}&period_type=${periodType}`),
        apiCall<ComparisonData>(`/reports/comparison?period=${periodParam}&period_type=${periodType}`),
      ]);
      setTrend(trendRes || []);
      setCategoryBreakdown(categoryRes || []);
      setMemberBreakdown(memberRes || []);
      setComparison(comparisonRes);
    } catch (err) {
      if (err instanceof ApiError && err.statusCode === 401) {
        router.push('/login');
      } else {
        setError('Gagal memuat laporan');
      }
    } finally {
      setLoading(false);
    }
  }, [period, periodType, yearOnly, router]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleExport = async (format: 'pdf' | 'excel') => {
    setError('');
    setExporting(true);
    try {
      const periodParam = periodType === 'year' ? yearOnly : period;
      await downloadFile(`/reports/export?format=${format}&period=${periodParam}&period_type=${periodType}`);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Gagal export laporan');
      }
    } finally {
      setExporting(false);
    }
  };

  const maxTrendValue = Math.max(1, ...trend.flatMap((t) => [t.income, t.expense]));

  return (
    <AppShell active="/reports">
      <div className="border-b border-base-300 bg-base-200 sticky top-0 z-50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
          <h1 className="text-2xl sm:text-3xl font-bold text-base-content">Laporan</h1>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" onClick={() => handleExport('pdf')} disabled={exporting}>
              {exporting ? <Loader2 className="w-4 h-4 animate-spin" /> : <FileDown className="w-4 h-4" />}
              <span className="hidden sm:inline">PDF</span>
            </Button>
            <Button variant="ghost" size="sm" onClick={() => handleExport('excel')} disabled={exporting}>
              {exporting ? <Loader2 className="w-4 h-4 animate-spin" /> : <FileSpreadsheet className="w-4 h-4" />}
              <span className="hidden sm:inline">Excel</span>
            </Button>
          </div>
        </div>
      </div>

      <div className="max-w-4xl mx-auto p-4 sm:p-6">
        {error && <Alert type="error" message={error} />}

        {/* Period selector */}
        <div className="flex items-center gap-3 mb-6">
          <Select
            value={periodType}
            onChange={(e) => setPeriodType(e.target.value as 'month' | 'year')}
            className="w-32"
          >
            <option value="month">Bulan</option>
            <option value="year">Tahun</option>
          </Select>
          {periodType === 'month' ? (
            <Input type="month" value={period} onChange={(e) => setPeriod(e.target.value)} className="flex-1" />
          ) : (
            <Input
              type="number"
              value={yearOnly}
              onChange={(e) => setPeriod(`${e.target.value}-01`)}
              min="2000"
              max="2100"
              className="flex-1"
            />
          )}
        </div>

        {/* Tabs */}
        <div className="flex gap-2 mb-6 overflow-x-auto">
          {([
            ['trend', 'Tren'],
            ['category', 'Kategori'],
            ['member', 'Anggota'],
            ['comparison', 'Perbandingan'],
          ] as [Tab, string][]).map(([key, label]) => (
            <button
              key={key}
              onClick={() => setTab(key)}
              className={`px-4 py-2 rounded-2xl text-sm font-medium whitespace-nowrap transition-colors ${
                tab === key ? 'bg-primary text-primary-content' : 'bg-base-200 text-base-content hover:bg-base-300'
              }`}
            >
              {label}
            </button>
          ))}
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="w-8 h-8 animate-spin text-primary" />
          </div>
        ) : (
          <>
            {tab === 'trend' && (
              <Card>
                <h3 className="text-lg font-bold text-base-content mb-4">Tren 6 Bulan Terakhir</h3>
                {trend.length === 0 ? (
                  <p className="text-sm text-base-content/60">Belum ada data</p>
                ) : (
                  <>
                    <div className="flex items-end justify-between gap-2 h-48 mb-2">
                      {trend.map((t) => (
                        <div key={t.month} className="flex-1 flex flex-col items-center gap-1 h-full justify-end">
                          <div className="w-full flex gap-1 items-end justify-center h-full">
                            <div
                              className="w-1/3 max-w-4 bg-success rounded-t"
                              style={{ height: `${(t.income / maxTrendValue) * 100}%` }}
                              title={formatCurrency(t.income)}
                            />
                            <div
                              className="w-1/3 max-w-4 bg-error rounded-t"
                              style={{ height: `${(t.expense / maxTrendValue) * 100}%` }}
                              title={formatCurrency(t.expense)}
                            />
                          </div>
                          <p className="text-xs text-base-content/60">{monthLabel(t.month)}</p>
                        </div>
                      ))}
                    </div>
                    <div className="flex items-center gap-4 justify-center text-xs text-base-content/60 mb-4">
                      <span className="flex items-center gap-1.5">
                        <span className="w-2.5 h-2.5 rounded-full bg-success" /> Pemasukan
                      </span>
                      <span className="flex items-center gap-1.5">
                        <span className="w-2.5 h-2.5 rounded-full bg-error" /> Pengeluaran
                      </span>
                    </div>
                    <div className="flex flex-col gap-2 pt-4 border-t border-base-300">
                      {trend.map((t) => (
                        <div key={t.month} className="flex items-center justify-between text-sm">
                          <span className="text-base-content/60">{t.month}</span>
                          <span className="text-success">{formatCurrency(t.income)}</span>
                          <span className="text-error">{formatCurrency(t.expense)}</span>
                        </div>
                      ))}
                    </div>
                  </>
                )}
              </Card>
            )}

            {tab === 'category' && (
              <Card>
                <h3 className="text-lg font-bold text-base-content mb-4">Breakdown Pengeluaran per Kategori</h3>
                {categoryBreakdown.length === 0 ? (
                  <p className="text-sm text-base-content/60">Belum ada data untuk periode ini</p>
                ) : (
                  <div className="flex flex-col gap-3">
                    {categoryBreakdown.map((c, i) => (
                      <div key={c.category_id}>
                        <div className="flex items-center justify-between text-sm mb-1">
                          <span className="flex items-center gap-2 text-base-content">
                            <span
                              className="w-2.5 h-2.5 rounded-full flex-shrink-0"
                              style={{ backgroundColor: PIE_COLORS[i % PIE_COLORS.length] }}
                            />
                            {c.category_name}
                          </span>
                          <span className="font-medium text-base-content">
                            {formatCurrency(c.total)} ({c.percentage.toFixed(0)}%)
                          </span>
                        </div>
                        <div className="w-full h-2 rounded-full bg-base-300 overflow-hidden">
                          <div
                            className="h-full rounded-full transition-all"
                            style={{ width: `${c.percentage}%`, backgroundColor: PIE_COLORS[i % PIE_COLORS.length] }}
                          />
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </Card>
            )}

            {tab === 'member' && (
              <Card>
                <h3 className="text-lg font-bold text-base-content mb-4">Breakdown per Anggota</h3>
                {memberBreakdown.length === 0 ? (
                  <p className="text-sm text-base-content/60">Belum ada data</p>
                ) : (
                  <div className="flex flex-col gap-4">
                    {memberBreakdown.map((m) => (
                      <div key={m.user_id} className="px-4 py-3 bg-base-100 border border-base-300 rounded-2xl">
                        <p className="font-medium text-base-content mb-2">{m.name}</p>
                        <div className="flex items-center justify-between text-sm">
                          <span className="text-base-content/60">Pengeluaran</span>
                          <span className="text-error font-medium">{formatCurrency(m.total_expense)}</span>
                        </div>
                        <div className="flex items-center justify-between text-sm">
                          <span className="text-base-content/60">Pemasukan</span>
                          <span className="text-success font-medium">{formatCurrency(m.total_income)}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </Card>
            )}

            {tab === 'comparison' && comparison && (
              <div className="flex flex-col gap-4">
                <Card>
                  <div className="flex items-center justify-between mb-4">
                    <div>
                      <p className="text-xs text-base-content/60">
                        {comparison.previous.period} → {comparison.current.period}
                      </p>
                      <p className="text-2xl font-bold text-base-content">{formatCurrency(comparison.current.total_expense)}</p>
                    </div>
                    <div
                      className={`flex items-center gap-1 px-3 py-1.5 rounded-full text-sm font-semibold ${
                        comparison.diff_amount > 0 ? 'bg-error/20 text-error' : 'bg-success/20 text-success'
                      }`}
                    >
                      {comparison.diff_amount > 0 ? (
                        <TrendingUp className="w-4 h-4" />
                      ) : (
                        <TrendingDown className="w-4 h-4" />
                      )}
                      {Math.abs(comparison.diff_percentage).toFixed(1)}%
                    </div>
                  </div>
                  <p className="text-xs text-base-content/60">
                    Periode sebelumnya ({comparison.previous.period}): {formatCurrency(comparison.previous.total_expense)}
                  </p>
                </Card>

                <Card>
                  <h3 className="text-lg font-bold text-base-content mb-4">Perbandingan per Kategori</h3>
                  {comparison.by_category.length === 0 ? (
                    <p className="text-sm text-base-content/60">Belum ada data</p>
                  ) : (
                    <div className="flex flex-col gap-3">
                      {comparison.by_category.map((c) => (
                        <div key={c.category_name} className="flex items-center justify-between text-sm">
                          <span className="text-base-content">{c.category_name}</span>
                          <div className="text-right">
                            <p className="text-base-content font-medium">{formatCurrency(c.current)}</p>
                            <p
                              className={`text-xs ${
                                c.diff_percentage > 0 ? 'text-error' : c.diff_percentage < 0 ? 'text-success' : 'text-base-content/60'
                              }`}
                            >
                              {c.diff_percentage > 0 ? '+' : ''}
                              {c.diff_percentage.toFixed(1)}% dari {formatCurrency(c.previous)}
                            </p>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </Card>
              </div>
            )}
          </>
        )}
      </div>

      </AppShell>
  );
}
