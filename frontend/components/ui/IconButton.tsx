import { ButtonHTMLAttributes, ReactNode } from 'react';

interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
}

export function IconButton({ className = '', children, ...props }: IconButtonProps) {
  return (
    <button
      className={`inline-flex items-center justify-center w-10 h-10 rounded-full bg-base-200 border border-base-300 text-base-content hover:bg-base-300 transition-colors flex-shrink-0 ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}
