'use client';

import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Home, Wallet, PieChart, Tag, Target, Receipt, BarChart3, Users, Settings, LogOut, Plus } from 'lucide-react';

const NAV_ITEMS = [
  { href: '/dashboard', label: 'Beranda', icon: Home },
  { href: '/transactions', label: 'Transaksi', icon: Receipt },
  { href: '/budgets', label: 'Anggaran', icon: PieChart },
  { href: '/accounts', label: 'Akun', icon: Wallet },
] as const;

const HUB_ITEMS = [
  { href: '/categories', label: 'Kategori', icon: Tag },
  { href: '/goals', label: 'Goals', icon: Target },
  { href: '/bills', label: 'Tagihan', icon: Receipt },
  { href: '/reports', label: 'Laporan', icon: BarChart3 },
  { href: '/settings/household', label: 'Anggota', icon: Users },
  { href: '/settings/profile', label: 'Profil Saya', icon: Settings },
] as const;

interface SidebarProps {
  active: string;
}

export function Sidebar({ active }: SidebarProps) {
  const router = useRouter();

  const handleLogout = () => {
    localStorage.removeItem('token');
    router.push('/login');
  };

  return (
    <aside className="hidden lg:flex fixed inset-y-0 left-0 z-40 w-64 flex-col border-r border-base-300 bg-base-200">
      <div className="px-5 py-5 border-b border-base-300">
        <h1 className="text-lg font-bold text-base-content">Shared Finance</h1>
        <p className="text-xs text-base-content/60 mt-0.5">Kelola keuangan bersama</p>
      </div>

      <nav className="flex-1 overflow-y-auto p-3 flex flex-col gap-1">
        {NAV_ITEMS.map((item) => {
          const Icon = item.icon;
          const isActive = item.href === active;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 px-4 py-2.5 rounded-2xl transition-colors ${
                isActive ? 'bg-primary/10 text-primary' : 'text-base-content hover:bg-base-300'
              }`}
            >
              <Icon className="w-5 h-5" />
              <span className="text-sm font-medium">{item.label}</span>
            </Link>
          );
        })}

        <div className="my-2 h-px bg-base-300" />

        {HUB_ITEMS.map((item) => {
          const Icon = item.icon;
          const isActive = active === item.href || active.startsWith(item.href + '/');
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 px-4 py-2.5 rounded-2xl transition-colors ${
                isActive ? 'bg-primary/10 text-primary' : 'text-base-content hover:bg-base-300'
              }`}
            >
              <Icon className="w-5 h-5" />
              <span className="text-sm font-medium">{item.label}</span>
            </Link>
          );
        })}
      </nav>

      <div className="p-3 border-t border-base-300 flex flex-col gap-1">
        <Link
          href="/transactions/new"
          className="hidden lg:flex items-center justify-center gap-2 px-4 py-2.5 rounded-2xl bg-primary text-primary-content font-semibold text-sm hover:opacity-90 transition-opacity"
        >
          <Plus className="w-4 h-4" />
          Tambah Transaksi
        </Link>
        <button
          onClick={handleLogout}
          className="flex items-center gap-3 px-4 py-2.5 rounded-2xl text-base-content/70 hover:bg-base-300 hover:text-base-content transition-colors"
        >
          <LogOut className="w-5 h-5" />
          <span className="text-sm font-medium">Keluar</span>
        </button>
      </div>
    </aside>
  );
}
