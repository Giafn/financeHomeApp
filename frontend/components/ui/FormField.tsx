import { ReactNode } from 'react';

interface FormFieldProps {
  label: string;
  hint?: string;
  children: ReactNode;
}

export function FormField({ label, hint, children }: FormFieldProps) {
  return (
    <div className="flex flex-col gap-2">
      <label className="text-sm font-medium text-base-content/80">{label}</label>
      {children}
      {hint && <span className="text-xs text-base-content/50">{hint}</span>}
    </div>
  );
}
