'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiCall } from '@/lib/api';
import { Menu, X, Settings, LogOut, Loader2, DollarSign, BarChart3, TrendingUp, Users } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { BottomNav } from '@/components/ui/BottomNav';

interface UserProfile {
  id: string;
  name: string;
  email: string;
  household: { id: string; name: string; role: string } | null;
}

export default function DashboardPage() {
  const router = useRouter();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [menuOpen, setMenuOpen] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/login');
      return;
    }

    const loadProfile = async () => {
      try {
        const data = await apiCall<UserProfile>('/users/me');
        setProfile(data);
      } catch {
        // Handle error silently for now
      } finally {
        setLoading(false);
      }
    };

    loadProfile();
  }, [router]);

  const handleLogout = () => {
    localStorage.removeItem('token');
    router.push('/login');
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

  const statCards = [
    { label: 'Total Saldo', value: '-', Icon: DollarSign },
    { label: 'Pengeluaran Bulan Ini', value: '-', Icon: BarChart3 },
    { label: 'Total Budget', value: '-', Icon: TrendingUp },
    { label: 'Anggota Keluarga', value: '1', Icon: Users },
  ];

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

          {/* Desktop menu */}
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

          {/* Mobile menu button */}
          <div className="sm:hidden">
            <button
              className="inline-flex items-center justify-center w-10 h-10 rounded-full hover:bg-base-300 transition-colors"
              onClick={() => setMenuOpen(!menuOpen)}
            >
              {menuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </button>
          </div>
        </div>

        {/* Mobile dropdown menu */}
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

      {/* Main content */}
      <div className="max-w-6xl mx-auto px-4 sm:px-6 py-6 sm:py-8 pb-24 sm:pb-8">
        <div className="mb-6 sm:mb-8">
          <h2 className="text-2xl sm:text-3xl font-bold text-base-content mb-2">
            Selamat datang, {profile?.name}!
          </h2>
          <p className="text-sm sm:text-base text-base-content/60">
            Dashboard ini akan menampilkan ringkasan keuangan keluarga Anda.
          </p>
        </div>

        {/* Stat cards */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4 mb-6 sm:mb-8">
          {statCards.map((item, i) => (
            <Card key={i} className="p-4 sm:p-6">
              <div className="flex items-start justify-between gap-2">
                <div className="flex-1 min-w-0">
                  <p className="text-xs sm:text-sm text-base-content/60 mb-1 truncate">{item.label}</p>
                  <p className="text-lg sm:text-2xl font-bold text-base-content truncate">{item.value}</p>
                </div>
                <item.Icon className="w-6 h-6 sm:w-8 sm:h-8 text-primary flex-shrink-0" />
              </div>
            </Card>
          ))}
        </div>

        {/* Info card */}
        <Card>
          <h3 className="text-lg sm:text-xl font-bold text-base-content mb-2 sm:mb-3">Fitur akan datang</h3>
          <p className="text-sm sm:text-base text-base-content/60 mb-2 sm:mb-4">
            Dashboard ini merupakan placeholder. Fitur lengkap seperti manajemen akun, pencatatan transaksi,
            budget tracking, dan laporan akan ditambahkan di phase berikutnya.
          </p>
          <p className="text-xs sm:text-sm text-base-content/60">
            Saat ini Anda dapat mengakses pengaturan profil melalui menu di atas.
          </p>
        </Card>
      </div>

      <BottomNav active="/dashboard" />
    </div>
  );
}
