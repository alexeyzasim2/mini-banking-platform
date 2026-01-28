import { useState } from 'react';
import { transactionsApi } from '../services/api';
import { ConfirmationModal } from './ConfirmationModal';
import { dollarsToCents } from '../types';

interface Props {
  onSuccess: () => void;
}

const EXCHANGE_RATE = 0.92;

export const ExchangeForm = ({ onSuccess }: Props) => {
  const [fromCurrency, setFromCurrency] = useState('USD');
  const [toCurrency, setToCurrency] = useState('EUR');
  const [amount, setAmount] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showConfirm, setShowConfirm] = useState(false);

  const calculateConverted = () => {
    if (!amount) return 0;
    const num = parseFloat(amount);
    return fromCurrency === 'USD' ? (num * EXCHANGE_RATE).toFixed(2) : (num / EXCHANGE_RATE).toFixed(2);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setShowConfirm(true);
  };

  const handleConfirm = async () => {
    setShowConfirm(false);
    setLoading(true);

    try {
      await transactionsApi.exchange({
        from_currency: fromCurrency,
        amount_cents: dollarsToCents(amount),
      });
      setAmount('');
      onSuccess();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Exchange failed');
    } finally {
      setLoading(false);
    }
  };

  const swapCurrencies = () => {
    setFromCurrency(toCurrency);
    setToCurrency(fromCurrency);
  };

  return (
    <>
      <ConfirmationModal
        isOpen={showConfirm}
        title="Confirm Exchange"
        message={
          <div>
            <p>Exchange <strong>{parseFloat(amount || '0').toFixed(2)} {fromCurrency}</strong></p>
            <p className="mt-2">You will receive: <strong>{calculateConverted()} {toCurrency}</strong></p>
            <p className="mt-2 text-sm text-gray-600">Rate: 1 USD = {EXCHANGE_RATE} EUR</p>
          </div>
        }
        onConfirm={handleConfirm}
        onCancel={() => setShowConfirm(false)}
        confirmText="Exchange"
        confirmColor="bg-green-600 hover:bg-green-700"
      />

      <form onSubmit={handleSubmit} className="flex flex-col h-full space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700">
          From Currency
        </label>
        <div className="flex gap-2 mt-1">
          <select
            value={fromCurrency}
            onChange={(e) => {
              setFromCurrency(e.target.value);
              setToCurrency(e.target.value === 'USD' ? 'EUR' : 'USD');
            }}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 px-3 py-2 border"
          >
            <option value="USD">USD</option>
            <option value="EUR">EUR</option>
          </select>
          <button
            type="button"
            onClick={swapCurrencies}
            className="px-3 py-2 bg-gray-200 rounded-md hover:bg-gray-300"
          >
            ⇄
          </button>
        </div>
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

      <div className="bg-gray-50 p-4 rounded-md">
        <div className="text-sm text-gray-600">
          Exchange Rate: 1 USD = {EXCHANGE_RATE} EUR
        </div>
        <div className="mt-2 text-lg font-semibold">
          You will receive: {calculateConverted()} {toCurrency}
        </div>
      </div>

      {error && (
        <div className="text-red-600 text-sm">{error}</div>
      )}

      <div className="mt-auto">
        <button
          type="submit"
          disabled={loading}
          className="w-full px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:bg-gray-400"
        >
          {loading ? 'Processing...' : 'Exchange'}
        </button>
      </div>
    </form>
    </>
  );
};

