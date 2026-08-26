import { InputHTMLAttributes, forwardRef } from 'react';

interface CurrencyInputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'onChange' | 'value' | 'type'> {
  value: string;
  onChange: (rawValue: string) => void;
}

function formatThousands(digits: string) {
  if (!digits) return '';
  return digits.replace(/\B(?=(\d{3})+(?!\d))/g, '.');
}

// Input nominal uang — tampil dengan separator ribuan ("1.000.000") + prefix "Rp", tapi
// value/onChange tetap string digit murni ("1000000") supaya parseFloat() di caller tidak berubah.
export const CurrencyInput = forwardRef<HTMLInputElement, CurrencyInputProps>(
  ({ value, onChange, className = '', ...props }, ref) => {
    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      onChange(e.target.value.replace(/\D/g, ''));
    };

    return (
      <div className="relative">
        <span className="absolute left-4 top-1/2 -translate-y-1/2 text-base-content/40 text-sm pointer-events-none">
          Rp
        </span>
        <input
          ref={ref}
          type="text"
          inputMode="numeric"
          value={formatThousands(value)}
          onChange={handleChange}
          className={`w-full h-12 pl-10 pr-4 rounded-2xl bg-base-100 border border-base-300 text-base-content placeholder:text-base-content/40 focus:outline-none focus:border-primary transition-colors ${className}`}
          {...props}
        />
      </div>
    );
  }
);
CurrencyInput.displayName = 'CurrencyInput';
