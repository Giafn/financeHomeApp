import { InputHTMLAttributes, SelectHTMLAttributes, forwardRef } from 'react';

export const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  ({ className = '', ...props }, ref) => (
    <input
      ref={ref}
      className={`w-full h-12 px-4 rounded-2xl bg-base-100 border border-base-300 text-base-content placeholder:text-base-content/40 focus:outline-none focus:border-primary transition-colors ${className}`}
      {...props}
    />
  )
);
Input.displayName = 'Input';

export const Select = forwardRef<HTMLSelectElement, SelectHTMLAttributes<HTMLSelectElement>>(
  ({ className = '', children, ...props }, ref) => (
    <select
      ref={ref}
      className={`w-full h-12 px-4 rounded-2xl bg-base-100 border border-base-300 text-base-content focus:outline-none focus:border-primary transition-colors ${className}`}
      {...props}
    >
      {children}
    </select>
  )
);
Select.displayName = 'Select';
