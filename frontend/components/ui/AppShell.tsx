'use client';

import { ReactNode } from 'react';
import { Sidebar } from '@/components/ui/Sidebar';
import { BottomNav } from '@/components/ui/BottomNav';

interface AppShellProps {
  active: string;
  children: ReactNode;
}

export function AppShell({ active, children }: AppShellProps) {
  return (
    <div className="min-h-screen bg-base-100">
      <Sidebar active={active} />
      <div className="lg:pl-64">
        <div className="pb-24 lg:pb-10">{children}</div>
      </div>
      <BottomNav active={active} />
    </div>
  );
}
