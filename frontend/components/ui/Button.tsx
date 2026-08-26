import { ButtonHTMLAttributes, ReactNode } from 'react';

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'outline';
type ButtonSize = 'sm' | 'md';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  fullWidth?: boolean;
  children: ReactNode;
}

const VARIANT_CLASSES: Record<ButtonVariant, string> = {
  primary: 'bg-primary text-primary-content hover:brightness-95',
  secondary: 'bg-secondary text-secondary-content hover:brightness-95',
  ghost: 'bg-transparent text-base-content hover:bg-base-200',
  outline: 'bg-transparent text-base-content border border-base-300 hover:bg-base-200',
};

const SIZE_CLASSES: Record<ButtonSize, string> = {
  sm: 'px-4 h-9 text-xs',
  md: 'px-6 h-12 text-sm',
};

export function Button({
  variant = 'primary',
  size = 'md',
  fullWidth = false,
  className = '',
  children,
  disabled,
  ...props
}: ButtonProps) {
  return (
    <button
      disabled={disabled}
      className={`inline-flex items-center justify-center gap-2 rounded-full font-semibold transition-all active:scale-[0.97] disabled:opacity-40 disabled:cursor-not-allowed disabled:active:scale-100 ${SIZE_CLASSES[size]} ${VARIANT_CLASSES[variant]} ${fullWidth ? 'w-full' : ''} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}
