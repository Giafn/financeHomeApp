import Link from 'next/link';
import { ChevronLeft } from 'lucide-react';
import { ReactNode } from 'react';

interface TopBarProps {
  title: string;
  subtitle?: string;
  backHref?: string;
  action?: ReactNode;
}

export function TopBar({ title, subtitle, backHref, action }: TopBarProps) {
  return (
    <div className="sticky top-0 z-40 bg-base-100/90 backdrop-blur border-b border-base-300">
      <div className="max-w-2xl mx-auto px-4 sm:px-6 py-4 flex items-center gap-4">
        {backHref && (
          <Link
            href={backHref}
            className="inline-flex items-center justify-center w-10 h-10 rounded-full bg-base-200 border border-base-300 text-base-content hover:bg-base-300 transition-colors flex-shrink-0"
          >
            <ChevronLeft className="w-5 h-5" />
          </Link>
        )}
        <div className="flex-1 min-w-0">
          <h1 className="text-xl sm:text-2xl font-bold text-base-content truncate">{title}</h1>
          {subtitle && <p className="text-xs sm:text-sm text-base-content/60 truncate">{subtitle}</p>}
        </div>
        {action}
      </div>
    </div>
  );
}
