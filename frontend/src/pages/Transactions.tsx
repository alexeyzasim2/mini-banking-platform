import { useState, useEffect, useCallback } from 'react';
import { transactionsApi } from '../services/api';
import { TransactionList } from '../components/TransactionList';
import type { Transaction } from '../types';

export const Transactions = () => {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<string>('');
  const [page, setPage] = useState(1);
  const limit = 10;

  const loadTransactions = useCallback(async () => {
    setLoading(true);
    try {
      const params: any = { page, limit };
      if (filter) {
        params.type = filter;
      }
      const res = await transactionsApi.getTransactions(params);
      setTransactions(res.data.transactions);
    } catch (error) {
      console.error('Failed to load transactions', error);
    } finally {
      setLoading(false);
    }
  }, [page, filter]);

  useEffect(() => {
    loadTransactions();
  }, [loadTransactions]);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Transaction History</h1>
      </div>

      <div className="bg-white rounded-lg shadow p-6">
        <div className="mb-4 flex justify-between items-center">
          <div>
            <label className="text-sm font-medium text-gray-700 mr-2">
              Filter by type:
            </label>
            <select
              value={filter}
              onChange={(e) => {
                setFilter(e.target.value);
                setPage(1);
              }}
              className="rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 px-3 py-2 border"
            >
              <option value="">All</option>
              <option value="transfer">Transfer</option>
              <option value="exchange">Exchange</option>
            </select>
          </div>
        </div>

        {loading ? (
          <div className="flex justify-center items-center h-32">
            <div className="text-gray-500">Loading...</div>
          </div>
        ) : (
          <>
            <TransactionList transactions={transactions} />
            
            <div className="mt-4 flex justify-between items-center">
              <button
                onClick={() => setPage(Math.max(1, page - 1))}
                disabled={page === 1}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400"
              >
                Previous
              </button>
              <span className="text-sm text-gray-600">Page {page}</span>
              <button
                onClick={() => setPage(page + 1)}
                disabled={transactions.length < limit}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400"
              >
                Next
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

