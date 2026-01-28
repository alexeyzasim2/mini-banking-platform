import { useState } from 'react';
import { transactionsApi } from '../services/api';
import { ConfirmationModal } from './ConfirmationModal';
import type { Account } from '../types';
import { dollarsToCents } from '../types';

interface Props {
  accounts: Account[];
  onSuccess: () => void;
}

export const TransferForm = ({ accounts, onSuccess }: Props) => {
  const [recipientId, setRecipientId] = useState('');
  const [currency, setCurrency] = useState('USD');
  const [amount, setAmount] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showConfirm, setShowConfirm] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setShowConfirm(true);
  };

  const handleConfirm = async () => {
    setShowConfirm(false);
    setLoading(true);

    try {
      await transactionsApi.transfer({
        to_user_id: recipientId,
        currency,
        amount_cents: dollarsToCents(amount),
      });
      setRecipientId('');
      setAmount('');
      onSuccess();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Transfer failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <ConfirmationModal
        isOpen={showConfirm}
        title="Confirm Transfer"
        message={
          <div>
            <p>Transfer <strong>{parseFloat(amount || '0').toFixed(2)} {currency}</strong> to:</p>
            <p className="mt-2 text-sm text-gray-600">User ID: {recipientId}</p>
          </div>
        }
        onConfirm={handleConfirm}
        onCancel={() => setShowConfirm(false)}
        confirmText="Transfer"
        confirmColor="bg-blue-600 hover:bg-blue-700"
      />

      <form onSubmit={handleSubmit} className="flex flex-col h-full space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700">
          Recipient Email
        </label>
        <input
          type="email"
          value={recipientId}
          onChange={(e) => setRecipientId(e.target.value)}
          required
          placeholder="alice@example.com"
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 px-3 py-2 border"
        />
        <p className="mt-1 text-xs text-gray-500">
          Available users: alice@example.com, bob@example.com, charlie@example.com
        </p>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700">
          Currency
        </label>
        <select
          value={currency}
          onChange={(e) => setCurrency(e.target.value)}
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 px-3 py-2 border"
        >
          <option value="USD">USD</option>
          <option value="EUR">EUR</option>
        </select>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700">
          Amount
        </label>
        <input
          type="number"
          step="0.01"
          min="0.01"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          required
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 px-3 py-2 border"
        />
      </div>

      {error && (
        <div className="text-red-600 text-sm">{error}</div>
      )}

      <div className="mt-auto">
        <button
          type="submit"
          disabled={loading}
          className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400"
        >
          {loading ? 'Processing...' : 'Transfer'}
        </button>
      </div>
    </form>
    </>
  );
};

