'use client';

import Link from 'next/link';
import { Home, Wallet, BarChart2, Settings } from 'lucide-react';

const NAV_ITEMS = [
  { href: '/dashboard', label: 'Beranda', icon: Home },
  { href: '/accounts', label: 'Akun', icon: Wallet },
  { href: null, label: 'Laporan', icon: BarChart2 },
  { href: '/settings/profile', label: 'Profil', icon: Settings },
] as const;

interface BottomNavProps {
  active: string;
}

export function BottomNav({ active }: BottomNavProps) {
  return (
    <div className="sm:hidden fixed bottom-0 left-0 right-0 border-t border-base-300 bg-base-200 z-50">
      <div className="flex items-center justify-around h-16">
        {NAV_ITEMS.map((item) => {
          const Icon = item.icon;

          if (!item.href) {
            return (
              <button
                key={item.label}
                disabled
                className="flex flex-col items-center justify-center gap-1 flex-1 h-full text-base-content/30 cursor-not-allowed"
              >
                <Icon className="w-5 h-5" />
                <span className="text-xs font-medium">{item.label}</span>
              </button>
            );
          }

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
      </div>
    </div>
  );
}
