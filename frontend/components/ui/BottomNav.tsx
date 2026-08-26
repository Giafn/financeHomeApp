'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Home, Wallet, PieChart, Settings, Plus, Menu, X, Tag, Target, Receipt, Users, BarChart3 } from 'lucide-react';

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
  { href: '/settings/household', label: 'Rumah Tangga', icon: Users },
  { href: '/settings/profile', label: 'Profil Saya', icon: Settings },
] as const;

function isHubRoute(active: string) {
  return HUB_ITEMS.some((item) => active === item.href || active.startsWith(item.href + '/'));
}

interface BottomNavProps {
  active: string;
}

export function BottomNav({ active }: BottomNavProps) {
  const [hubOpen, setHubOpen] = useState(false);
  const hubActive = isHubRoute(active);

  return (
    <>
      <Link
        href="/transactions/new"
        className="sm:hidden fixed bottom-20 right-4 z-50 inline-flex items-center justify-center w-14 h-14 rounded-full bg-primary text-primary-content shadow-lg active:scale-95 transition-transform"
        aria-label="Tambah transaksi"
      >
        <Plus className="w-6 h-6" />
      </Link>

      {/* Hub sheet — Kategori/Goals/Tagihan/Rumah Tangga/Profil, reachable from anywhere */}
      {hubOpen && (
        <div className="fixed inset-0 z-50 flex items-end sm:items-center justify-center" onClick={() => setHubOpen(false)}>
          <div className="absolute inset-0 bg-black/60" />
          <div
            className="relative w-full sm:max-w-sm bg-base-200 border border-base-300 rounded-t-3xl sm:rounded-3xl p-4 pb-6 sm:pb-4"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between mb-2 px-2 pt-1">
              <p className="text-sm font-semibold text-base-content/60">Lainnya</p>
              <button
                onClick={() => setHubOpen(false)}
                className="inline-flex items-center justify-center w-8 h-8 rounded-full hover:bg-base-300 transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
            <div className="flex flex-col gap-1">
              {HUB_ITEMS.map((item) => {
                const Icon = item.icon;
                const isActive = active === item.href || active.startsWith(item.href + '/');
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    onClick={() => setHubOpen(false)}
                    className={`flex items-center gap-3 px-4 py-3 rounded-2xl transition-colors ${
                      isActive ? 'bg-primary/10 text-primary' : 'text-base-content hover:bg-base-300'
                    }`}
                  >
                    <Icon className="w-5 h-5" />
                    <span className="text-sm font-medium">{item.label}</span>
                  </Link>
                );
              })}
            </div>
          </div>
        </div>
      )}

      <div className="fixed bottom-0 left-0 right-0 border-t border-base-300 bg-base-200 z-40">
        <div className="max-w-md sm:max-w-2xl mx-auto flex items-center justify-around h-16">
          {NAV_ITEMS.map((item) => {
            const Icon = item.icon;
            const isActive = item.href === active;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`flex flex-col items-center justify-center gap-1 flex-1 h-full transition-colors ${
                  isActive ? 'text-primary' : 'text-base-content/60 hover:text-primary'
                }`}
              >
                <Icon className="w-5 h-5" />
                <span className="text-xs font-medium">{item.label}</span>
              </Link>
            );
          })}
          <button
            onClick={() => setHubOpen(true)}
            className={`flex flex-col items-center justify-center gap-1 flex-1 h-full transition-colors ${
              hubActive ? 'text-primary' : 'text-base-content/60 hover:text-primary'
            }`}
          >
            <Menu className="w-5 h-5" />
            <span className="text-xs font-medium">Lainnya</span>
          </button>
        </div>
      </div>
    </>
  );
}
