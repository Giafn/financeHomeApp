import { AlertCircle, CheckCircle } from 'lucide-react';

interface AlertProps {
  type: 'error' | 'success';
  message: string;
}

export function Alert({ type, message }: AlertProps) {
  const isError = type === 'error';
  return (
    <div
      role="alert"
      className={`flex items-start gap-3 rounded-2xl px-4 py-3 mb-6 ${
        isError ? 'bg-error/10 text-error' : 'bg-success/10 text-success'
      }`}
    >
      {isError ? (
        <AlertCircle className="w-5 h-5 flex-shrink-0" />
      ) : (
        <CheckCircle className="w-5 h-5 flex-shrink-0" />
      )}
      <span className="text-sm">{message}</span>
    </div>
  );
}
