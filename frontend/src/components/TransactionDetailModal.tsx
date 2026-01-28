import type { Transaction } from '../types';
import { centsToDollars } from '../types';

interface Props {
  isOpen: boolean;
  transaction: Transaction | null;
  onClose: () => void;
}

export const TransactionDetailModal = ({ isOpen, transaction, onClose }: Props) => {
  if (!isOpen || !transaction) return null;

  const formatDate = (date: string) => {
    return new Date(date).toLocaleString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  const isExchange = transaction.type === 'exchange';
  const exchangeRate = isExchange && transaction.description.includes('USD')
    ? '1 USD = 0.92 EUR'
    : '1 EUR = 1.087 USD';

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
        <div className="p-6">
          <div className="flex justify-between items-start mb-6">
            <div>
              <h2 className="text-2xl font-bold text-gray-900">Transaction Receipt</h2>
              <p className="text-sm text-gray-500 mt-1">ID: {transaction.id.substring(0, 8)}...</p>
            </div>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 text-2xl leading-none"
            >
              ×
            </button>
          </div>

          <div className="space-y-4">
            <div className="flex justify-between items-center py-3 border-b">
              <span className="text-sm font-medium text-gray-500">Type</span>
              <span className={`px-3 py-1 text-sm font-semibold rounded-full ${
                transaction.type === 'transfer' 
                  ? 'bg-blue-100 text-blue-800' 
                  : 'bg-green-100 text-green-800'
              }`}>
                {transaction.type.charAt(0).toUpperCase() + transaction.type.slice(1)}
              </span>
            </div>

            <div className="flex justify-between items-center py-3 border-b">
              <span className="text-sm font-medium text-gray-500">Amount</span>
              <span className="text-lg font-bold text-gray-900">
                {centsToDollars(transaction.amount_cents)} {transaction.currency}
              </span>
            </div>

            <div className="flex justify-between items-center py-3 border-b">
              <span className="text-sm font-medium text-gray-500">Date & Time</span>
              <span className="text-sm text-gray-900">{formatDate(transaction.created_at)}</span>
            </div>

            {isExchange && (
              <div className="flex justify-between items-center py-3 border-b">
                <span className="text-sm font-medium text-gray-500">Exchange Rate</span>
                <span className="text-sm text-gray-900">{exchangeRate}</span>
              </div>
            )}

            <div className="py-3 border-b">
              <span className="text-sm font-medium text-gray-500 block mb-2">Description</span>
              <p className="text-sm text-gray-900">{transaction.description}</p>
            </div>

            <div className="flex justify-between items-center py-3 border-b">
              <span className="text-sm font-medium text-gray-500">Status</span>
              <span className="px-3 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">
                Completed
              </span>
            </div>

            <div className="flex justify-between items-center py-3">
              <span className="text-sm font-medium text-gray-500">Transaction ID</span>
              <span className="text-xs text-gray-600 font-mono">{transaction.id}</span>
            </div>
          </div>

          <div className="mt-6 pt-6 border-t">
            <button
              onClick={onClose}
              className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              Close
            </button>
          </div>

          <div className="mt-4 p-3 bg-gray-50 rounded-md">
            <p className="text-xs text-gray-500 text-center">
              This receipt is for your records. Keep it for future reference.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

