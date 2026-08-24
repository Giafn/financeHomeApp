import { HTMLAttributes, ReactNode } from 'react';

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

export function Card({ className = '', children, ...props }: CardProps) {
  return (
    <div
      className={`bg-base-200 border border-base-300 rounded-3xl p-6 sm:p-8 ${className}`}
      {...props}
    >
      {children}
    </div>
  );
}
